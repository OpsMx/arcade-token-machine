module github.com/OpsMx/arcade-token-machine

go 1.18

require (
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/skandragon/gohealthcheck v1.0.2
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.31.0
	go.opentelemetry.io/otel v1.6.1
	go.opentelemetry.io/otel/exporters/jaeger v1.6.1
	go.opentelemetry.io/otel/sdk v1.6.1
	go.opentelemetry.io/otel/trace v1.6.1
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

require (
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	golang.org/x/sys v0.0.0-20220405052023-b1e9470b6e64 // indirect
)
