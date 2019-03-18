package main

import (
	"flag"
	"fmt"
	"github.com/K-Phoen/gostore/client"
	"github.com/peterh/liner"
	"io"
	"os"
	"path/filepath"
)

func executeRequest(gostore client.Client, request string, out io.Writer) {
	res, err := gostore.Exec(request)
	if err != nil {
		out.Write([]byte("Error: "))
		out.Write([]byte(err.Error()))
		out.Write([]byte{'\n'})
		return
	}

	out.Write([]byte(res))
	out.Write([]byte{'\n'})
}

func main() {
	var host string
	var port int
	var historyFile = filepath.Join(os.TempDir(), ".gostore_history")

	flag.StringVar(&host, "host", "0.0.0.0", "Host")
	flag.IntVar(&port, "port", 4224, "Port")

	flag.Parse()

	gostore := client.Client{
		Host: host,
		Port: port,
	}

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		if request, err := line.Prompt("> "); err == nil {
			executeRequest(gostore, request, os.Stdout)
			line.AppendHistory(request)
		} else if err == liner.ErrPromptAborted {
			fmt.Println("Aborted")
			break
		} else {
			fmt.Printf("\nError reading line: %s\n", err)
		}
	}

	if f, err := os.Create(historyFile); err != nil {
		fmt.Printf("Error writing history file: %s\n", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}