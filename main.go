package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Arg("config", "Load config from this file").Required().File()
	reload     = kingpin.Flag("-r", "Reload the config file automatically").Default("true").Bool()
	listenAddr = kingpin.Flag("--listen", "listen address").Short('L').Default(":8080").TCP()
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	kingpin.Parse()

	handler, err := ConfiguredProxyHandler(*configFile)
	if err != nil {
		fmt.Printf("fatal: %v\n", err)
		return
	}

	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, err := cfg.New(
		"your_service_name",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	log.Printf("s3-proxy is listening on %s\n", *listenAddr)
	log.Fatal(http.ListenAndServe((*listenAddr).String(), handler))
}
