package utils

func MapContain[T comparable, V any](tMap map[T]V, tKey T) bool {
	_, ok := tMap[tKey]
	return ok
}
