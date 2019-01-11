package main

import (
	"flag"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "Host to listen to")
	flag.IntVar(&port, "port", 4224, "Port to listen to")

	flag.Parse()

	NewServer().Start(host, port)
}