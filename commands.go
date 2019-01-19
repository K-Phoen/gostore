package gostore

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type Command interface {
	execute(server *Server) (Result, error)
}

type Result interface {
	String() string
}

type VoidResult struct{}

type PayloadResult struct {
	data string
}

type StoreCmd struct {
	key   string
	value string
}

type FetchCmd struct {
	key string
}

type DelCmd struct {
	key string
}

type ClusterListNodesCmd struct {
}

type ClusterJoinCmd struct {
	address string
}

func (r VoidResult) String() string {
	return "OK"
}

func (r PayloadResult) String() string {
	return fmt.Sprintf("%d\n%s", len(r.data), r.data)
}

func NewStoreCmd(arguments string) (*StoreCmd, error) {
	key, rest, err := extractUntil(arguments, " ")
	if err != nil {
		return nil, errors.Wrap(err, "Could not extract key")
	}

	if len(rest) == 0 {
		return nil, errors.New("No value given")
	}

	return &StoreCmd{
		key:   key,
		value: rest,
	}, nil
}

func (cmd *StoreCmd) execute(server *Server) (Result, error) {
	server.store.Set(cmd.key, cmd.value)

	return VoidResult{}, nil
}

func NewFetchCmd(arguments string) (*FetchCmd, error) {
	if len(arguments) == 0 {
		return nil, errors.New("No key given")
	}

	return &FetchCmd{
		key: arguments,
	}, nil
}

func (cmd *FetchCmd) execute(server *Server) (Result, error) {
	val, _ := server.store.Get(cmd.key)

	return PayloadResult{
		data: val,
	}, nil
}

func NewDelCmd(arguments string) (*DelCmd, error) {
	if len(arguments) == 0 {
		return nil, errors.New("No key given")
	}

	return &DelCmd{
		key: arguments,
	}, nil
}

func (cmd *DelCmd) execute(server *Server) (Result, error) {
	server.store.Delete(cmd.key)

	return VoidResult{}, nil
}

func NewClusterListNodesCmd() (*ClusterListNodesCmd, error) {
	return &ClusterListNodesCmd{}, nil
}

func (cmd *ClusterListNodesCmd) execute(server *Server) (Result, error) {
	var buffer bytes.Buffer

	for _, member := range server.cluster.memberList.Members() {
		buffer.WriteString(fmt.Sprintf("%s %s:%d\n", member.Name, member.Addr, member.Port))
	}

	return PayloadResult{data: buffer.String()}, nil
}

func NewClusterJoinCmd(arguments string) (*ClusterJoinCmd, error) {
	if len(arguments) == 0 {
		return nil, errors.New("No address given")
	}

	return &ClusterJoinCmd{
		address: arguments,
	}, nil
}

func (cmd *ClusterJoinCmd) execute(server *Server) (Result, error) {
	err := server.cluster.Join(cmd.address)
	if err != nil {
		return nil, errors.Wrap(err, "Could not join cluster")
	}

	return VoidResult{}, nil
}

func extractUntil(input string, delimiter string) (string, string, error) {
	delimiterPos := strings.Index(input, delimiter)
	if delimiterPos == -1 {
		return "", "", errors.New(fmt.Sprintf("Could not extract until delimiter %q from input %q", delimiter, input))
	}

	// FIXME what if delimiterPos+1 doesn't exist?
	return input[:delimiterPos], input[delimiterPos+1:], nil
}

func parseClusterCommand(input string) (Command, error) {
	// first, handle the subcommands that do NOT have any argument
	switch input {
	case "nodes":
		return NewClusterListNodesCmd()
	}

	// then, try to parse subcommands that do have arguments
	fmt.Println("→", input)
	action, arguments, err := extractUntil(string(input), " ")
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse cluster subcommand")
	}

	switch action {
	case "join":
		return NewClusterJoinCmd(arguments)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown cluster subcommand %q", action))
	}
}

func parseCommand(reader io.Reader) (Command, error) {
	line, err := bufio.NewReader(reader).ReadBytes('\n')
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
	case "cluster":
		return parseClusterCommand(arguments)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown action %q", action))
	}
}