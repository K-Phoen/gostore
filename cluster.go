package gostore

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
	"math/rand"
	"os"
	"time"
)

type Cluster struct {
	logger *log.Logger
	memberList *memberlist.Memberlist
}

func (cluster *Cluster) createMemberList(port int) {
	hostNumber := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	hostName, _ := os.Hostname()

	config := memberlist.DefaultLocalConfig()
	config.Name = fmt.Sprintf("%s-%X", hostName, hostNumber)
	config.BindPort = port
	config.AdvertisePort = port
	config.Logger = cluster.logger

	list, err := memberlist.Create(config)
	if err != nil {
		cluster.logger.Fatalf("Failed to create memberlist: %s", err)
	}

	cluster.memberList = list
}

func (cluster *Cluster) Join(member string) error {
	_, err := cluster.memberList.Join([]string{member})

	return err
}

func NewCluster(logger *log.Logger, port int) *Cluster {
	cluster := &Cluster{
		logger: logger,
	}

	cluster.createMemberList(port)

	return cluster
}
