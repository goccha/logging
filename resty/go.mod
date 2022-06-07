module github.com/goccha/logging/resty

go 1.17

require (
	github.com/go-resty/resty/v2 v2.7.0
	github.com/goccha/logging v0.0.4
	github.com/rs/zerolog v1.26.1
)

require (
	github.com/goccha/http-constants v0.0.3 // indirect
	golang.org/x/net v0.0.0-20220531201128-c960675eff93 // indirect
)

replace github.com/goccha/logging => ../.
