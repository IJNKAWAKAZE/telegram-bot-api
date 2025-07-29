package Timer

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/google/uuid"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils/optional"
)

type timeUUIDTree struct {
	tree       *rbt.Tree
	uuidMapper map[uuid.UUID]timedCloser
}

func timeCloserComparator(a, b any) int {
	aAsserted := a.(timedCloser)
	bAsserted := b.(timedCloser)
	result := utils.TimeComparator(aAsserted.tiggerTime, bAsserted.tiggerTime)
	if result == 0 {
		return utils.StringComparator(aAsserted.closerUUID.String(), bAsserted.closerUUID.String())
	}
	return result
}
func initTree() timeUUIDTree {
	return timeUUIDTree{
		tree:       rbt.NewWith(timeCloserComparator),
		uuidMapper: make(map[uuid.UUID]timedCloser),
	}

}
func (t timeUUIDTree) add(newCloser timedCloser) {
	t.tree.Put(newCloser, newCloser)
	t.uuidMapper[newCloser.closerUUID] = newCloser
}
func (t timeUUIDTree) remove(newUUID uuid.UUID) {
	targetCloser := t.uuidMapper[newUUID]
	t.tree.Remove(targetCloser)
	delete(t.uuidMapper, newUUID)
}
func (t timeUUIDTree) find(uuid uuid.UUID) optional.Optional[timedCloser] {
	obj, ok := t.uuidMapper[uuid]
	if !ok {
		return optional.Optional[timedCloser]{}
	} else {
		return optional.Some(obj)
	}
}
func (t timeUUIDTree) PopMin() optional.Optional[timedCloser] {
	closer := t.FindMin()
	if closer.IsSome() {
		closer := closer.Unwarp()
		t.remove(closer.closerUUID)
		return optional.Some(closer)
	}
	return optional.None[timedCloser]()
}
func (t timeUUIDTree) FindMin() optional.Optional[timedCloser] {
	minNode := t.tree.Root
	if minNode == nil {
		return optional.None[timedCloser]()
	}
	for minNode.Left != nil {
		minNode = minNode.Left
	}
	closer := minNode.Key.(timedCloser)
	return optional.Some(closer)
}
func (t timeUUIDTree) PopByUUID(targetUUID uuid.UUID) timedCloser {
	closer := t.uuidMapper[targetUUID]
	t.remove(closer.closerUUID)
	return closer
}
