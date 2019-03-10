package gostore

import (
	"bytes"
	"fmt"
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

	StabilizeInterval time.Duration
	// percentage of keys in the store to stabilize per batch
	StabilizeBatchSize int
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

		StabilizeInterval: 10 * time.Second,
		StabilizeBatchSize: 20, // percent
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
	if server.cluster.LocalNode().SameAs(responsibleNode) {
		server.execute(conn, cmd)
		return
	}

	server.relayCommand(conn, cmd, responsibleNode)
}

func (server Server) relayCommand(dest io.Writer, cmd Command, remote Node) {
	remoteConn, err := net.Dial("tcp", remote.Address())
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

	server.startStabilizationRoutine()

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

func (server *Server) startStabilizationRoutine() {
	// this could be triggered by "node joined" events instead of periodically
	ticker := time.NewTicker(server.config.StabilizeInterval)

	go func() {
		for {
			if server.stopped {
				ticker.Stop()
				break
			}

			select {
			case <- ticker.C:
				server.stabilize()
			}
		}
	}()
}

func (server *Server) stabilize() {
	server.logger.Printf("Starting stabilization routine")

	if len(server.cluster.Members()) < 2 {
		server.logger.Printf("Not enough nodes in the cluster for a stabilization to be needed")
		return
	}

	localNode := server.cluster.LocalNode()
	batchSize := int(float64(server.store.Len()) * float64(server.config.StabilizeBatchSize) / 100.0)
	stabilizedKeys := 0

	server.store.Keys(func (key string) bool {
		responsibleNode := server.cluster.ResponsibleNode(key)

		if !localNode.SameAs(responsibleNode) {
			go server.stabilizeKey(key, responsibleNode)

			stabilizedKeys++
		}

		return stabilizedKeys < batchSize
	})

	server.logger.Printf("Stabilized %d keys (maximum batch size: %d)", stabilizedKeys, batchSize)
}

func (server *Server) stabilizeKey(key string, remote Node) {
	value, _, err := server.store.Get(key)
	if err != nil {
		return
	}

	storeCmd := &StoreCmd{
		key: key,
		value: value,
	}
	delCmd := DelCmd{key: key}

	// send the key-value pair to the remote server
	storeBuffer := bytes.NewBufferString("")
	server.relayCommand(storeBuffer, storeCmd, remote)

	result, err := storeBuffer.ReadString('\n')
	if  result != "OK" && err != nil {
		server.logger.Printf("Could not stabilize key %q to node %q: %s", key, remote, err)
		return
	}

	if result != "OK" {
		server.logger.Printf("Could not stabilize key %q to node %s: %q", key, remote, storeBuffer.String())
		return
	}

	// delete our own copy of it
	_, err = delCmd.execute(server)
	if err != nil {
		server.logger.Printf("Could not delete local copy of stabilized key %q", key)
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
