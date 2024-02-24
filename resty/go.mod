module github.com/goccha/logging/resty

go 1.17

require (
	github.com/go-resty/resty/v2 v2.11.0
	github.com/goccha/logging v0.1.4
	github.com/rs/zerolog v1.32.0
)

require (
	github.com/goccha/http-constants v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace github.com/goccha/logging => ../.
