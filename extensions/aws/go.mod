module github.com/goccha/logging/extensions/aws

go 1.17

require (
	github.com/aws/aws-lambda-go v1.32.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.13.2
	github.com/goccha/envar v0.1.4
	github.com/goccha/http-constants v0.0.4
	github.com/goccha/logging v0.0.4
	github.com/rs/zerolog v1.27.0
	go.opentelemetry.io/otel/trace v1.7.0
)

require (
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/goccha/log v0.0.2 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/sys v0.0.0-20220610221304-9f5ed59c137d // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/goccha/logging => ../../.
