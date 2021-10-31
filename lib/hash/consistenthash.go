package hash

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/mapping"
)

const (
	// TopWeight 可设置的最大权重值。
	TopWeight = 100

	minReplicas = 100
	prime       = 16777619
)

type (
	// Func 定义哈希计算方法。
	Func func(data []byte) uint64

	// ConsistentHash 是一个环形哈希实现。
	//
	// 	-> ring: {hash1: [node1, node2]}
	//	-> keys: [hash1, hash2]
	//	-> nodes: {'': {}}
	ConsistentHash struct {
		hashFunc Func
		replicas int
		keys     []uint64                 // 以hash值为key, Sorted hash list
		ring     map[uint64][]interface{} // 以hash值为键，node列表为值的map
		nodes    map[string]lang.PlaceholderType
		lock     sync.RWMutex
	}
)

// NewConsistentHash 返回一个一致性哈希。
func NewConsistentHash() *ConsistentHash {
	return NewCustomConsistentHash(minReplicas, Hash)
}

// NewCustomConsistentHash 返回一个具有指定副本和哈希函数的一致性哈希。
func NewCustomConsistentHash(replicas int, hashFunc Func) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	if hashFunc == nil {
		hashFunc = Hash
	}

	return &ConsistentHash{
		hashFunc: hashFunc,
		replicas: replicas,
		ring:     make(map[uint64][]interface{}),
		nodes:    make(map[string]lang.PlaceholderType),
	}
}

// Add 添加指定副本集数量的节点，后续调用将覆盖之前的副本。
func (h *ConsistentHash) Add(node interface{}) {
	h.AddWithReplicas(node, h.replicas)
}

// AddWithWeight 添加带权重的节点，权重值为1-100，代表百分比。后续调用将覆盖之前的副本。
func (h *ConsistentHash) AddWithWeight(node interface{}, weight int) {
	replicas := h.replicas * weight / TopWeight
	h.AddWithReplicas(node, replicas)
}

// AddWithReplicas 添加指定副本集数量的节点。
// 如果副本数量比 h.replicas 多将被截断，后添加会覆盖之前添加的。
func (h *ConsistentHash) AddWithReplicas(node interface{}, replicas int) {
	h.Remove(node)

	if replicas > h.replicas {
		replicas = h.replicas
	}

	nodeRepr := repr(node)
	h.lock.Lock()
	defer h.lock.Unlock()

	h.addNode(nodeRepr)

	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.ring[hash] = append(h.ring[hash], node)
	}

	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

// Get 获取 h 中指定 v 对应的 node 节点。
func (h *ConsistentHash) Get(v interface{}) (interface{}, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.ring) == 0 {
		return nil, false
	}

	hash := h.hashFunc([]byte(repr(v)))
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	}) % len(h.keys)

	nodes := h.ring[h.keys[index]]
	switch len(nodes) {
	case 0:
		return nil, false
	case 1:
		return nodes[0], true
	default:
		innerIndex := h.hashFunc([]byte(innerRepr(v)))
		pos := int(innerIndex % uint64(len(nodes)))
		return nodes[pos], true
	}
}

// Remove 从 h 中移除指定节点。
func (h *ConsistentHash) Remove(node interface{}) {
	nodeRepr := repr(node)

	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.containsNode(nodeRepr) {
		return
	}

	for i := 0; i < h.replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i)))
		index := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hash
		})
		if index < len(h.keys) && h.keys[index] == hash {
			h.keys = append(h.keys[:index], h.keys[index+1:]...)
		}
		h.removeRingNode(hash, nodeRepr)
	}

	h.removeNode(nodeRepr)
}

// 删除 ring 中 该 hash 对应的 node
func (h *ConsistentHash) removeRingNode(hash uint64, nodeRepr string) {
	if nodes, ok := h.ring[hash]; ok {
		newNodes := nodes[:0]
		for _, x := range nodes {
			if repr(x) != nodeRepr {
				newNodes = append(newNodes, x)
			}
		}
		if len(newNodes) > 0 {
			h.ring[hash] = newNodes
		} else {
			// 若是节点列表中的最后一个节点，则删除该映射对
			delete(h.ring, hash)
		}
	}
}

func (h *ConsistentHash) removeNode(nodeRepr string) {
	delete(h.nodes, nodeRepr)
}

func (h *ConsistentHash) addNode(nodeRepr string) {
	h.nodes[nodeRepr] = lang.Placeholder
}

func (h *ConsistentHash) containsNode(nodeRepr string) bool {
	_, ok := h.nodes[nodeRepr]
	return ok
}

func repr(node interface{}) string {
	return mapping.Repr(node)
}

func innerRepr(node interface{}) string {
	return fmt.Sprintf("%d:%v", prime, node)
}
