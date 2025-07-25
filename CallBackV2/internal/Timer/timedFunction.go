package Timer

import (
	"github.com/google/uuid"
	"time"
)

type timedCloser struct {
	closer     func() error
	tiggerTime time.Time
	closerUUID uuid.UUID
}

func (c timedCloser) cancel() {
	c.closer = nil
}
func (c timedCloser) exec() error {
	if c.closer != nil {
		return c.closer()
	}
	return nil
}

type rawTimedCloser struct {
	closer     func() error
	tiggerTime time.Time
}

func newRawTimedCloser(closer func() error, tiggerTime time.Time) rawTimedCloser {
	return rawTimedCloser{
		closer:     closer,
		tiggerTime: tiggerTime,
	}
}
func fromRawTimedCloser(rawcloser rawTimedCloser, uuid2 uuid.UUID) timedCloser {
	return timedCloser{
		closer:     rawcloser.closer,
		tiggerTime: rawcloser.tiggerTime,
		closerUUID: uuid2,
	}
}
