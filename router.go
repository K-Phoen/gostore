package gostore

import (
	"encoding/binary"
	"github.com/dgryski/go-farm"
	"math/rand"
	"sync"
	"time"
)

type Router struct {
	seed uint64
	seedsMap map[Node]uint64

	mutex sync.RWMutex
}

func NewRouter() Router {
	return Router {
		seed: rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(),
		seedsMap: make(map[Node]uint64),
	}
}

func (router Router) SeedBytes() []byte {
	b := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(b, router.seed)

	return b
}

func (router Router) AddNode(node Node, seed []byte) {
	// todo error checking
	remoteSeed, _ := binary.Uvarint(seed)

	router.mutex.Lock()
	router.seedsMap[node] = remoteSeed
	router.mutex.Unlock()
}

func (router Router) RemoveNode(node Node) {
	router.mutex.Lock()
	delete(router.seedsMap, node)
	router.mutex.Unlock()
}

func (router Router) ResponsibleNode(key string) Node {
	router.mutex.RLock()
	defer router.mutex.RUnlock()

	var candidate Node
	maxScore := uint64(0)

	for node, seed := range router.seedsMap {
		score := router.hash(key, seed)

		if score > maxScore {
			maxScore = score
			candidate = node
		}
	}

	return candidate
}

func (router Router) hash(key string, seed uint64) uint64 {
	return farm.Hash64WithSeed([]byte(key), seed)
}