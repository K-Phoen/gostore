package main

import (
	"flag"
	"fmt"
	"github.com/K-Phoen/gostore/client"
	"sync"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "Host")
	flag.IntVar(&port, "port", 4224, "Port")

	flag.Parse()

	gostore := client.Client{
		Host: host,
		Port: port,
	}

	count := 1000
	var wg sync.WaitGroup

	for i := 1; i <= count; i++ {
		if i%10 == 0 {
			fmt.Printf("Sending request %d/%d\n", i, count)
		}

		wg.Add(1)
		go func(i int) {
			err := gostore.Set(fmt.Sprintf("some-key-%d", i), "some-value")
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
