package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
)

func sendRequest(host string, port int, request string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Could not connect to server: %s", err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\n", request)
	status, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Printf("Status: %s\nErr: %s\n", status, err)
}

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "Host to listen to")
	flag.IntVar(&port, "port", 4224, "Port to listen to")

	flag.Parse()

	cmd := "store key some-value"
	count := 1000

	for i := 1; i <= count; i++ {
		if i%10 == 0 {
			fmt.Printf("Sending request %d/%d\n", i, count)
		}

		go sendRequest(host, port, cmd)
	}
}
