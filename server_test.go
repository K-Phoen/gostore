package gostore

import (
	"bufio"
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
}

func (suite *serverTestSuite) SetupSuite() {
	config := DefaultConfig()
	logger, _ := logging.NewNullLogger()
	server := NewServer(logger, config)

	suite.server = &server

	go server.Start()
}


func (suite *serverTestSuite) TearDownSuite() {
	suite.server.Stop()
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(serverTestSuite))
}

func sendRequest(require *require.Assertions, payload []byte) []byte {
	conn, err := net.Dial("tcp", ":4224")
	require.NoError(err, "could not connect to test server")
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(1*time.Second))

	_, err = conn.Write(payload)
	require.NoError(err, "could not send test payload")

	response, err := ioutil.ReadAll(bufio.NewReader(conn))
	require.NoError(err, "could not read result")

	return response
}

func (suite *serverTestSuite) TestItHandlesRequests() {
	tt := []struct {
		test    string
		payload []byte
		want    []byte
	}{
		{
			"A local command can be executed",
			[]byte("node stats\n"),
			[]byte("+7\nKeys: 0"),
		},
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
			"Invalid requests do not crash the server",
			[]byte("store key \n"),
			[]byte("-14\nNo value given"),
		},
	}

	test := suite.Require()

	for _, tc := range tt {
		suite.Run(tc.test, func() {
			response := sendRequest(test, tc.payload)

			test.Equal(tc.want, response)
		})
	}
}

func (suite *serverTestSuite) TestItHandlesExpiringKeys() {
	test := suite.Require()

	response := sendRequest(test, []byte("storex expiring-key 1s some-value\n"))
	test.Equal([]byte("+0\n"), response)

	response = sendRequest(test, []byte("fetch expiring-key\n"))
	test.Equal([]byte("+10\nsome-value"), response)

	time.Sleep(time.Second)

	response = sendRequest(test, []byte("fetch expiring-key\n"))
	test.Equal([]byte("+0\n"), response)
}