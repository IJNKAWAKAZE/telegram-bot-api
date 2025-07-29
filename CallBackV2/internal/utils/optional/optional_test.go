package optional

import "testing"

func timeTwo(i int) int {
	return i * 2
}
func isOdd(i int) bool {
	return i%2 != 0
}
func TestAll(t *testing.T) {
	assert := func(b bool) {
		if !b {
			t.FailNow()
		}
	}
	assert(Some(1).IsSome())
	assert(!None[any]().IsSome())
	assert(Some(123).Unwarp() == 123)
	assert(OptionApply(Some(123), timeTwo).Unwarp() == 246)
	assert(OptionApply(Some(123), isOdd).Unwarp())
	assert(OptionApply(None[int](), isOdd).hasValue == false)

}
