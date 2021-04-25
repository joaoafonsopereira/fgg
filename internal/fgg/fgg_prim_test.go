package fgg_test

import (
	"github.com/rhu1/fgg/internal/base/testutils"
	"testing"
)

// For now, all these tests are just the adaptation of the fg tests in fg_primitives_test

func Test019(t *testing.T) {
	A := "type A(type ) struct { a int32 }"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){5}.id()(10)"
	fggParseAndOkGood(t, A, Am, e)
}

// int literal implements Any
func Test019b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) m(type )(i Any()) Any() { return i }"
	e := "A(){}.m()(5)"

	fggParseAndOkGood(t, Any, A, Am, e)
}

// struct doesn't implement int32
func Test019c(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){}.id()(A(){})"

	expectedPanic := "Arg expr type must implement param type: arg=A, param=int32" // !(A <: int32)
	fggParseAndOkBad(t, expectedPanic, A, Am, e)
}

// literal doesn't fit in an int32 (and hence doesn't "implement" int32)
func Test019d(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){}.id()(2147483648)" // 1 << 31 (math.MaxInt32 + 1)
	fggParseAndOkBad(t, A, Am, e)
}

// can't mix different types
func Test020(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) add(type )(x float32, y float64) float64 { return x+y }"
	e := "A(){}"
	fggParseAndOkBad(t, A, Am, e)
	//prog := fgParseAndOkGood(t, A, Am, e)
	//testutils.EvalAndOkGood(t, prog, 2)
}

// edge case - number of the form 'x.0' can either be int or float
func Test021(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){}.id()(1 + 41.0)"
	fggParseAndOkGood(t, A, Am, e)
}

func Test021b(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i float32) float32 { return i }"
	e := "A(){}.id()(1 + 41.0)"
	fggParseAndOkGood(t, A, Am, e)
}

func Test021c(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){}.id()(1 + 41.1)"
	fggParseAndOkBad(t, A, Am, e)
}



// TODO conversions NumericLiteral -> ...Val


// Comparisons and logical ops
func Test022(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) cmp(type )(x float64, y float64) bool { return x > y || (x < x && false) }"
	e := "A(){}.cmp()(2, 4.2)"
	//fgParseAndOkGood(t, A, Am, e)
	prog := fggParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)
}

func Test023(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) add(type )(x int32, y int32) int32 { return x+y }"
	e := "A(){}.add()(2147483647, 1)"
	//fgParseAndOkGood(t, A, Am, e)
	prog := fggParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)
}

func Test023b(t *testing.T) {
	A := "type A(type ) struct {x int32}"
	Am := "func (x0 A(type )) add1(type )() int32 { return x0.x + 1 }"
	e := "A(){2147483647}.add1()()"
	//fgParseAndOkGood(t, A, Am, e)
	prog := fggParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 4)
}
