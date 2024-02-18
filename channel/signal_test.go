package channel

import "testing"

// cd /channel
// go test -run TestGracefulExit1
func TestGracefulExit1(t *testing.T) {
	GracefulExit1()
}

// cd /channel
// go test -run TestBetterGracefulExit
func TestBetterGracefulExit(t *testing.T) {
	BetterGracefulExit()
}
