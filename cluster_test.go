package gostore

import (
	"github.com/sirupsen/logrus"
	logging "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type clusterTestSuite struct {
	suite.Suite

	logger *logrus.Logger
}

func (suite *clusterTestSuite) SetupSuite() {
	logger, _ := logging.NewNullLogger()

	suite.logger = logger
}


func (suite *clusterTestSuite) TearDownSuite() {
}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(clusterTestSuite))
}

func (suite *clusterTestSuite) TestASingleNodeClusterCanBeCreated() {
	require := suite.Require()

	// if the server runs on the port 4224, the cluster management will use the port 4225
	cluster := NewCluster(suite.logger, 4225)
	defer cluster.Shutdown()

	require.Len(cluster.Members(),1, "A new cluster has only one member")

	localAddress := cluster.LocalNode().Address()
	localPort := strings.Split(localAddress, ":")

	require.Equal("4224", localPort[1],"It returns the service's port, not the management one")
}

func (suite *clusterTestSuite) TestAMultiNodesClusterCanBeCreated() {
	require := suite.Require()

	clusterA := NewCluster(suite.logger, 4224)
	clusterB := NewCluster(suite.logger, 5225)
	defer clusterA.Shutdown()
	defer clusterB.Shutdown()

	require.Len(clusterA.Members(),1, "A new cluster has only one member")
	require.Len(clusterB.Members(),1, "A new cluster has only one member")

	err := clusterA.Join("127.0.0.1:5225")
	require.NoError(err, "clusterA should be able to join clusterB")

	require.Len(clusterA.Members(),2)
	require.Len(clusterB.Members(),2)
}