package gostore

import (
	"github.com/dgryski/go-farm"
	"sync"
)

type Router struct {
	nodes map[Node]uint64

	mutex sync.RWMutex
}

func NewRouter() Router {
	return Router{
		nodes: make(map[Node]uint64),
	}
}

func (router *Router) AddNode(node Node) {
	router.mutex.Lock()
	router.nodes[node] = router.hash(node.Address())
	router.mutex.Unlock()
}

func (router *Router) RemoveNode(node Node) {
	router.mutex.Lock()
	delete(router.nodes, node)
	router.mutex.Unlock()
}

func (router *Router) ResponsibleNode(key string) Node {
	router.mutex.RLock()
	defer router.mutex.RUnlock()

	var candidate Node
	maxScore := uint64(0)

	keyHash := router.hash(key)

	// fixme: bug if two nodes have the save score (iterating over a map doesn't guarantee the order, so two servers
	//  	  could choose different nodes)
	for node, nodeHash := range router.nodes {
		score := router.mergeHash(nodeHash, keyHash)

		if score > maxScore {
			maxScore = score
			candidate = node
		}
	}

	return candidate
}

func (router Router) hash(key string) uint64 {
	return farm.Hash64([]byte(key))
}

func (router Router) mergeHash(serverHash, keyHash uint64) uint64 {
	a := uint64(1103515245)
	b := uint64(12345)

	return 	(a * ((a * serverHash + b) ^ keyHash) + b) % (2^63)
}
