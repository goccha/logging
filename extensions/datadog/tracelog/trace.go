package tracelog

import (
	"github.com/goccha/logging/extensions/datadog"
	"github.com/goccha/logging/tracing/tracelog"
)

func Setup(opt ...tracelog.Option) {
	opt = append(opt, tracelog.WithNewFunc(datadog.New()))
	tracelog.Setup(opt...)
}
