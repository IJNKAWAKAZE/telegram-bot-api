package utils

import "fmt"

const (
	notFound = iota
	alreadyEnabled
	notSameTime
)

type Errors struct {
	errorType uint
	detail    any
}
type notSameTimeStruct struct {
	a string
	b string
}

func (e *Errors) Error() string {
	switch e.errorType {
	case notFound:
		return fmt.Sprintf("%s not found", e.detail.(string))
	case alreadyEnabled:
		return fmt.Sprintf("%s is already enabled", e.detail.(string))
	case notSameTime:
		detail := e.detail.(notSameTimeStruct)
		return fmt.Sprintf("%s and %s can not be enabled at same time", detail.a, detail.b)
	default:
		return fmt.Sprintf("unknown error: %d", e.errorType)
	}
}
func NotFoundError(detail string) *Errors {
	return &Errors{
		errorType: notFound,
		detail:    detail,
	}
}
func AlreadyEnabledError(detail string) *Errors {
	return &Errors{
		errorType: alreadyEnabled,
		detail:    detail,
	}
}
func NotSameTimeError(a string, b string) *Errors {
	return &Errors{
		errorType: notSameTime,
		detail:    notSameTimeStruct{a: a, b: b},
	}
}
