module github.com/open-telemetry/opentelemetry-collector

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.1-0.20190430175949-e8b55949d948
	contrib.go.opencensus.io/exporter/ocagent v0.6.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.0 // indirect
	contrib.go.opencensus.io/resource v0.1.2
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7
	github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1
	github.com/client9/misspell v0.3.4
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/go-kit/kit v0.9.0
	github.com/gogo/googleapis v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6
	github.com/golang/protobuf v1.3.2
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/addlicense v0.0.0-20190510175307-22550fa7c1b0
	github.com/google/go-cmp v0.3.1
	github.com/google/go-containerregistry v0.0.0-20200212224832-c629a66d7231 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/grpc-ecosystem/grpc-gateway v1.11.1
	github.com/hashicorp/consul/api v1.2.0 // indirect
	github.com/jaegertracing/jaeger v1.14.0
	github.com/jstemmer/go-junit-report v0.0.0-20190106144839-af01ea7f8024
	github.com/mitchellh/mapstructure v1.1.2
	github.com/openzipkin/zipkin-go v0.2.1
	github.com/orijtech/prometheus-go-metrics-exporter v0.0.3-0.20190313163149-b321c5297f60
	github.com/pavius/impi v0.0.0-20180302134524-c1cbdcb8df2b
	github.com/pkg/errors v0.8.1
	github.com/prashantv/protectmem v0.0.0-20171002184600-e20412882b3a // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/common v0.7.0
	github.com/prometheus/procfs v0.0.3
	github.com/prometheus/prometheus v1.8.2-0.20190924101040-52e0504f83ea
	github.com/rs/cors v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.1-0.20190911140308-99520c81d86e
	github.com/streadway/quantile v0.0.0-20150917103942-b0c588724d25 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/uber/tchannel-go v1.10.0
	go.opencensus.io v0.22.1
	go.uber.org/zap v1.10.0
	golang.org/x/sys v0.0.0-20191010194322-b09406accb47
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	google.golang.org/api v0.10.0
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.4
	istio.io/client-go v0.0.0-20200212163417-ad75bb5565ef // indirect
	k8s.io/client-go v12.0.0+incompatible // indirect
	knative.dev/pkg v0.0.0-20200213212836-df0629984fa4 // indirect
	knative.dev/serving v0.12.1
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
