package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/K-Phoen/gostore"
)

type simulatedNode struct {
	name string
	keys map[string]bool
}

func (node simulatedNode) Address() string {
	return node.name
}

func (node simulatedNode) keysCount() int {
	return len(node.keys)
}

func (node *simulatedNode) AddKey(key string) {
	node.keys[key] = true
}

func (node *simulatedNode) RemoveKey(key string) {
	delete(node.keys, key)
}

type fakeServer struct {
	nodesMap map[string]*simulatedNode
	router gostore.Router
	keysCount int
}

func newFakeServer() *fakeServer {
	return &fakeServer{
		nodesMap: make(map[string]*simulatedNode),
		router: gostore.NewRouter(),
	}
}

func (server *fakeServer) addNodes(nodes int) {
	expectedNodeCount := len(server.nodesMap) + nodes

	for i := len(server.nodesMap); i < expectedNodeCount; i++ {
		name := fmt.Sprintf("node-%d", i+1)
		seed := rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()
		bytes := make([]byte, binary.MaxVarintLen64)
		binary.PutUvarint(bytes, seed)

		node := &simulatedNode{name: name, keys: make(map[string]bool)}

		server.nodesMap[node.Address()] = node
		server.router.AddNode(node, bytes)
	}
}

func (server *fakeServer) simulateNodeRemoval(nodes int) (int, int) {
	var removedNodes []*simulatedNode

	// select the nodes that will be removed
	for _, node := range server.nodesMap {
		removedNodes = append(removedNodes, node)

		if len(removedNodes) >= nodes {
			break
		}
	}

	keysToMoveFromRemovedNodes := 0
	keysToMoveFromOtherNodes := 0

	// remove the nodes from the router and server
	for _, node := range removedNodes {
		delete(server.nodesMap, node.Address())
		server.router.RemoveNode(node)

		keysToMoveFromRemovedNodes += node.keysCount()
	}

	// check if keys have to be moved from "alive" nodes
	for _, node := range server.nodesMap {
		for key := range node.keys {
			newNode := server.router.ResponsibleNode(key)

			if newNode.Address() == node.Address() {
				continue
			}

			keysToMoveFromOtherNodes += 1

			server.nodesMap[newNode.Address()].AddKey(key)
		}
	}

	// and relocate the keys from the removed nodes
	for _, node := range removedNodes {
		for key := range node.keys {
			newNode := server.router.ResponsibleNode(key)

			server.nodesMap[newNode.Address()].AddKey(key)
		}
	}

	return keysToMoveFromRemovedNodes, keysToMoveFromOtherNodes
}

func (server *fakeServer) simulateNodeAddition(nodes int) int {
	keysMoved := 0

	// add "empty" nodes to the server and router
	server.addNodes(nodes)

	// for each nodes, see if it has keys that should be relocated
	for _, node := range server.nodesMap {
		for key := range node.keys {
			newNode := server.router.ResponsibleNode(key)

			if newNode.Address() == node.Address() {
				continue
			}

			server.nodesMap[newNode.Address()].AddKey(key)
			server.nodesMap[node.Address()].RemoveKey(key)

			keysMoved += 1
		}
	}

	return keysMoved
}

func (server *fakeServer) insertKeys(keys int) {
	for i := 0; i < keys; i++ {
		key := fmt.Sprintf("key-%d", i)
		node := server.router.ResponsibleNode(key)

		server.nodesMap[node.Address()].AddKey(key)
	}

	server.keysCount += keys
}

func (server fakeServer) printSummary() {
	var standardDeviation float64
	nodes := len(server.nodesMap)
	mean := float64(server.keysCount/nodes)

	fmt.Printf("Summary\n=======\n")

	for _, node := range server.nodesMap {
		standardDeviation += math.Pow(float64(node.keysCount()) - mean, 2)
		fmt.Printf("%8s: % 4d keys\n", node.Address(), node.keysCount())
	}

	standardDeviation = math.Sqrt(standardDeviation/float64(nodes))

	fmt.Printf("\nPerfect routing would give %.0f keys per node\n", mean)
	fmt.Printf("Standard deviation: ~%.2f\n", standardDeviation)
}

func main() {
	var keys int
	var nodes int
	var clusterActivity int

	flag.IntVar(&keys, "keys", 100, "Number of keys to route")
	flag.IntVar(&nodes, "nodes", 10, "Number of available nodes")
	flag.IntVar(&clusterActivity, "movement", 0, "Number of nodes to add/remove to simulate cluster activity (positive number means that nodes will be added, negative means that they will be removed)")

	flag.Parse()

	fmt.Printf("Simulating routing for %d keys with %d nodes\n", keys, nodes)

	store := newFakeServer()

	store.addNodes(nodes)
	store.insertKeys(keys)

	store.printSummary()

	if clusterActivity == 0 {
		return
	}

	fmt.Printf("\n-----\n\n")

	if clusterActivity < 0 {
		fmt.Printf("Simulating the removal of %d nodes\n", -clusterActivity)

		keysToMoveFromRemovedNodes, keysToMoveFromOthers := store.simulateNodeRemoval(-clusterActivity)

		fmt.Printf("To remove %d nodes, %d keys have to be moved from nodes still in the cluster (and %d more, from the removed nodes)\n", -clusterActivity, keysToMoveFromOthers, keysToMoveFromRemovedNodes)
	} else {
		fmt.Printf("Simulating the addition of %d nodes\n", clusterActivity)

		keysToMove := store.simulateNodeAddition(clusterActivity)

		fmt.Printf("To add %d nodes, %d keys have to be moved\n", clusterActivity, keysToMove)
	}

	store.printSummary()
}