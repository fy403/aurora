package tracers

// Uncomment the import statement for the jaeger tracer.
// make sure you run dep ensure to pull in the jaeger client
//
import (
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// SetupTracer is the place where you'd setup your specific tracer.
// The jaeger tracer is given as an example.
// To capture the jaeger traces you should run the jaeger backend.
// This can be done using the following docker command:
//
// `docker run -ti --rm -p6831:6831/udp -p16686:16686 jaegertracing/all-in-one:latest`
//
// The collector will be listening on localhost:6831
// and the query UI is reachable on localhost:16686.
func SetupTracer(serviceName, collectorEndpoint string, logSpans bool) (func(), error) {

	// Jaeger setup code
	config := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1, // It either samples all traces (sampler.param=1) or none of them (sampler.param=0)
		},
		Reporter: &config.ReporterConfig{
			// LocalAgentHostPort: localAgentHostPort, // "localhost:6831"
			CollectorEndpoint: collectorEndpoint, //"http://127.0.0.1:14268/api/traces"
			LogSpans:          logSpans,          // true
		},
	}

	closer, err := config.InitGlobalTracer(serviceName)
	if err != nil {
		return nil, err
	}

	cleanupFunc := func() {
		closer.Close()
	}

	return cleanupFunc, nil
}
