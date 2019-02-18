package gostore

import (
	"encoding/binary"
	"fmt"
	"github.com/dgryski/go-farm"
	"github.com/hashicorp/memberlist"
	"log"
	"math/rand"
	"os"
	"time"
)

type Cluster struct {
	logger *log.Logger
	memberList *memberlist.Memberlist
	seed uint64
	seedsMap map[string]uint64
}

func (cluster *Cluster) createMemberList(port int) {
	hostNumber := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	hostName, _ := os.Hostname()

	config := memberlist.DefaultLocalConfig()
	config.Name = fmt.Sprintf("%s-%X", hostName, hostNumber)
	config.BindPort = port
	config.AdvertisePort = port
	config.Logger = cluster.logger
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
	// todo error checking
	remoteSeed, _ := binary.Uvarint(node.Meta)

	cluster.seedsMap[node.Address()] = remoteSeed
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (cluster *Cluster) NotifyLeave(node *memberlist.Node) {

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
	b := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(b, cluster.seed)

	return b
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

func (cluster Cluster) LocalNode() *memberlist.Node {
	return cluster.memberList.LocalNode()
}

func (cluster Cluster) ResponsibleNode(key string) *memberlist.Node {
	var node *memberlist.Node
	maxScore := uint64(0)

	for _, member := range cluster.memberList.Members() {
		seed, ok := cluster.seedsMap[member.Address()]
		if !ok {
			cluster.logger.Printf("Could not find seed for node %s", member.Address())
			continue
		}

		score := cluster.hash(key, seed)

		if score > maxScore {
			maxScore = score
			node = member
		}
	}

	return node
}

func (cluster Cluster) hash(key string, seed uint64) uint64 {
	return farm.Hash64WithSeed([]byte(key), seed)
}

func (cluster *Cluster) Join(member string) error {
	_, err := cluster.memberList.Join([]string{member})

	return err
}

func NewCluster(logger *log.Logger, port int) *Cluster {
	cluster := &Cluster{
		logger: logger,
		seed: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
		seedsMap: make(map[string]uint64),
	}

	cluster.createMemberList(port)

	return cluster
}
