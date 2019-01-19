package main

import (
	"flag"
	"github.com/K-Phoen/gostore"
	"log"
	"os"
)

func main() {
	var cluster string

	config := gostore.DefaultConfig()

	flag.StringVar(&cluster, "cluster", "", "Cluster to join")

	flag.StringVar(&config.Host, "host", config.Host, "Host to listen to")
	flag.IntVar(&config.Port, "port", config.Port, "Port to listen to")

	flag.Parse()

	logger := log.New(os.Stdout, "gostore ", log.Flags()|log.Lshortfile)
	server := gostore.NewServer(logger, config)

	go server.Start()

	if len(cluster) > 0 {
		server.JoinCluster(cluster)
	}

	forever := make(chan bool)
	<-forever
}
