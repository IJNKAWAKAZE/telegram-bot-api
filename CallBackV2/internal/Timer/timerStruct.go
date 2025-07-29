package Timer

import (
	"errors"
	"github.com/google/uuid"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils/Result"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils/optional"
	"time"
)

type timer struct {
	currentMin optional.Optional[timedCloser]
	tree       timeUUIDTree
	holdUUIDs  utils.Set[uuid.UUID]
}

type operation = struct {
	uint8
	any
}

const (
	getAndHoldUUID = iota // use with nil (second filed omitted) return an uuid
	add                   // use with rawTimedCloser return uuid,error
	addWithUUID           // use with timedCloser return bool,error
	remove                // use with uuid:UUID
	find                  // use with uuid:UUID read result from fromTimer
	stop                  // use with nil (second filed omitted)
)
const FOREVER time.Duration = 1<<63 - 1 // Forever

func timerLoop(operations chan operation, fromTimer chan any, logFunc func(err error)) {
	t := timer{
		tree:       initTree(),
		currentMin: optional.None[timedCloser](),
		holdUUIDs:  utils.NewSet[uuid.UUID](),
	}
	var waitDuration = FOREVER
	for {
		select {
		case op := <-operations:
			switch op.uint8 {
			case add:
				rawCloser, ok := op.any.(rawTimedCloser)
				if !ok {
					fromTimer <- Result.GErr[uuid.UUID]("invalid operation should be type rawTimedCloser")
				} else {
					fromTimer <- Result.GOk(t.add(rawCloser))
				}
			case getAndHoldUUID:
				fromTimer <- t.getAndHoldUUID()
			case addWithUUID:
				closer, ok := op.any.(timedCloser)
				if !ok {
					fromTimer <- Result.GErr[bool]("invalid operation should be type timedCloser")
				} else {
					err := t.addWithHoldedUUID(closer)
					if err != nil {
						fromTimer <- Result.GErr[bool](err.Error())
					} else {
						fromTimer <- Result.GOk(true)
					}
				}
			case remove:
				targetUUID, ok := op.any.(uuid.UUID)
				if !ok {
					fromTimer <- Result.GErr[bool]("invalid operation should be type uuid")
				} else {
					fromTimer <- Result.GOk(t.remove(targetUUID))
				}
			case find:
				targetUUID, ok := op.any.(uuid.UUID)
				if !ok {
					fromTimer <- Result.GErr[bool]("invalid operation should be type uuid")
				} else {
					fromTimer <- Result.GOk(t.contains(targetUUID))
				}
			case stop:
				return
			}
		case <-time.After(waitDuration):
			err := t.exec()
			if err != nil {
				logFunc(err)
			}

		}
		waitDuration = t.getDuration()
	}

}
func (timer *timer) add(rawCloser rawTimedCloser) uuid.UUID {
	closerUUID := timer.getRandomAvaliableUUID()
	timer.rawAdd(fromRawTimedCloser(rawCloser, closerUUID))
	return closerUUID
}
func (timer *timer) rawAdd(newCloser timedCloser) {
	if timer.currentMin.IsNone() {
		timer.currentMin = optional.Some(newCloser)
		return
	}
	smallestCloser := timer.currentMin.Unwarp()
	if smallestCloser.tiggerTime.After(newCloser.tiggerTime) {
		timer.currentMin = optional.Some(newCloser)
		timer.tree.add(smallestCloser)
	} else {
		timer.tree.add(newCloser)
	}

}
func (timer *timer) getAndHoldUUID() uuid.UUID {
	result := timer.getRandomAvaliableUUID()
	timer.holdUUIDs.Add(result)
	return result
}
func (timer *timer) addWithHoldedUUID(closer timedCloser) error {
	closerUUID := closer.closerUUID
	if timer.holdUUIDs.Contains(closerUUID) {
		timer.holdUUIDs.Remove(closerUUID)
		timer.rawAdd(closer)
		return nil
	}
	return errors.New("closer UUID not exists or already used")
}
func (timer timer) getRandomAvaliableUUID() uuid.UUID {
	targetUUID := uuid.New()
	for timer.contains(targetUUID) {
		targetUUID = uuid.New()
	}
	return targetUUID
}
func (timer *timer) remove(targetUUID uuid.UUID) bool {
	if !timer.contains(targetUUID) {
		return false
	}
	if timer.holdUUIDs.Contains(targetUUID) {
		timer.holdUUIDs.Remove(targetUUID)
	} else if timer.currentMin.Unwarp().closerUUID == targetUUID {
		timer.getNewMin()

	} else {
		timer.tree.remove(targetUUID)
	}
	return true
}
func (timer timer) getDuration() time.Duration {
	if timer.currentMin.IsNone() {
		return FOREVER
	} else {
		return timer.currentMin.Unwarp().tiggerTime.Sub(time.Now())
	}
}
func (timer *timer) getNewMin() {
	timer.currentMin = timer.tree.PopMin()
}
func (timer timer) contains(targetUUID uuid.UUID) bool {
	if timer.holdUUIDs.Contains(targetUUID) {
		return true
	}
	if timer.currentMin.IsNone() {
		return false
	}
	if timer.currentMin.Unwarp().closerUUID == targetUUID {
		return true
	}
	return timer.tree.find(targetUUID).IsSome()
}
func (timer *timer) exec() error {
	if timer.currentMin.IsNone() {
		return nil
	} else {
		oldMin := timer.currentMin.Take().Unwarp()
		timer.getNewMin()
		return oldMin.exec()
	}
}
