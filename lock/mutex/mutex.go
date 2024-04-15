package mutex

import "sync"

type MyConcurrentMap struct {
	mp map[int]int
	sync.Mutex
}

func NewMyConcurrentMap() *MyConcurrentMap {
	return &MyConcurrentMap{mp: make(map[int]int)}
}

func (m *MyConcurrentMap) Set(k, v int) {
	m.Lock()
	defer m.Unlock()
	m.mp[k] = v
}

func (m *MyConcurrentMap) Get(k int) (int, bool) {
	// 加锁与解锁之间的代码是临界区，在被锁包住的代码块之外是可以并发访问的
	m.Mutex.Lock()
	v, ok := m.mp[k]
	m.Mutex.Unlock()

	return v, ok
}

func (m *MyConcurrentMap) Delete(k int) {
	m.Lock()
	defer m.Unlock()
	delete(m.mp, k)
}
