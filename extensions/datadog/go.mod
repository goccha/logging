module github.com/goccha/logging/extensions/datadog

go 1.21

require (
	github.com/goccha/http-constants v0.1.1
	github.com/goccha/logging v0.1.7
	github.com/rs/zerolog v1.33.0
	go.opentelemetry.io/otel/trace v1.28.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
)

replace github.com/goccha/logging => ../../.
