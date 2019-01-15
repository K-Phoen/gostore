package gostore

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
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

func extractUntil(input string, delimiter string) (string, string, error) {
	delimiterPos := strings.Index(input, delimiter)
	if delimiterPos == -1 {
		return "", "", errors.New(fmt.Sprintf("Could not extract until delimiter %q from input %q", delimiter, input))
	}

	// FIXME what if delimiterPos+1 doesn't exist?
	return input[:delimiterPos], input[delimiterPos+1:], nil
}

func (server Server) parseCommand(conn net.Conn) (Command, error) {
	reader := bufio.NewReader(conn)

	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse command")
	}

	// remove the trailing \n
	line = line[:len(line)-1]

	// the action is the first word
	action, arguments, err := extractUntil(string(line), " ")
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse action")
	}

	switch action {
	case "store":
		return NewStoreCmd(arguments)
	case "fetch":
		return NewFetchCmd(arguments)
	case "del":
		return NewDelCmd(arguments)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown action %q", action))
	}
}

func (server Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	cmd, err := server.parseCommand(conn)
	if err != nil {
		server.logger.Printf("Invalid command received: %s", err)
		return
	}

	res, err := cmd.execute(server.store)
	if err != nil {
		server.logger.Printf("Error while executing command: %s", err)
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
