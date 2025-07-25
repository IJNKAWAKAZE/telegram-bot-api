package Result

import (
	"errors"
)

type Result[T any, R error] struct {
	t     T
	e     R
	isErr bool
}

func Ok[T any, R error](t T) Result[T, R] {
	return Result[T, R]{
		t:     t,
		isErr: false,
	}
}
func Err[T any, R error](r R) Result[T, R] {
	return Result[T, R]{
		e:     r,
		isErr: true,
	}
}
func (r Result[T, R]) IsOk() bool {
	return !r.IsErr()
}
func (r Result[T, R]) IsErr() bool {
	return r.isErr
}
func (r Result[T, R]) Unwarp() T {
	if !r.IsOk() {
		panic("Unwarp On error")
	}
	return r.t
}
func (r Result[T, R]) UnwarpErr() R {
	if r.IsOk() {
		panic("UnwarpErr On Ok")
	}
	return r.e
}
func (r Result[T, R]) UnwarpToGo() (T, R) {
	return r.t, r.e
}
func GOk[T any](t T) Result[T, error] {
	return Ok[T, error](t)
}

func GErr[T any](errorMsg string) Result[T, error] {
	return Err[T, error](errors.New(errorMsg))
}
