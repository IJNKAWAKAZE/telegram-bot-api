package Timer

import (
	"testing"
	"time"
)

func TestTimerAdd(t *testing.T) {
	var k = 1
	ti := NewTimer()
	ti.TimerAdd(func() error {
		k = k + 1
		return nil
	}, 2*time.Second)
	if k != 1 {
		t.FailNow()
	}
	time.Sleep(3 * time.Second)
	if k != 2 {
		t.FailNow()
	}
}
func TestTimerAddMulti(t *testing.T) {
	var k = 1
	ti := NewTimer()
	ti.TimerAdd(func() error {
		k = k + 1
		return nil
	}, 2*time.Second)
	ti.TimerAdd(func() error {
		k = k + 1
		return nil
	}, 3*time.Second)
	if k != 1 {
		t.FailNow()
	}
	time.Sleep(4 * time.Second)
	if k != 3 {
		t.FailNow()
	}
}

func TestTimerRemove(t *testing.T) {
	var k = 1
	ti := NewTimer()
	removeThat := ti.TimerAdd(func() error {
		k = 2
		return nil
	}, 2*time.Second)
	time.Sleep(1 * time.Second)
	if !ti.TimerRemove(removeThat) {
		t.FailNow()
	}
	time.Sleep(3 * time.Second)
	if k != 1 {
		t.FailNow()
	}
}
func TestTimerHoldAdd(t *testing.T) {
	var k = 1
	ti := NewTimer()

	closerUUID := ti.TimerGetAvailableUUID()
	b := ti.TimerAddWithUUID(func() error {
		k = 2
		return nil
	}, 2*time.Second, closerUUID)
	if b.IsErr() {
		t.FailNow()
	}
	b = ti.TimerAddWithUUID(func() error {
		k = 3
		return nil
	}, 2*time.Second, closerUUID)
	if b.IsOk() {
		t.FailNow()
	}
	time.Sleep(3 * time.Second)
	if k != 2 {
		t.FailNow()
	}
}
func TestTimerRemoveHold(t *testing.T) {
	ti := NewTimer()

	closerUUID := ti.TimerGetAvailableUUID()
	if !ti.TimerRemove(closerUUID) {
		t.FailNow()
	}
}
