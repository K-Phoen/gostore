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
	fmt.Stringer

	execute(server *Server) (Result, error)
	distributed() bool
	hashingKey() string
}

type Result interface {
	fmt.Stringer
}

type VoidResult struct{}

type PayloadResult struct {
	data string
}

type distributedCmd struct {
}

type localCmd struct {
}

type StoreCmd struct {
	distributedCmd

	key   string
	value string
}

type FetchCmd struct {
	distributedCmd

	key string
}

type DelCmd struct {
	distributedCmd

	key string
}

type NodeStatsCmd struct {
	localCmd
}

type ClusterListNodesCmd struct {
	localCmd
}

type ClusterStatsCmd struct {
	localCmd
}

type ClusterJoinCmd struct {
	localCmd

	address string
}

func (r VoidResult) String() string {
	return "OK"
}

func (r PayloadResult) String() string {
	return fmt.Sprintf("%d\n%s", len(r.data), r.data)
}

func (cmd distributedCmd) distributed() bool {
	return true
}

func (cmd localCmd) distributed() bool {
	return false
}

func (cmd localCmd) hashingKey() string {
	return ""
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

func (cmd StoreCmd) hashingKey() string {
	return cmd.key
}

func (cmd StoreCmd) String() string {
	return fmt.Sprintf("store %s %s", cmd.key, cmd.value)
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
	val, _, _ := server.store.Get(cmd.key)

	return PayloadResult{
		data: val,
	}, nil
}

func (cmd FetchCmd) hashingKey() string {
	return cmd.key
}

func (cmd FetchCmd) String() string {
	return fmt.Sprintf("fetch %s", cmd.key)
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

func (cmd DelCmd) hashingKey() string {
	return cmd.key
}

func (cmd DelCmd) String() string {
	return fmt.Sprintf("del %s", cmd.key)
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

func (cmd ClusterListNodesCmd) String() string {
	return "cluster nodes"
}

func NewNodeStatsCmd() (*NodeStatsCmd, error) {
	return &NodeStatsCmd{}, nil
}

func (cmd *NodeStatsCmd) execute(server *Server) (Result, error) {
	return PayloadResult{data: fmt.Sprintf("Keys: %d", server.store.Len())}, nil
}

func (cmd NodeStatsCmd) String() string {
	return "node stats"
}

func NewClusterStatsCmd() (*ClusterStatsCmd, error) {
	return &ClusterStatsCmd{}, nil
}

func (cmd *ClusterStatsCmd) execute(server *Server) (Result, error) {
	var buffer bytes.Buffer
	nodeCmd := &NodeStatsCmd{}

	buffer.WriteString(fmt.Sprintf("%s\n", server.cluster.LocalNode().Address()))
	buffer.WriteString(fmt.Sprintf("Keys: %d\n", server.store.Len()))

	for _, member := range server.cluster.Members() {
		if server.cluster.LocalNode().Address() == member.Address() {
			continue
		}

		buffer.WriteString("---\n")
		buffer.WriteString(fmt.Sprintf("%s\n", member.Address()))
		server.relayCommand(&buffer, nodeCmd, member)
	}

	return PayloadResult{data: buffer.String()}, nil
}

func (cmd ClusterStatsCmd) String() string {
	return "cluster stats"
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

func (cmd ClusterJoinCmd) String() string {
	return fmt.Sprintf("cluster join %s", cmd.address)
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
	case "stats":
		return NewClusterStatsCmd()
	}

	// then, try to parse subcommands that do have arguments
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

func parseNodeCommand(input string) (Command, error) {
	// first, handle the subcommands that do NOT have any argument
	switch input {
	case "stats":
		return NewNodeStatsCmd()
	}

	return nil, errors.New(fmt.Sprintf("Unknown cluster subcommand %q", input))
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
	case "node":
		return parseNodeCommand(arguments)
	case "cluster":
		return parseClusterCommand(arguments)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown action %q", action))
	}
}