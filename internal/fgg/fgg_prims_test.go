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
	//fggParseAndOkGood(t, A, Am, e)
	fggParseAndOkMonomGood(t, A, Am, e)
}

// int literal implements Any
func Test019b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) m(type )(i Any()) Any() { return i }"
	e := "A(){}.m()(5)"

	//fggParseAndOkGood(t, Any, A, Am, e)
	fggParseAndOkMonomGood(t, Any, A, Am, e)
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
	//fggParseAndOkGood(t, A, Am, e)
	fggParseAndOkMonomGood(t, A, Am, e)
}

func Test021b(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i float32) float32 { return i }"
	e := "A(){}.id()(1 + 41.0)"
	//fggParseAndOkGood(t, A, Am, e)
	fggParseAndOkMonomGood(t, A, Am, e)
}

func Test021c(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type )(i int32) int32 { return i }"
	e := "A(){}.id()(1 + 41.1)"
	fggParseAndOkBad(t, A, Am, e)
}

// Comparisons and logical ops
func Test022(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) cmp(type )(x float64, y float64) bool { return x > y || (x < x && false) }"
	e := "A(){}.cmp()(2, 4.2)"
	//fgParseAndOkGood(t, A, Am, e)
	//prog := fggParseAndOkGood(t, A, Am, e)
	prog := fggParseAndOkMonomGood(t, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)
}

func Test023(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) add(type )(x int32, y int32) int32 { return x+y }"
	e := "A(){}.add()(2147483647, 1)"
	//fggParseAndOkGood(t, A, Am, e)
	fggParseAndOkMonomGood(t, A, Am, e)
	//prog := fggParseAndOkGood(t, A, Am, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

func Test023b(t *testing.T) {
	A := "type A(type ) struct {x int32}"
	Am := "func (x0 A(type )) add1(type )() int32 { return x0.x + 1 }"
	e := "A(){2147483647}.add1()()"
	//fggParseAndOkGood(t, A, Am, e)
	fggParseAndOkMonomGood(t, A, Am, e)
	//prog := fggParseAndOkGood(t, A, Am, e)
	//testutils.EvalAndOkGood(t, prog, 4)
}

func Test023c(t *testing.T) {
	A := "type A(type ) struct {}"
	e := "\"a\" + \"b\""
	//prog := fggParseAndOkGood(t, A, e)

	//prog := fggParseAndOkMonomGood(t, A, e)
	prog := fggParseAndOkMonomGood(t, A, e)
	testutils.EvalAndOkGood(t, prog, 1)
}

// Test instantiation of a generic type with int32
func Test024(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type a Any()) struct { x a }"
	//Am := "func (x0 A(type a Any())) add1(type b Any())() b { return x0.x + 1 }"
	e := "A(int32){2147483647}.x + 1"
	//fggParseAndOkGood(t, Any, A, e)
	//fggParseAndOkMonomGood(t, Any, A, e)

	prog := fggParseAndOkMonomGood(t, Any, A, e)
	testutils.EvalAndOkGood(t, prog, 2)
}

// Test instantiation of a generic method with int32
func Test024b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) id(type b Any())(x b) b { return x }"
	e := "A(){}.id(int32)(52)"
	//fggParseAndOkGood(t, Any, A, Am, e)
	//fggParseAndOkMonomGood(t, Any, A, Am, e)

	prog := fggParseAndOkMonomGood(t, Any, A, Am, e)
	testutils.EvalAndOkGood(t, prog, 1)
}

// Test primitive ops with defined/named types
func Test025(t *testing.T) {
	A := "type A(type ) struct {}"
	AFact :="func (a A(type )) NewMyInt(type )(x MyInt()) MyInt() {return x}"
	MyInt := "type MyInt(type ) int32"
	Am := "func (x MyInt(type )) incr(type )() MyInt() {return x + 1}"
	e := "A(){}.NewMyInt()(1).incr()()"
	//fggParseAndOkGood(t, A, AFact, MyInt, Am, e)
	prog := fggParseAndOkGood(t, A, AFact, MyInt, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)

}

func Test025b(t *testing.T) {
	A := "type A(type ) struct {}"
	AFact :="func (a A(type )) NewMyString(type )(x MyString()) MyString() {return x}"
	MyString := "type MyString(type ) string"
	Am := "func (x MyString(type )) incr(type )() MyString() {return x + \"a\"}"
	e := "A(){}.NewMyString()(\"ol\").incr()()"
	//fggParseAndOkGood(t, A, AFact, MyString, Am, e)
	prog := fggParseAndOkGood(t, A, AFact, MyString, Am, e)
	testutils.EvalAndOkGood(t, prog, 3)
}


// Testing case where a variable (e.g. a struct field) has an interface
// literal - containing a generic method spec -  as its declared type.
func Test111(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct { x interface{ id(type T Any())(x T) T } }  "
	SFoo := "type SFoo1(type ) struct {}"
	SFoof := "func (s SFoo1(type )) id(type T Any())(x T) T { return x }"
	e := "A(){SFoo1(){}}.x.id(int32)(1)"
	prog := fggParseAndOkMonomGood(t, Any, A, SFoo, SFoof, e)
	testutils.EvalAndOkGood(t, prog, 3)
	_ = prog
}
// Testing cases where the only instantiation of a generic type
// appears in the definition of another type.
func Test200(t *testing.T) {
	Any := "type Any(type ) interface {}"
	Pair := "type Pair(type X Any(), Y Any()) struct { x X; y Y}"
	PairInt := "type PairInt(type ) Pair(int32, int32)"
	e := "PairInt(){1,2}.x + 1"
	prog := fggParseAndOkMonomGood(t, Any, Pair, PairInt, e)
	testutils.EvalAndOkGood(t, prog, 2)
	_ =  prog
}

func Test200b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	Pair := "type Pair(type X Any(), Y Any()) struct { x X; y Y}"
	PairEq := "type PairEq(type T Any()) Pair(T, T)"
	PairInt := "type PairInt(type ) PairEq(int32)"
	e := "PairInt(){1,2}.x + 1"
	prog := fggParseAndOkMonomGood(t, Any, Pair, PairEq, PairInt, e)
	testutils.EvalAndOkGood(t, prog, 2)
	_ =  prog
}