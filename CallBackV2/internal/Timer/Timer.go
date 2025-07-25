package Timer

import (
	"github.com/google/uuid"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils/Result"
	"log"
	"sync"
	"time"
)

var defaultLogFunction = func(err error) { log.Printf("error %s", err.Error()) }

// const (
//
//	getAndHoldUUID = iota // use with nil (second filed omitted) return an uuid
//	add                   // use with rawTimedCloser return uuid,error
//	addWithUUID           // use with timedCloser return bool,error
//	remove                // use with uuid:UUID
//	find                  // use with uuid:UUID read result from fromTimer
//	stop                  // use with nil (second filed omitted)
//
// )
type Timer struct {
	toTimer   chan operation
	fromTimer <-chan any
	mutex     sync.Mutex
	active    bool
}

func NewTimer() Timer {
	return NewTimerWithLogFunc(defaultLogFunction)
}
func NewTimerWithLogFunc(logFunction func(err error)) Timer {
	toTimer := make(chan operation)
	fromTimer := make(chan any)
	go timerLoop(toTimer, fromTimer, logFunction)
	return Timer{
		fromTimer: fromTimer,
		toTimer:   toTimer,
		mutex:     sync.Mutex{},
		active:    true,
	}
}
func (t *Timer) TimerStop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.toTimer <- operation{stop, nil}
	t.active = false
}
func (t *Timer) TimerAdd(rawCloser func() error, after time.Duration) uuid.UUID {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.active {
		panic("timer not start")
	}
	closer := newRawTimedCloser(rawCloser, time.Now().Add(after))
	t.toTimer <- operation{add, closer}
	closerUUID := <-t.fromTimer
	return closerUUID.(Result.Result[uuid.UUID, error]).Unwarp()
}
func (t *Timer) TimerAddWithUUID(rawCloser func() error, after time.Duration, uuid2 uuid.UUID) Result.Result[bool, error] {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.active {
		panic("timer not start")
	}
	closer := timedCloser{
		closer:     rawCloser,
		tiggerTime: time.Now().Add(after),
		closerUUID: uuid2,
	}
	t.toTimer <- operation{addWithUUID, closer}
	result := <-t.fromTimer
	return result.(Result.Result[bool, error])

}
func (t *Timer) TimerRemove(targetUUID uuid.UUID) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.active {
		panic("timer not start")
	}
	t.toTimer <- operation{remove, targetUUID}
	boolResult := <-t.fromTimer
	return boolResult.(Result.Result[bool, error]).Unwarp()

}
func (t *Timer) TimerGetAvailableUUID() uuid.UUID {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.active {
		panic("timer not start")
	}
	t.toTimer <- operation{getAndHoldUUID, nil}
	result := <-t.fromTimer
	return result.(uuid.UUID)
}
