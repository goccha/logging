package xray

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
	AwsTraceId = "aws_trace_id"
	SpanId     = "span_id"
	TraceId    = "trace_id"
	Sampled    = "sampled"
)

var awsEnv = envar.String("AWS_EXECUTION_ENV")
var isLambda = strings.HasPrefix(awsEnv, "AWS_Lambda_")

// New creates a new tracing context for HTTP requests.
// tracing.NewFunc is used to create a new tracing context based on the incoming HTTP request.
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

func RequestId() tracelog.RequestIdFunc {
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

func Context(ctx context.Context) *TracingContext {
	value := ctx.Value(tracing.Key())
	if value != nil {
		if tc, ok := value.(*TracingContext); ok {
			return tc
		}
	}
	return nil
}

type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
}

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

func (tc *TracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	traceId := spanCtx.TraceID().String()
	event = event.Str(TraceId, traceId).
		Str(AwsTraceId, Format(traceId)).
		Str(SpanId, spanCtx.SpanID().String()).
		Str(Sampled, cond(spanCtx.IsSampled(), "01", "00"))
	if tc.RequestID != "" {
		event = event.Str("request_id", tc.RequestID)
	}
	return event
}

// Format converts a trace ID to the format used by AWS X-Ray.
func Format(traceId string) string {
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
