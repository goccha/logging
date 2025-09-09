package tracelog

import (
	"github.com/goccha/logging/extensions/xray"
	"github.com/goccha/logging/tracing/tracelog"
)

func Setup(opt ...tracelog.Option) {
	opt = append(opt, tracelog.WithNewFunc(xray.New()))
	tracelog.Setup(opt...)
}
