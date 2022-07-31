module github.com/goccha/logging/resty

go 1.17

require (
	github.com/go-resty/resty/v2 v2.7.0
	github.com/goccha/logging v0.0.8
	github.com/rs/zerolog v1.27.0
)

require (
	github.com/goccha/http-constants v0.0.4 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/net v0.0.0-20220728211354-c7608f3a8462 // indirect
	golang.org/x/sys v0.0.0-20220730100132-1609e554cd39 // indirect
)

replace github.com/goccha/logging => ../.
