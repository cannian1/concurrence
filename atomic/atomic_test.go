package atomic

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestAddXXX(t *testing.T) {
	// 对于有符号整数，第二个参数传负数就实现了减法
	// 而对于无符号整数，需要使用补码变减法为加法
	var x uint64 = 0

	newXValue := atomic.AddUint64(&x, 100) // newXValue == 100
	assert.Equal(t, uint64(100), newXValue)

	// 利用补码变减法为加法
	newXValue = atomic.AddUint64(&x, ^uint64(0)) // newXValue == 99
	assert.Equal(t, uint64(99), newXValue)

	newXValue = atomic.AddUint64(&x, ^uint64(10-1)) // newXValue == 89
	assert.Equal(t, uint64(89), newXValue)

	var y atomic.Int64
	y.Store(100) // y.v == 100
	assert.Equal(t, int64(100), y.Load())

	y.Add(100) // y.v == 200
	assert.Equal(t, int64(200), y.Load())
}

func TestCasXXX(t *testing.T) {
	// 比较新值是否与旧值相等，不相等返回 false
	// 相等则将新值与旧值交换，返回 true
	var x atomic.Uint64
	ok1 := x.CompareAndSwap(0, 100)
	assert.Equal(t, true, ok1)

	var y uint64 = 0
	ok2 := atomic.CompareAndSwapUint64(&y, 0, 100)
	assert.Equal(t, true, ok2)
}

func TestSwapXXX(t *testing.T) {
	// 不需要比较旧值，直接交换
	var x uint64
	old := atomic.SwapUint64(&x, 123) // old == 0, x == 123
	assert.Equal(t, uint64(0), old)
	assert.Equal(t, uint64(123), x)

	var y atomic.Uint64
	y.Swap(321)
	assert.Equal(t, uint64(321), y.Load())
}

func TestLoadXXX(t *testing.T) {
	// 取出 addr 地址中的值，原子操作
	var x uint64
	v := atomic.LoadUint64(&x)
	assert.Equal(t, uint64(0), v)
	x = 100
	v = atomic.LoadUint64(&x)
	assert.Equal(t, uint64(100), v)
}
