package hashring

import (
	"crypto/sha1"
	"sort"
	"strconv"
	"sync"
)

type HashRing struct {
	virtualReplicas int
	nodes           map[string]struct{}
	ring            []uint32
	nodeMap         map[uint32]string
	lock            sync.RWMutex
}

func New(virtualReplicas int) *HashRing {
	return &HashRing{
		virtualReplicas: virtualReplicas,
		nodes:           make(map[string]struct{}),
		nodeMap:         make(map[uint32]string),
	}
}

func (h *HashRing) AddNode(node string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, exists := h.nodes[node]; exists {
		return
	}
	h.nodes[node] = struct{}{}
	for i := 0; i < h.virtualReplicas; i++ {
		vNode := node + "#" + strconv.Itoa(i)
		hash := hashKey(vNode)
		h.ring = append(h.ring, hash)
		h.nodeMap[hash] = node
	}
	sort.Slice(h.ring, func(i, j int) bool { return h.ring[i] < h.ring[j] })
}

func (h *HashRing) RemoveNode(node string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, exists := h.nodes[node]; !exists {
		return
	}
	delete(h.nodes, node)
	newRing := make([]uint32, 0, len(h.ring))
	for i := 0; i < h.virtualReplicas; i++ {
		vNode := node + "#" + strconv.Itoa(i)
		hash := hashKey(vNode)
		delete(h.nodeMap, hash)
	}
	for _, hash := range h.ring {
		if h.nodeMap[hash] != "" {
			newRing = append(newRing, hash)
		}
	}
	h.ring = newRing
}

func (h *HashRing) GetNode(key string) string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if len(h.ring) == 0 {
		return ""
	}
	hash := hashKey(key)
	idx := sort.Search(len(h.ring), func(i int) bool { return h.ring[i] >= hash })
	if idx == len(h.ring) {
		idx = 0
	}
	return h.nodeMap[h.ring[idx]]
}

func (h *HashRing) Nodes() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	result := make([]string, 0, len(h.nodes))
	for n := range h.nodes {
		result = append(result, n)
	}
	return result
}

func hashKey(key string) uint32 {
	h := sha1.New()
	h.Write([]byte(key))
	sum := h.Sum(nil)
	return (uint32(sum[16]) << 24) | (uint32(sum[13]) << 16) | (uint32(sum[7]) << 8) | uint32(sum[3])
}
