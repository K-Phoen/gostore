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
	require.True(t, storeCmd.distributed())
	require.Equal(t, "some-key", storeCmd.hashingKey())
	require.Equal(t, "store some-key some-value", storeCmd.String())
}

func TestValidFetchCmdParsing(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("fetch some-key\n"))

	require.NoError(t, err, "Parsing a valid fetch command should not return errors")
	require.IsType(t, &FetchCmd{}, cmd)

	fetchCmd := cmd.(*FetchCmd)
	require.Equal(t, "some-key", fetchCmd.key)
	require.True(t, fetchCmd.distributed())
	require.Equal(t, "some-key", fetchCmd.hashingKey())
	require.Equal(t, "fetch some-key", fetchCmd.String())
}

func TestValidDelCmdParsing(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("del some-key\n"))

	require.NoError(t, err, "Parsing a valid del command should not return errors")
	require.IsType(t, &DelCmd{}, cmd)

	delCmd := cmd.(*DelCmd)
	require.Equal(t, "some-key", delCmd.key)
	require.True(t, delCmd.distributed())
	require.Equal(t, "some-key", delCmd.hashingKey())
	require.Equal(t, "del some-key", delCmd.String())
}

func TestValidNodeStats(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("node stats\n"))

	require.NoError(t, err, "Parsing a valid node stats command should not return errors")
	require.IsType(t, &NodeStatsCmd{}, cmd)

	nodeStatsCmd := cmd.(*NodeStatsCmd)
	require.False(t, nodeStatsCmd.distributed())
	require.Empty(t, nodeStatsCmd.hashingKey())
	require.Equal(t, "node stats", nodeStatsCmd.String())
}

func TestValidClusterStats(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("cluster stats\n"))

	require.NoError(t, err, "Parsing a valid cluster stats command should not return errors")
	require.IsType(t, &ClusterStatsCmd{}, cmd)

	clusterStatsCmd := cmd.(*ClusterStatsCmd)
	require.False(t, clusterStatsCmd.distributed())
	require.Empty(t, clusterStatsCmd.hashingKey())
	require.Equal(t, "cluster stats", clusterStatsCmd.String())
}

func TestValidClusterListNodes(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("cluster nodes\n"))

	require.NoError(t, err, "Parsing a valid cluster list nodes command should not return errors")
	require.IsType(t, &ClusterListNodesCmd{}, cmd)

	clusterListNodesCmd := cmd.(*ClusterListNodesCmd)
	require.False(t, clusterListNodesCmd.distributed())
	require.Empty(t, clusterListNodesCmd.hashingKey())
	require.Equal(t, "cluster nodes", clusterListNodesCmd.String())
}

func TestValidClusterJoin(t *testing.T) {
	cmd, err := parseCommand(strings.NewReader("cluster join 192.168.1.42:2424\n"))

	require.NoError(t, err, "Parsing a valid cluster join command should not return errors")
	require.IsType(t, &ClusterJoinCmd{}, cmd)

	joinCmd := cmd.(*ClusterJoinCmd)
	require.Equal(t, "192.168.1.42:2424", joinCmd.address)
	require.False(t, joinCmd.distributed())
	require.Empty(t, joinCmd.hashingKey())
	require.Equal(t, "cluster join 192.168.1.42:2424", joinCmd.String())
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

		"node unknown\n",

		"cluster join\n",
		"cluster unknown\n",
		"cluster unknown arg\n",

		"unknown some-key\n",
	}

	for _, input := range commands {
		cmd, err := parseCommand(strings.NewReader(input))

		require.Error(t, err, fmt.Sprintf("Parsing an valid command should return an error (input: %q", input))
		require.Nil(t, cmd, "Parsing an invalid command should not return a Command struct")
	}
}