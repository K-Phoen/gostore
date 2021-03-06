package gostore

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type routerTestSuite struct {
	suite.Suite

	router Router
}

func (suite *routerTestSuite) SetupTest() {
	suite.router = NewRouter()

	suite.router.AddNode(NodeRef{host: "192.168.1.20", port: 4242})
	suite.router.AddNode(NodeRef{host: "192.168.1.30", port: 4242})
	suite.router.AddNode(NodeRef{host: "192.168.1.40", port: 4242})
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(routerTestSuite))
}

func (suite *routerTestSuite) TestItRoutesKeysBetweenNodes() {
	tt := []struct {
		key             string
		responsibleNode string
	}{
		{
			"some-key",
			"192.168.1.20:4242",
		},
		{
			"some-other-key",
			"192.168.1.30:4242",
		},
		{
			"yet-another-key",
			"192.168.1.30:4242",
		},
		{
			"last-key-promise",
			"192.168.1.40:4242",
		},
	}

	require := suite.Require()

	for _, tc := range tt {
		node := suite.router.ResponsibleNode(tc.key)

		require.Equal(tc.responsibleNode, node.Address(), tc.key)
	}
}

func (suite *routerTestSuite) TestItDoesNotUseRemovedNodes() {
	require := suite.Require()

	suite.router.RemoveNode(NodeRef{host: "192.168.1.30", port: 4242})

	// this key should be routed to 192.168.1.30:4242
	node := suite.router.ResponsibleNode("some-key")

	require.Equal("192.168.1.20:4242", node.Address())
}
