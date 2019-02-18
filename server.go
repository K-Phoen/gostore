package gostore

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type Config struct {
	Host string
	Port int

	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	config Config

	logger *log.Logger
	store  *Store
	cluster *Cluster

	listener net.Listener
	stopped bool
}

func DefaultConfig() Config {
	return Config{
		Host: "0.0.0.0",
		Port: 4224,

		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

func (server Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(server.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(server.config.WriteTimeout))

	cmd, err := parseCommand(conn)
	if err != nil {
		server.logger.Printf("Invalid command received: %s", err)
		io.Copy(conn, strings.NewReader(fmt.Sprintf("ERR\n%s", err)))
		return
	}

	if !cmd.distributed() {
		server.execute(conn, cmd)
		return
	}

	responsibleNode := server.cluster.ResponsibleNode(cmd.hashingKey())

	// distributed command, but we happen to be the node responsible for it
	if server.cluster.LocalNode() == responsibleNode {
		server.execute(conn, cmd)
		return
	}

	server.relayCommand(conn, cmd, responsibleNode)
}

func (server Server) relayCommand(dest io.Writer, cmd Command, remote *memberlist.Node) {
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remote.Addr, remote.Port-1))
	if err != nil {
		server.logger.Printf("Could not connect to node: %s", remote.Address())
		return
	}
	defer remoteConn.Close()

	_, err = fmt.Fprintf(remoteConn, "%s\n", cmd)
	if err != nil {
		server.logger.Printf("Could not relay command to node: %s", remote.Address())
		return
	}

	io.Copy(dest, remoteConn)
}

func (server Server) execute(dest io.Writer, cmd Command) {
	res, err := cmd.execute(&server)
	if err != nil {
		server.logger.Printf("Error while executing command: %s", err)
		io.Copy(dest, strings.NewReader(fmt.Sprintf("ERR\n%s", err)))
		return
	}

	_, err = io.Copy(dest, strings.NewReader(res.String()))
	if err != nil {
		server.logger.Printf("Could not send response: %s", err)
	}
}

func (server Server) JoinCluster(member string) {
	err := server.cluster.Join(member)
	if err != nil {
		server.logger.Fatalf("Failed to join cluster: %s", err)
	}
}

func (server *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.config.Host, server.config.Port))
	if err != nil {
		server.logger.Fatalf("Could not listen to %s:%d. %s", server.config.Host, server.config.Port, err)
	}

	server.listener = listener
	server.logger.Printf("Listening to %s:%d", server.config.Host, server.config.Port)

	for {
		if server.stopped {
			break
		}

		conn, err := listener.Accept()
		if err != nil {
			server.logger.Printf("Could not accept connection: %s", err)
			continue
		}

		go server.handleConnection(conn)
	}
}

func (server *Server) Stop() {
	server.stopped = true

	err := server.listener.Close()
	if err != nil {
		server.logger.Printf("Error while stopping server: %s", err)
	}
}

func NewServer(logger *log.Logger, config Config) Server {
	return Server{
		logger: logger,
		config: config,
		store:  NewStore(),
		cluster: NewCluster(logger, config.Port+1),
	}
}
