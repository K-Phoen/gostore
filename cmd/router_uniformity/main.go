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

type simulationResult struct {
	node string
	keys int
}

type simulatedNode struct {
	name string
}

func (node simulatedNode) Address() string {
	return node.name
}

func simulate(keys int, nodes int) []simulationResult {
	var results []simulationResult
	router := gostore.NewRouter()
	nodesMap := make(map[string]int)

	for i := 0; i < nodes; i++ {
		name := fmt.Sprintf("node-%d", i+1)
		seed := rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()
		bytes := make([]byte, binary.MaxVarintLen64)
		binary.PutUvarint(bytes, seed)

		router.AddNode(simulatedNode{name: name}, bytes)
	}

	for i := 0; i < keys; i++ {
		key := fmt.Sprintf("key-%d", i)

		node := router.ResponsibleNode(key)

		nodesMap[node.Address()] += 1
	}

	for key, count := range nodesMap {
		results = append(results, simulationResult{node: key, keys: count})
	}

	return results
}

func printResults(keys, nodes int, results []simulationResult) {
	var standardDeviation float64
	mean := float64(keys/nodes)

	for _, result := range results {
		standardDeviation += math.Pow(float64(result.keys)- mean, 2)
		fmt.Printf("% 8s: % 4d keys\n", result.node, result.keys)
	}

	standardDeviation = math.Sqrt(standardDeviation/float64(nodes))

	fmt.Printf("\nPerfect routing would give %.0f keys per node\n", mean)
	fmt.Printf("Standard deviation: ~%.2f\n", standardDeviation)
}

func main() {
	var keys int
	var nodes int

	flag.IntVar(&keys, "keys", 100, "Number of keys to route")
	flag.IntVar(&nodes, "nodes", 10, "Number of available nodes")

	flag.Parse()

	fmt.Printf("Simulating routing for %d keys with %d nodes\n", keys, nodes)

	results := simulate(keys, nodes)

	printResults(keys, nodes, results)
}