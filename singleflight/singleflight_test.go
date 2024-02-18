package singleflight

import "testing"

func TestDemo(t *testing.T) {
	Demo()
}

func ExampleDemo() {
	Demo()
	// output:
	// get data
	// one call v: data_1, shared: true
	// two call v: data_1, shared: true
}

func TestDemo2(t *testing.T) {
	Demo2()
}

func ExampleDemo2() {
	Demo2()
	// output:
	// get data
	// one call v: , shared: false err: context deadline exceeded
	// two call v: , shared: false err: context deadline exceeded
}

func TestDemo3(t *testing.T) {
	Demo3()
}

func ExampleDemo3() {
	Demo3()
	// output:
	// get data
	// one call v: , shared: false err: context deadline exceeded
	// two call v: data_1, shared: true err: <nil>
}
