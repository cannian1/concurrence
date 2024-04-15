package rwmutex

import "sync"

type MyConcurrentMap struct {
	mp map[int]int
	sync.RWMutex
}

// 读和写是互斥的，写和写也是互斥的，读和读可以并发

func NewMyConcurrentMap() *MyConcurrentMap {
	return &MyConcurrentMap{mp: make(map[int]int)}
}

func (m *MyConcurrentMap) Set(k, v int) {
	m.Lock()
	defer m.Unlock()
	m.mp[k] = v
}

func (m *MyConcurrentMap) Get(k int) (int, bool) {
	m.RWMutex.RLock()
	v, ok := m.mp[k]
	m.RWMutex.RUnlock()

	return v, ok
}

func (m *MyConcurrentMap) Delete(k int) {
	m.Lock()
	defer m.Unlock()
	delete(m.mp, k)
}
