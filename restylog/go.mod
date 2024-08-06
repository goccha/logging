module github.com/goccha/logging/restylog

go 1.18

require (
	github.com/go-resty/resty/v2 v2.14.0
	github.com/goccha/http-constants v0.1.1
	github.com/goccha/logging v0.1.7
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
)

replace github.com/goccha/logging => ../.
