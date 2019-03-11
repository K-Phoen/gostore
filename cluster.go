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

	config := memberlist.DefaultLocalConfig()
	config.Name = fmt.Sprintf("%s-%X", hostName, hostNumber)
	config.BindPort = port
	config.AdvertisePort = port
	config.Logger = log.New(cluster.logger.Writer(), "", 0)
	config.Delegate = cluster
	config.Events = cluster

	list, err := memberlist.Create(config)
	if err != nil {
		cluster.logger.Fatalf("Failed to create memberlist: %s", err)
	}

	cluster.memberList = list
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (cluster *Cluster) NotifyJoin(node *memberlist.Node) {
	cluster.router.AddNode(NodeRef{host: node.Addr.String(), port: node.Port-1}, node.Meta)
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (cluster *Cluster) NotifyLeave(node *memberlist.Node) {
	cluster.router.RemoveNode(NodeRef{host: node.Addr.String(), port: node.Port-1})
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (cluster *Cluster) NotifyUpdate(node *memberlist.Node) {

}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (cluster *Cluster) NodeMeta(limit int) []byte {
	return cluster.router.SeedBytes()
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (cluster *Cluster) NotifyMsg(msg []byte) {
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (cluster *Cluster) GetBroadcasts(overhead, limit int) [][]byte {
	var b [][]byte

	return b
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (cluster *Cluster) LocalState(join bool) []byte {
	var b []byte

	return b
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (cluster *Cluster) MergeRemoteState(buf []byte, join bool) {

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

func NewCluster(logger *logrus.Logger, port int) *Cluster {
	cluster := &Cluster{
		logger: logger,
		router: NewRouter(),
	}

	cluster.createMemberList(port)

	return cluster
}
