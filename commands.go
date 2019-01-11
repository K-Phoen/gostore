package main

import (
	"fmt"
	"github.com/pkg/errors"
)

type Command interface {
	execute(store *Store) (Result, error)
}

type Result interface {
	String() string
}

type VoidResult struct {}

type PayloadResult struct {
	data string
}

type StoreCmd struct {
	key string
	value string
}

type FetchCmd struct {
	key string
}

type DelCmd struct {
	key string
}

func (r VoidResult) String() string{
	return "OK"
}

func (r PayloadResult) String() string{
	return fmt.Sprintf("%d\n%s", len(r.data), r.data)
}

func NewStoreCmd(arguments string) (*StoreCmd, error) {
	key, rest, err := extractUntil(arguments, " ")
	if err != nil {
		return nil, errors.Wrap(err, "Could not extract key")
	}

	return &StoreCmd{
		key: key,
		value: rest,
	}, nil
}

func (cmd *StoreCmd) execute(store *Store) (Result, error) {
	store.Set(cmd.key, cmd.value)

	return VoidResult{}, nil
}

func NewFetchCmd(arguments string) (*FetchCmd, error) {
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
	return &DelCmd{
		key: arguments,
	}, nil
}

func (cmd *DelCmd) execute(store *Store) (Result, error) {
	store.Delete(cmd.key)

	return VoidResult{}, nil
}