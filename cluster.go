package gostore

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type Node interface {
	Address() string
	SameAs(other Node) bool
}

type NodeRef struct {
	host string
	port uint16
}

type memberlistDelegate struct {
	router *Router
}

type Cluster struct {
	logger *logrus.Logger
	memberList *memberlist.Memberlist
	router Router
}

func (node NodeRef) Address() string {
	return net.JoinHostPort(node.host, strconv.Itoa(int(node.port)))
}

func (node NodeRef) SameAs(other Node) bool {
	return node.Address() == other.Address()
}

func (node NodeRef) String() string {
	return node.Address()
}

func (cluster *Cluster) createMemberList(port int) {
	hostNumber := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	hostName, _ := os.Hostname()

	delegate := &memberlistDelegate{router: &cluster.router}

	config := memberlist.DefaultLocalConfig()
	config.Name = fmt.Sprintf("%s-%X", hostName, hostNumber)
	config.BindPort = port
	config.AdvertisePort = port
	config.Logger = log.New(cluster.logger.Writer(), "", 0)
	config.Events = delegate

	list, err := memberlist.Create(config)
	if err != nil {
		cluster.logger.Fatalf("Failed to create memberlist: %s", err)
	}

	cluster.memberList = list
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (delegate *memberlistDelegate) NotifyJoin(node *memberlist.Node) {
	delegate.router.AddNode(NodeRef{host: node.Addr.String(), port: node.Port-1})
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (delegate *memberlistDelegate) NotifyLeave(node *memberlist.Node) {
	delegate.router.RemoveNode(NodeRef{host: node.Addr.String(), port: node.Port-1})
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (delegate *memberlistDelegate) NotifyUpdate(node *memberlist.Node) {
	// nothing to do
}

func (cluster Cluster) LocalNode() Node {
	local := cluster.memberList.LocalNode()

	return NodeRef{host: local.Addr.String(), port: local.Port-1}
}

func (cluster Cluster) Members() []Node {
	var nodes []Node

	for _, member := range cluster.memberList.Members() {
		nodes = append(nodes, NodeRef{host: member.Addr.String(), port: member.Port-1})
	}

	return nodes
}

func (cluster Cluster) ResponsibleNode(key string) Node {
	return cluster.router.ResponsibleNode(key)
}

func (cluster *Cluster) Join(member string) error {
	_, err := cluster.memberList.Join([]string{member})

	return err
}

func (cluster *Cluster) Shutdown() error {
	if cluster.memberList == nil {
		return nil
	}

	return cluster.memberList.Shutdown()
}

func NewCluster(logger *logrus.Logger, port int) *Cluster {
	cluster := &Cluster{
		logger: logger,
		router: NewRouter(),
	}

	cluster.createMemberList(port)

	return cluster
}
