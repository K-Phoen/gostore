package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"sync"
)

func sendRequest(host string, port int, request string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Could not connect to server: %s", err)
		return
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\n", request)
	if err != nil {
		fmt.Printf("Could not send request\n")
		return
	}

	status, err := bufio.NewReader(conn).ReadString('\n')

	if status != "OK" {
		fmt.Printf("Status: %s\nErr: %s\n", status, err)
	}
}

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "Host to listen to")
	flag.IntVar(&port, "port", 4224, "Port to listen to")

	flag.Parse()

	count := 1000
	var wg sync.WaitGroup

	for i := 1; i <= count; i++ {
		if i%10 == 0 {
			fmt.Printf("Sending request %d/%d\n", i, count)
		}

		wg.Add(1)
		go func(i int) {
			sendRequest(host, port, fmt.Sprintf("store key-%d some-value", i))
			wg.Done()
		}(i)
	}

	wg.Wait()
}
