package collection

import "sync"

// Ring 用于充当一个固定大小的环装容器。
type Ring struct {
	elements []any
	index    int
	lock     sync.Mutex
}

// NewRing 返回一个给定大小 n 的 Ring。
func NewRing(n int) *Ring {
	if n < 1 {
		panic("n 必须大于 0")
	}

	return &Ring{
		elements: make([]any, n),
	}
}

// Add 添加 v 至 r。
func (r *Ring) Add(v any) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.elements[r.index%len(r.elements)] = v
	r.index++
}

// Take 从 r 中取出所有元素。
func (r *Ring) Take() []any {
	r.lock.Lock()
	defer r.lock.Unlock()

	var size int
	var start int
	if r.index > len(r.elements) {
		size = len(r.elements)
		start = r.index % len(r.elements)
	} else {
		size = r.index
	}

	elements := make([]any, size)
	for i := 0; i < size; i++ {
		elements[i] = r.elements[(start+i)%len(r.elements)]
	}

	return elements

}
