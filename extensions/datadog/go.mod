module github.com/goccha/logging/extensions/datadog

go 1.21.7

require (
	github.com/goccha/http-constants v0.1.0
	github.com/goccha/logging v0.1.5
	github.com/rs/zerolog v1.32.0
	go.opentelemetry.io/otel/trace v1.24.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

replace github.com/goccha/logging => ../../.
