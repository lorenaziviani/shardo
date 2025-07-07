package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"shardo/pkg/hashring"
)

func main() {
	nodesStr := flag.String("nodes", "", "Comma-separated list of node names")
	keys := flag.Int("keys", 1000, "Number of keys to distribute")
	addNode := flag.String("add", "", "Add a node and show redistribution")
	removeNode := flag.String("remove", "", "Remove a node and show redistribution")
	virtualReplicas := flag.Int("replicas", 100, "Number of virtual replicas per node")
	flag.Parse()

	if *nodesStr == "" {
		fmt.Println("Usage: hashring-cli --nodes node1,node2 --keys 1000 [--add node3] [--remove node2] [--replicas 100]")
		os.Exit(1)
	}
	nodes := strings.Split(*nodesStr, ",")
	ring := hashring.New(*virtualReplicas)
	for _, n := range nodes {
		ring.AddNode(n)
	}

	if *addNode != "" {
		ring.AddNode(*addNode)
		nodes = append(nodes, *addNode)
		fmt.Printf("Added node: %s\n", *addNode)
	}
	if *removeNode != "" {
		ring.RemoveNode(*removeNode)
		fmt.Printf("Removed node: %s\n", *removeNode)
	}

	distribution := make(map[string]int)
	for _, n := range ring.Nodes() {
		distribution[n] = 0
	}
	for i := 0; i < *keys; i++ {
		key := fmt.Sprintf("key%d", i)
		node := ring.GetNode(key)
		distribution[node]++
	}

	fmt.Println("Key distribution:")
	for _, n := range ring.Nodes() {
		fmt.Printf("%s: %d\n", n, distribution[n])
	}
}
