package fg_test

import (
	"github.com/rhu1/fgg/internal/base/testutils"
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

// literal doesn't fit in an int32 (and hence doesn't "implement" int32)
func Test019d(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "A{}.id(2147483648)" // 1 << 31 (math.MaxInt32 + 1)
	fgParseAndOkBad(t, A, Am, e)
}

// can't mix different types
func Test020(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) add(x float32, y float64) float64 { return x+y }"
	e := "A{}"
	fgParseAndOkBad(t, A, Am, e)
	//prog := fgParseAndOkGood(t, A, Am, e)
	//testutils.EvalAndOkGood(t, prog, 2)
}

// edge case - number of the form 'x.0' can either be int or float
func Test021(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "A{}.id(1 + 41.0)"
	fgParseAndOkGood(t, A, Am, e)
}

func Test021b(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i float32) float32 { return i }"
	e := "A{}.id(1 + 41.0)"
	fgParseAndOkGood(t, A, Am, e)
}

func Test021c(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "A{}.id(1 + 41.1)"
	fgParseAndOkBad(t, A, Am, e)
}

// Comparisons and logical ops
func Test022(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) cmp(x float64, y float64) bool { return x > y || (x < x && false) }"
	e := "A{}.cmp(2, 4.2)"
	//fgParseAndOkGood(t, A, Am, e)
	prog := fgParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)
}

// test overflow
func Test023(t *testing.T) {
	A := "type A struct {}"
	Am := "func (x0 A) add(x int32, y int32) int32 { return x+y }"
	e := "A{}.add(2147483647, 1)"
	prog := fgParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 4)
}

func Test023b(t *testing.T) {
	A := "type A struct {x int32}"
	Am := "func (x0 A) add1() int32 { return x0.x + 1 }"
	e := "A{2147483647}.add1()"
	prog := fgParseAndOkGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 5)
}

/******************************************************************************/
/* Tests that show that some form of coercion is necessary */

// Passing a struct literal to a method expecting
// a TNamed-struct type - which defines methods
func TestStructLit(t *testing.T) {
	S := "type S struct {}"
	Sm := "func (x S) callId(a A) int32 { return a.id(1) }"
	A := "type A struct { a int32 }"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "S{}.callId( struct{a int32}{5} )"
	prog := fgParseAndOkGood(t, S, Sm, A, Am, e)

	testutils.EvalAndOkGood(t, prog, 3) // Unexpected panic: Method not found: id in  struct { a int32 }
}

// Returning a struct literal from a method that "promises"
// a TNamed-struct type - which defines methods
func TestStructLit2(t *testing.T) {
	S := "type S struct {}"
	Sm := "func (x S) retA() A { return struct{a int32}{5} }"
	A := "type A struct { a int32 }"
	Am := "func (x0 A) id(i int32) int32 { return i }"
	e := "S{}.retA().id(1)"
	prog := fgParseAndOkGood(t, S, Sm, A, Am, e)

	testutils.EvalAndOkGood(t, prog, 3) //Unexpected panic: Method not found: id in  struct { a int32 }
}

// Returning an integer literal from a method that "promises"
// a MyInt - which defines methods
func TestIntLit(t *testing.T) {
	S := "type S struct {}"
	Sm := "func (x S) retMyInt() MyInt { return 5 }"
	A := "type MyInt int32"
	Am := "func (x0 MyInt) id(i int32) int32 { return i }"
	//e := "struct{x int32}{5}.x"
	e := "S{}.retMyInt().id(1)"
	prog := fgParseAndOkGood(t, S, Sm, A, Am, e)

	testutils.EvalAndOkGood(t, prog, 3) //Unexpected panic: Method not found: id in int32(undefined)
}

// Testing conversions in "constant expressions"
func TestConvs(t *testing.T) {
	S := "type S struct {}"
	Sm := "func (this S) foo(x float64) float64 { return x }"
	e := "S{}.foo(100 + 25.52)"
	prog := fgParseAndOkGood(t, S, Sm, e)
	testutils.EvalAndOkGood(t, prog, 4)
}