package gostore

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Server struct {
	logger *log.Logger
	store  *Store
}

func (server Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	cmd, err := parseCommand(conn)
	if err != nil {
		server.logger.Printf("Invalid command received: %s", err)
		io.Copy(conn, strings.NewReader(fmt.Sprintf("ERR\n%s", err)))
		return
	}

	res, err := cmd.execute(server.store)
	if err != nil {
		server.logger.Printf("Error while executing command: %s", err)
		io.Copy(conn, strings.NewReader(fmt.Sprintf("ERR\n%s", err)))
		return
	}

	_, err = io.Copy(conn, strings.NewReader(res.String()))
	if err != nil {
		server.logger.Printf("Could not send response: %s", err)
	}
}

func (server Server) Start(host string, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		server.logger.Fatalf("Could not listen to %s:%d. %s", host, port, err)
	}

	server.logger.Printf("Listening to %s:%d", host, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			server.logger.Printf("Could not accept connection: %s", err)
		}

		go server.handleConnection(conn)
	}
}

func NewServer() Server {
	return Server{
		logger: log.New(os.Stdout, "gostore ", log.Flags()|log.Lshortfile),
		store:  NewStore(),
	}
}
