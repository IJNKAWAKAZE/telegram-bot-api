package optional

type Optional[T any] struct {
	value    T
	hasValue bool
}

func Some[T any](value T) Optional[T] {
	return Optional[T]{
		value:    value,
		hasValue: true,
	}
}
func None[T any]() Optional[T] {
	return Optional[T]{
		hasValue: false,
	}
}
func (t Optional[T]) IsSome() bool {
	return t.hasValue
}
func (t Optional[T]) IsNone() bool {
	return !t.hasValue
}
func (t Optional[T]) Unwarp() T {
	if !t.hasValue {
		panic("unwarp on None")
	} else {
		return t.value
	}
}
func (t Optional[T]) UnwarpDefault(d T) T {
	if !t.hasValue {
		return d
	} else {
		return t.value
	}
}
func (t *Optional[T]) Take() Optional[T] {
	if !t.hasValue {
		return None[T]()
	} else {
		result := Some(t.value)
		t.hasValue = false
		return result

	}
}
func OptionApply[T any, R any](op Optional[T], function func(T) R) Optional[R] {
	if op.IsSome() {
		return Some(function(op.Unwarp()))
	}
	return None[R]()
}
