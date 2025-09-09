package tracelog

import (
	"github.com/goccha/envar"
	"github.com/goccha/logging/extensions/cloudtrace"
	"github.com/goccha/logging/tracing"
	"github.com/goccha/logging/tracing/tracelog"
)

func init() {
	tracing.Setup(tracing.ServiceName(envar.String("GAE_SERVICE", "K_SERVICE")))
}

func Setup(opt ...tracelog.Option) {
	opt = append(opt, tracelog.WithNewFunc(cloudtrace.New()))
	tracelog.Setup(opt...)
}
