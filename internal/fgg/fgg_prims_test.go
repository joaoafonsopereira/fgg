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
	fggParseAndOkBad(t, "", A, Am, e)
}

// can't mix primitive types
func Test020(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) add(type )(x float32, y float64) float64 { return x+y }"
	e := "A(){}"
	fggParseAndOkBad(t, "", A, Am, e)
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
	fggParseAndOkBad(t,"", A, Am, e)
}

// Comparisons and logical ops
func Test022(t *testing.T) {
	A := "type A(type ) struct {}"
	Am := "func (x0 A(type )) cmp(type )(x float64, y float64) bool { return x > y || (x < x && false) }"
	e := "A(){}.cmp()(2, 4.2)"
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
	testutils.EvalAndOkGood(t, prog, 2)
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


/* Type lists */

// testing primary ops over generic types
func TestTLists001(t *testing.T) {
	A := "type IConstr(type ) interface { type int32, int64 }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T IConstr())(x T) T { return x + x }"
	e := "S(){}.add(int32)(5)"
	//fggParseAndOkGood(t, A, S, Sm, e)
	prog := fggParseAndOkMonomGood(t, A, S, Sm, e)
	testutils.EvalAndOkGood(t, prog, 2)
}

// using MyInt to instantiate type list
func TestTLists001b(t *testing.T) {
	A := "type IConstr(type ) interface { type int32, int64 }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T IConstr())(x T) T { return x + x }"
	MyInt := "type MyInt(type ) int32"
	e := "S(){}.add(MyInt())(5)"
	prog := fggParseAndOkMonomGood(t,  A, S, Sm, MyInt, e)
	testutils.EvalAndOkGood(t, prog, 2)
}

// testing wrong instantiation: type doesn't belong to type list
func TestTLists001c(t *testing.T) {
	A := "type IConstr(type ) interface { type int32, int64 }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T IConstr())(x T) T { return x + x }"
	MyFloat := "type MyFloat(type ) float32"
	e := "S(){}.add(MyFloat())(5.5)"

	//msg :=
	fggParseAndOkBad(t, "", A, S, Sm, MyFloat, e)
}

// Interface with type list & methods
func TestTLists002(t *testing.T) {
	A := "type IConstr(type ) interface { type int32, int64 ; String(type )() string }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T IConstr())(x T) string { return x.String()() + \"b\" }"
	MyInt := "type MyInt(type ) int32"
	MyIntStr := "func (m MyInt(type )) String(type )() string { return \"myinttt\" }"
	e := "S(){}.add(MyInt())(1)"
	//fggParseAndOkGood(t, A, S, Sm, MyInt, MyIntStr, e)
	prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	testutils.EvalAndOkGood(t, prog, 3)
}

// Adding iface value and const (ok)
func TestTLists003(t *testing.T) {
	A := "type Numeric(type ) interface { type int32, int64, float32, float64 }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T Numeric())(x T) T { return x + 1 }"
	Sm2 := "func (s S(type )) add2(type T Numeric())(x T) T { return 1 + x }" // just to test different order in typing
	MyInt := "type MyInt(type ) int32"
	e := "S(){}.add(MyInt())(4)"
	//fggParseAndOkGood(t, A, S, Sm, MyInt, e)
	prog := fggParseAndOkMonomGood(t, A, S, Sm, Sm2, MyInt, e)
	testutils.EvalAndOkGood(t, prog, 2)
}

// Adding iface value and const (wrong -- string is not numeric)
// -> cannot convert 1 (untyped int constant) to T
func TestTLists003b(t *testing.T) {
	A := "type NumericS(type ) interface { type int32, int64, float32, float64, string }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T NumericS())(x T) T { return x + 1 }"
	MyInt := "type MyString(type ) string"
	e := "S(){}.add(MyString())(\"a\")"

	msg := "mismatched types T and int32(undefined)" // Constraint includes string; can't convert 1 to a string
	fggParseAndOkBad(t, msg, A, S, Sm, MyInt, e)
}

// Adding iface value and int32, wrong todo why? -> mismatched types T and int32
func TestTLists003c(t *testing.T) {
	A := "type NumericS(type ) interface { type int32, int64, float32, float64, string }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T NumericS())(x T, y int32) T { return x + y }"
	MyInt := "type MyInt(type ) int32"
	e := "S(){}.add(MyInt())(4)"

	msg := "mismatched types T and int32"
	fggParseAndOkBad(t, msg, A, S, Sm, MyInt, e) // this is the one I want

}

// Adding 2 type params with the same bounds (bad)
func TestTLists004(t *testing.T) {
	A := "type NumericS(type ) interface { type int32, int64, float32, float64 }"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) add(type T NumericS(), T2 NumericS())(x T, y T2) bool { return (x + y) > 0 }"
	MyInt := "type MyInt(type ) int32"
	e := "S(){}.add(MyInt(), MyInt())(4, 5)"

	fggParseAndOkBad(t, "mismatched types T and T2", A, S, Sm, MyInt, e)
}

// Types with methods in type lists: method can't be called unless
// it is explicitly specified in the constraint -- even though
// all the listed types have that method.
// Cf. https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#types-with-methods-in-type-lists
func TestTLists005(t *testing.T) {
	MyInt := "type MyInt(type ) int32"
	MyIntS := "func (m MyInt(type )) String(type )() string { return \"myint\" }"
	MyFloat := "type MyFloat(type ) float32"
	MyFloatS := "func (m MyFloat(type )) String(type )() string { return \"myfloat\" }"

	MyIorF := "type MyIntOrFloat(type ) interface { type MyInt(), MyFloat() }"
	S := "type S(type ) struct {}"
	ToString := "func (s S(type )) ToString(type T MyIntOrFloat())(v T) string { return v.String()() }"
	//e := "S(){}.ToString(MyInt())(1)"
	e := "S(){}" // the error must be caught at MethDecl.Ok

	msg := "Method not found: String in T" // or smthing like that
	fggParseAndOkBad(t, msg, S, MyInt, MyIntS, MyFloat, MyFloatS, MyIorF, ToString, e)
}

func TestTLists005b(t *testing.T) {
	MyInt := "type MyInt(type ) int32"
	MyIntS := "func (m MyInt(type )) String(type )() string { return \"myint\" }"
	MyFloat := "type MyFloat(type ) float32"
	MyFloatS := "func (m MyFloat(type )) String(type )() string { return \"myfloat\" }"

	MyIorF := "type MyIntOrFloatStringer(type ) interface { type MyInt(), MyFloat(); String(type )() string }"
	S := "type S(type ) struct {}"
	ToString := "func (s S(type )) ToString(type T MyIntOrFloatStringer())(v T) string { return v.String()() }"
	e := "S(){}.ToString(MyInt())(1)"
	fggParseAndOkGood(t, S, MyInt, MyIntS, MyFloat, MyFloatS, MyIorF, ToString, e)
}

