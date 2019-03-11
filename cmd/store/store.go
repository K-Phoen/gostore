package main

import (
	"flag"
	"github.com/K-Phoen/gostore"
	"github.com/sirupsen/logrus"
)

func main() {
	var cluster string

	config := gostore.DefaultConfig()
	config.StoragePath = "/tmp/gostore"

	flag.StringVar(&cluster, "cluster", "", "Cluster to join")

	flag.StringVar(&config.Host, "host", config.Host, "Host to listen to")
	flag.IntVar(&config.Port, "port", config.Port, "Port to listen to")

	flag.Parse()

	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		QuoteEmptyFields: true,
	})

	server := gostore.NewServer(logger, config)

	go server.Start()

	if len(cluster) > 0 {
		server.JoinCluster(cluster)
	}

	forever := make(chan bool)
	<-forever
}
