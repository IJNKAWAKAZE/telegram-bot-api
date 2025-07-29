package utils

type Set[T comparable] struct {
	m map[T]any
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{m: make(map[T]any)}
}
func (s Set[T]) Add(items ...T) bool {
	result := true
	for _, item := range items {
		if !s.Contains(item) {
			s.m[item] = nil
		} else {
			result = false
		}
	}
	return result
}
func (s Set[T]) Contains(item T) bool {
	_, ok := s.m[item]
	return ok
}
func (s Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s.m, item)
	}
}
