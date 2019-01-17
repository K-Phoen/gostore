package gostore

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestValidStoreCmdParsing(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("store some-key some-value\n"))

	require.NoError(t, err, "Parsing a valid store command should not return errors")
	require.IsType(t, &StoreCmd{}, cmd)

	storeCmd := cmd.(*StoreCmd)
	require.Equal(t, "some-key", storeCmd.key)
	require.Equal(t, "some-value", storeCmd.value)
}

func TestValidFetchCmdParsing(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("fetch some-key\n"))

	require.NoError(t, err, "Parsing a valid fetch command should not return errors")
	require.IsType(t, &FetchCmd{}, cmd)

	fetchCmd := cmd.(*FetchCmd)
	require.Equal(t, "some-key", fetchCmd.key)
}

func TestValidDelCmdParsing(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("del some-key\n"))

	require.NoError(t, err, "Parsing a valid del command should not return errors")
	require.IsType(t, &DelCmd{}, cmd)

	delCmd := cmd.(*DelCmd)
	require.Equal(t, "some-key", delCmd.key)
}

func TestInvalidCommandsReturnErrors(t *testing.T) {
	commands := []string{
		"",
		"\n",

		"fetch\n",
		"fetch \n",
		"fetch some-key",

		"del\n",
		"del \n",
		"del some-key",

		"store\n",
		"store \n",
		"store some-key\n",
		"store some-key \n",
		"store some-key some-value",

		"unknown some-key\n",
	}

	for _, input := range commands {
		cmd, err := parseCommand(strings.NewReader(input))

		require.Error(t, err, fmt.Sprintf("Parsing an valid command should return an error (input: %q", input))
		require.Nil(t, cmd, "Parsing an invalid command should not return a Command struct")
	}
}