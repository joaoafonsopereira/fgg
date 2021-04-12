package fg_test

import (
	"testing"
)

// Testing Impls (int literal implements int32)
func Test019(t *testing.T) {
	A := "type A struct { a int32 }"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "A{5}.id(10)"
	fgParseAndOkGood(t, A, Am, e)
}

// int literal implements Any
func Test019b(t *testing.T) {
	Any := "type Any interface {}"
	A := "type A struct {}"
	Am := "func (x0 A) m(i Any) Any { return i }"
	e := "A{}.m(5)"

	fgParseAndOkGood(t, Any, A, Am, e)
}

// struct doesn't implement int32
func Test019c(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "A{}.id(A{})"

	expectedPanic := "Arg expr type must implement param type: arg=A, param=int32" // !(A <: int32)
	fgParseAndOkBad(t, expectedPanic, A, Am, e)
}

// literal doesn't fit in an int8
func Test019d(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i int8) int8 { return i }"
	e := "A{}.id(257)"
	fgParseAndOkBad(t, A, Am, e)
}
