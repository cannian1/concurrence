package channel

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMutex(t *testing.T) {
	m := NewMutex()
	m.Lock()
	defer m.UnLock()
	assert.Equal(t, true, m.IsLocked())
	assert.Equal(t, false, m.TryLock())
	assert.Equal(t, false, m.LockTimeout(3*time.Second))
	m.UnLock()
	assert.Equal(t, true, m.LockTimeout(3*time.Second))
	m.UnLock()
	assert.Equal(t, true, m.TryLock())
}
