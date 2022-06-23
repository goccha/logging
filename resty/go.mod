module github.com/goccha/logging/resty

go 1.17

require (
	github.com/go-resty/resty/v2 v2.7.0
	github.com/goccha/logging v0.0.5
	github.com/rs/zerolog v1.27.0
)

require (
	github.com/goccha/http-constants v0.0.4 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/net v0.0.0-20220622184535-263ec571b305 // indirect
	golang.org/x/sys v0.0.0-20220622161953-175b2fd9d664 // indirect
)

replace github.com/goccha/logging => ../.
