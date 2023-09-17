package config

const (
	ServiceName    = "echo-server"
	ServiceVersion = "v0.1.1"

	Env  = "demo"
	Port = ":8080"

	JaegerCollectorEndpoint = "http://jaeger-collector:14268/api/traces"
)
