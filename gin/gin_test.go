package gin

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/goccha/logging/extensions/aws/tracelog"
	"github.com/goccha/logging/log"
	"go.opentelemetry.io/otel"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJsonLogger(t *testing.T) {
	reqBody := bytes.NewBufferString("request body")
	req := httptest.NewRequest(http.MethodGet, "http://dummy.url.com/user", reqBody)
	req.Header.Add("User-Agent", "Test/0.0.1")

	buf := &bytes.Buffer{}
	log.SetGlobalOut(buf)
	tracelog.Setup()

	var tracer = otel.Tracer("github.com/goccha-tracer")
	router := gin.New()
	router.Use(TraceRequest(tracer, true, tracelog.New()), AccessLog()).
		GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	w := PerformRequest(router, "GET", "/test")
	//fmt.Printf("%s\n", buf.String())
	for _, str := range strings.Split(buf.String(), "\n") {
		if len(str) == 0 {
			continue
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal([]byte(str), &m); err != nil {
			t.Error(err)
			return
		}
		if v, ok := m[tracelog.Id]; ok {
			assert.Equal(t, "00000000000000000000000000000000", v)
		}
		if v, ok := m[tracelog.SpanId]; ok {
			assert.Equal(t, "0000000000000000", v)
		}
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

type header struct {
	Key   string
	Value string
}

func PerformRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
