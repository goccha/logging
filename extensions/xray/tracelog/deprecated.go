package tracelog

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/goccha/envar"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/tracing"
	"github.com/goccha/logging/tracing/tracelog"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Deprecated: Use xray.AwsTraceId instead.
	AwsTraceId = "aws_trace_id" // AWS X-Ray format trace ID
	// Deprecated: Use xray.SpanId instead.
	SpanId = "span_id" // Span ID in hexadecimal format
	// Deprecated: Use xray.TraceId instead.
	TraceId = "trace_id" // Trace ID in hexadecimal format
	// Deprecated: Use xray.Sampled instead.
	Sampled = "sampled" // Sampling status, "01" for sampled, "00" for not sampled
)

var awsEnv = envar.String("AWS_EXECUTION_ENV")
var isLambda = strings.HasPrefix(awsEnv, "AWS_Lambda_")

// New creates a new tracing context for HTTP requests.
// tracing.NewFunc is used to create a new tracing context based on the incoming HTTP request.
// Deprecated: Use logging/tracing/tracelog.New instead.
func New() func(ctx context.Context, req *http.Request) tracing.Tracing {
	return func(ctx context.Context, req *http.Request) tracing.Tracing {
		return &TracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: getRequestId(ctx, req),
			Service:   tracing.Service(),
		}
	}
}

func AwsRequestId() tracelog.RequestIdFunc {
	return getRequestId
}

func getRequestId(ctx context.Context, req *http.Request) string {
	var requestId string
	if isLambda { // Lambda環境の場合
		requestId = getLambdaRequestId(ctx)
	} else {
		config := tracelog.TraceConfig()
		if config.RequestIdHeader != "" {
			requestId = req.Header.Get(config.RequestIdHeader)
		}
	}
	if requestId == "" {
		requestId = req.Header.Get(headers.AwsRequestID)
	}
	if requestId == "" {
		requestId = req.Header.Get(headers.RequestID)
	}
	return requestId
}

// getLambdaRequestId retrieves the request ID from the Lambda context or API Gateway context.
// It checks if the context contains an API Gateway context (for REST or HTTP APIs) or
// falls back to the Lambda context if not found.
// If no request ID is found, it returns an empty string.
// Deprecated: Use tracing.GetLambdaRequestId instead.
func getLambdaRequestId(ctx context.Context) string {
	var requestId string
	if rest, ok := core.GetAPIGatewayContextFromContext(ctx); ok { // RestAPI
		requestId = rest.RequestID
	} else {
		if v2, ok := core.GetAPIGatewayV2ContextFromContext(ctx); ok { // HttpAPI
			requestId = v2.RequestID
		}
	}
	if requestId == "" {
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			requestId = lc.AwsRequestID
		}
	}
	return requestId
}

// WithTrace
// Deprecated: Use tracing.LogOption instead.
func WithTrace() tracing.LogFunc {
	return func(ctx context.Context, event *zerolog.Event) *zerolog.Event {
		value := ctx.Value(tracing.Key())
		if value != nil {
			if tc, ok := value.(tracing.Tracing); ok {
				event = tc.WithTrace(ctx, event)
			}
		}
		return event
	}
}

// Context retrieves the TracingContext from the given context.
// If no TracingContext is found, it returns nil.
// Deprecated: Use logging/tracing/tracelog.Context instead.
func Context(ctx context.Context) *TracingContext {
	value := ctx.Value(tracing.Key())
	if value != nil {
		if tc, ok := value.(*TracingContext); ok {
			return tc
		}
	}
	return nil
}

// TracingContext holds tracing information for a request.
// It includes the request path, client IP, request ID, and service name.
// Deprecated: Use logging/tracing/tracelog.TracingContext instead.
type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
}

// Dump adds tracing information to the given log event.
// It includes trace ID, span ID, sampling status, client IP, and request ID.
// Deprecated: Use tracing.LogOption instead.
func (tc *TracingContext) Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	log = log.Str("trace_id", spanCtx.TraceID().String()).Str("span_id", spanCtx.SpanID().String()).
		Str("sampled", cond(spanCtx.IsSampled(), "01", "00"))
	if tc.Service != "" {
		log = log.Dict("serviceContext", zerolog.Dict().Str("service", tc.Service))
	}
	return log.Str("client_ip", tc.ClientIP).
		Str("request_id", tc.RequestID)
}

// WithTrace adds tracing information to the given event.
// It includes trace ID, AWS trace ID, span ID, sampling status, and request ID.
// Deprecated: Use tracing.LogOption instead.
func (tc *TracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	traceId := spanCtx.TraceID().String()
	event = event.Str(TraceId, traceId).
		Str(AwsTraceId, XrayFormat(traceId)).
		Str(SpanId, spanCtx.SpanID().String()).
		Str(Sampled, cond(spanCtx.IsSampled(), "01", "00"))
	if tc.RequestID != "" {
		event = event.Str("request_id", tc.RequestID)
	}
	return event
}

// XrayFormat converts a trace ID to the format used by AWS X-Ray.
// Deprecated: Use xray.Format from xray package instead.
func XrayFormat(traceId string) string {
	if len(traceId) < 16 {
		return traceId // 16文字未満の場合はそのまま返す
	}
	return "1-" + traceId[0:8] + "-" + traceId[8:]
}

func cond(is bool, trueValue string, falseValue string) string {
	if is {
		return trueValue
	}
	return falseValue
}
