package gostore

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type Command interface {
	execute(store *Store) (Result, error)
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

func (cmd *StoreCmd) execute(store *Store) (Result, error) {
	store.Set(cmd.key, cmd.value)

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

func (cmd *FetchCmd) execute(store *Store) (Result, error) {
	val, _ := store.Get(cmd.key)

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

func (cmd *DelCmd) execute(store *Store) (Result, error) {
	store.Delete(cmd.key)

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
	default:
		return nil, errors.New(fmt.Sprintf("Unknown action %q", action))
	}
}