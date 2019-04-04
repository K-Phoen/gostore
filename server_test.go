package gostore

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	logging "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

type serverTestSuite struct {
	suite.Suite

	server *Server
	port int
}

func (suite *serverTestSuite) SetupSuite() {
	config := DefaultConfig()
	logger, _ := logging.NewNullLogger()
	server := NewServer(logger, config)

	suite.server = &server
	suite.port = config.Port

	go server.Start()
}


func (suite *serverTestSuite) TearDownSuite() {
	suite.server.Stop()
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(serverTestSuite))
}

func sendRequest(require *require.Assertions, port int, payload []byte) []byte {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	require.NoError(err, "could not connect to test server")
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(time.Second))
	require.NoError(err, "could not set the deadline")

	_, err = conn.Write(payload)
	require.NoError(err, "could not send test payload")

	response, err := ioutil.ReadAll(bufio.NewReader(conn))
	require.NoError(err, "could not read result")

	return response
}

func (suite *serverTestSuite) itHandlesRequestsCorrectly(port int) {
	tt := []struct {
		test    string
		payload []byte
		want    []byte
	}{
		{
			"Data can be stored",
			[]byte("store key some-value\n"),
			[]byte("+0\n"),
		},
		{
			"Data can be fetched",
			[]byte("fetch key\n"),
			[]byte("+10\nsome-value"),
		},
		{
			"Data can be fetched",
			[]byte("fetch unknown-key\n"),
			[]byte("+0\n"),
		},
		{
			"Data can be deleted",
			[]byte("del key\n"),
			[]byte("+0\n"),
		},
		{
			"Data can be deleted twice",
			[]byte("del key\n"),
			[]byte("+0\n"),
		},
		{
			"A local command can be executed",
			[]byte("node stats\n"),
			[]byte("+7\nKeys: 0"),
		},
		{
			"Invalid requests do not crash the server",
			[]byte("store key \n"),
			[]byte("-14\nNo value given"),
		},
	}

	test := suite.Require()

	for _, tc := range tt {
		suite.Run(tc.test, func() {
			response := sendRequest(test, port, tc.payload)

			test.Equal(tc.want, response)
		})
	}
}

func (suite *serverTestSuite) TestASingleNodeHandlesRequests() {
	suite.itHandlesRequestsCorrectly(suite.port)
}

func (suite *serverTestSuite) TestItHandlesExpiringKeys() {
	test := suite.Require()

	response := sendRequest(test, suite.port, []byte("storex expiring-key 1s some-value\n"))
	test.Equal([]byte("+0\n"), response)

	response = sendRequest(test, suite.port, []byte("fetch expiring-key\n"))
	test.Equal([]byte("+10\nsome-value"), response)

	time.Sleep(time.Second)

	response = sendRequest(test, suite.port, []byte("fetch expiring-key\n"))
	test.Equal([]byte("+0\n"), response)
}

func (suite *serverTestSuite) TestWithATwoNodesCluster() {
	config := DefaultConfig()
	config.Port = 5225

	logger, _ := logging.NewNullLogger()

	secondNode := NewServer(logger, config)

	go secondNode.Start()
	defer secondNode.Stop()

	secondNode.JoinCluster(fmt.Sprintf("127.0.0.1:%d", suite.port+1))

	// it should behave the same from both nodes
	suite.itHandlesRequestsCorrectly(suite.port)
	suite.itHandlesRequestsCorrectly(config.Port)

	test := suite.Require()

	for i := 0; i < 10; i++ {
		sendRequest(test, suite.port, []byte(fmt.Sprintf("store some-key-%d some-value\n", i)))
	}

	test.NotEqual(0, suite.server.store.Len(), "The first node should have at least some keys")
	test.NotEqual(0, secondNode.store.Len(), "The second node should have at least some keys")
}

func (suite *serverTestSuite) TestKeyStabilization() {
	configA := DefaultConfig()
	configA.Port = 6226
	configA.StabilizeBatchSize = 100
	configB := DefaultConfig()
	configB.Port = 7227
	configB.StabilizeBatchSize = 100
	configC := DefaultConfig()
	configC.Port = 8228
	configC.StabilizeBatchSize = 100

	logger, _ := logging.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	nodeA := NewServer(newPrefixedLogger(logger, "[A] "), configA)
	nodeB := NewServer(newPrefixedLogger(logger, "[B] "), configB)
	nodeC := NewServer(newPrefixedLogger(logger, "[C] "), configC)

	go nodeA.Start()
	go nodeB.Start()
	go nodeC.Start()
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	// stabilizing a node which is not part of a cluster (yet) does not fail
	nodeA.stabilize()

	nodeA.JoinCluster(fmt.Sprintf("127.0.0.1:%d", configB.Port+1))

	test := suite.Require()

	for i := 0; i < 30; i++ {
		sendRequest(test, configA.Port, []byte(fmt.Sprintf("store some-key-%d some-value\n", i)))
		sendRequest(test, configA.Port, []byte(fmt.Sprintf("storex expiring-key-%d 10s some-value\n", i)))
	}

	test.NotEqual(0, nodeA.store.Len(), "The first node should have at least some keys")
	test.NotEqual(0, nodeB.store.Len(), "The second node should have at least some keys")
	test.Equal(0, nodeC.store.Len(), "The last node has no keys as it did NOT join the cluster")

	// make the last node join the cluster (we join explicitely the two nodes to
	// avoid having to wait for the cluster discovery to happen)
	nodeC.JoinCluster(fmt.Sprintf("127.0.0.1:%d", configB.Port+1))
	nodeC.JoinCluster(fmt.Sprintf("127.0.0.1:%d", configA.Port+1))

	// force stabilization routine to run on older nodes
	nodeA.stabilize()
	nodeB.stabilize()

	// each key is stabilized in its goroutine, so let's wait for them to complete
	time.Sleep(300 * time.Millisecond)

	test.NotEqual(0, nodeA.store.Len(), "The first node should have at least some keys")
	test.NotEqual(0, nodeB.store.Len(), "The second node should have at least some keys")
	test.NotEqual(0, nodeC.store.Len(), "The last node should have data after the stabilization process")
}