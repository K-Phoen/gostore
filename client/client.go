package client

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"strconv"
	"time"
)

type Client struct {
	Host string
	Port int
}

func (client Client) Get(key string) (string, error) {
	result, err := client.Exec("fetch "+key)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (client Client) Set(key, value string) error {
	_, err := client.Exec(fmt.Sprintf("store %s %s", key, value))

	return err
}

func (client Client) SetWithTTL(key, value string, lifetime time.Duration) error {
	_, err := client.Exec(fmt.Sprintf("storex %s %s %s", key, lifetime, value))

	return err
}

func (client Client) Delete(key string) error {
	_, err := client.Exec("del "+key)

	return err
}

func (client Client) Exec(request string) (string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\n", request)
	if err != nil {
		return "", err
	}

	return client.parseResult(conn)
}

func (client Client) parseResult(resultReader io.Reader) (string, error) {
	reader := bufio.NewReader(resultReader)
	status, err := reader.ReadByte()
	if err != nil {
		return "", errors.Wrap(err, "could not parse result status")
	}

	messageLength, err := reader.ReadBytes('\n')
	if err != nil {
		return "", errors.Wrap(err, "could not parse message length")
	}

	length, err := strconv.Atoi(string(messageLength[:len(messageLength)-1]))
	if err != nil {
		return "", errors.Wrap(err, "expected message length to be an int")
	}

	buffer := make([]byte, length)
	read, err := reader.Read(buffer)
	if err != nil {
		return "", errors.Wrap(err, "error while reading result")
	}

	if read < length {
		return "", errors.Wrap(err, fmt.Sprintf("expected to read %d bytes from result, read %d", length, read))
	}

	if status == '+' {
		return string(buffer), nil
	} else {
		return "", errors.New(string(buffer))
	}
}