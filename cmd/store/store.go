package main

import (
	"flag"
	"fmt"
	"github.com/K-Phoen/gostore"
	"github.com/sirupsen/logrus"
)

func main() {
	var cluster string

	config := gostore.DefaultConfig()

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

	config.StoragePath = fmt.Sprintf("/tmp/gostore-%d", config.Port)

	server := gostore.NewServer(logger, config)

	go server.Start()

	if len(cluster) > 0 {
		server.JoinCluster(cluster)
	}

	forever := make(chan bool)
	<-forever
}
