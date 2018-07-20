package main

import (
	"fmt"
	"log"
	"net/http"

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

	log.Printf("s3-proxy is listening on %s\n", *listenAddr)
	log.Fatal(http.ListenAndServe((*listenAddr).String(), handler))
}
