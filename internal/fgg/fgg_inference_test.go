package fgg_test

import (
	"github.com/rhu1/fgg/internal/base"
	"github.com/rhu1/fgg/internal/base/testutils"
	"github.com/rhu1/fgg/internal/fgg"
	"github.com/rhu1/fgg/internal/parser"
	"testing"
)

/* Harness funcs */

func fggParseAndInferGood(t *testing.T, elems ...string) base.Program {
	var adptr parser.FGGAdaptor
	p := testutils.ParseAndInferGood(t, &adptr,
		fgg.MakeFggProgram(elems...))
	return p
}

/* Common declarations */

const NL = ";\n"
const Any = "type Any(type ) interface {}"
const S   = "type S(type ) struct {}"

/* Simple inference, not taking untyped constants into account */

func TestInfer001(t *testing.T) {
	//S := "type S(type ) struct {}"
	AnyS := "type Any(type ) interface {}" + ";\n" + "type S(type ) struct {}"
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "S(){}.id()(S(){})"

	//fggParseAndInferGood(t, Any, S, Sm, e)
	fggParseAndInferGood(t, AnyS, Sm, e)
}

func TestInfer001b(t *testing.T) {
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	Sm2 := "func (s S(type )) id2(type T2 Any())(y T2) T2 { return s.id()(y) }"
	e := "S(){}.id2()(S(){})"
	fggParseAndInferGood(t, Any, S, Sm, Sm2, e)
}

// Infer method's type parameter
func TestInfer002b(t *testing.T) {
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "S(){}.id()(1)"

	fggParseAndInferGood(t, Any, S, Sm, e)
	//testutils.EvalAndOkGood(t, prog, 1)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
}


func TestInfer0000(t *testing.T) {
	Sm := "func (s S(type )) id(type )(x int32) Any() { return x }"
	e := "S(){}.id()(1)"

	fggParseAndInferGood(t, Any, S, Sm, e)
	//testutils.EvalAndOkGood(t, prog, 1)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
}



func TestInfer002c(t *testing.T) {
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "S(){}.id()(S(){}).id()(1)" // somewhat more complicated that in test 001

	fggParseAndInferGood(t, Any, S, Sm, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

func TestInfer002d(t *testing.T) {
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	Sm2 := "func (s S(type )) id2(type T2 Any())(y T2) T2 { return s.id()(y) }" // instantiates id's parameter with another type variable (T2)

	e := "S(){}.id()(1)" // somewhat more complicated that in test 001

	fggParseAndInferGood(t, Any, S, Sm, Sm2, e)
}

// Infer type's type parameter (TODO should this be possible?)
func TestInfer003(t *testing.T) {
	Box := "type Box(type T Any()) struct {x T}"
	e := "Box(){false}"

	fggParseAndInferGood(t, Any, Box, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

func TestInfer004(t *testing.T) {
	Box := "type Box(type T Any()) struct {x T}"
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "Box(){false}.x || true" // <- ver ha problema durante eval por TNamed nao estar "corretamente construido" (todo teste talvez nao seja bem assim)

	fggParseAndInferGood(t, Any, S, Box, Sm, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

const Function = "type Function(type a Any(), b Any()) interface { Apply(type )(x a) b }"
const IncrFunc = "type incr(type ) struct { n int32 }" + NL +
	"func (this incr(type )) Apply(x int32) int32 { " +
	"  return x + this.n" +
	"}"
const PosFunc = "type pos(type ) struct {}" + NL +
	"func (this pos(type )) Apply(type )(x int32) bool {" +
	"  return x > 0" +
	"}"

const List = "type List(type a Any()) interface { " +
	"  Map(type b Any())(f Function(a,b)) List(b) " +
	"}" + NL +
	"type Nil(type a Any()) struct {}" + NL +
	"type Cons(type a Any()) struct { " +
	"  head a; " +
	"  tail List(a) " +
	"}"
const MapImpl = "func (xs Nil(type a Any())) Map(type b Any())(f Function(a,b)) List(b) {" +
	"  return Nil(){}" +
	"}" + NL +
	"func (xs Cons(type a Any())) Map(type b Any())(f Function(a,b)) List(b) { " +
	"  return Cons(){f.Apply()(xs.head), xs.tail.Map()(f)}" +
	"}"

// Tests inference inside MethDecls
// -> todo "corner case" having to test Nil(aa1) <: List(b) in decl of Nil.Map
func TestInfer_mapExample001(t *testing.T) {
	e := "1"
	fggParseAndInferGood(t, Any, Function, List, MapImpl, e)
}

// testing <:unify Nil(aa1) and List(S{})
func TestInfer_mapExample001b(t *testing.T) {
	e := "Cons(){ S(){}, Nil(){} }"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, e)
}

// 2-elem list
func TestInfer_mapExample001bb(t *testing.T) {
	e := "Cons(){ S(){}, Cons(){S(){}, Nil(){}} }"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, e)
}

// list of lists
func TestInfer_mapExample001bbb(t *testing.T) {
	e := "Cons(){ Cons(){S(){}, Nil(){}}, Nil(){} }"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, e)
}

func TestInfer_mapExample002(t *testing.T) {
	idStruct := "type idStruct(type ) struct {}" + NL +
		"func (this idStruct(type )) Apply(type )(x S()) S() {" +
		"  return x" +
		"}"
	e := "Cons(){ S(){}, Nil(){} }.Map()(idStruct(){})"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, idStruct, e)
}

func TestInfer_mapExample003(t *testing.T) {
	foo := "type foo(type ) struct {}" + NL +
		"func (this foo(type )) Apply(type )(x S()) bool {" +
		"  return false" +
		"}"
	e := "Cons(){ S(){}, Nil(){} }.Map()(foo(){})"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, foo, e)
}


// Goal: having to unify Cons(S{}), List(αα1)
func TestInfer_mapExample004(t *testing.T) {
	SFoo := "type SFoo(type ) struct {}"
	SFooM := "func (this SFoo(type )) foo(type b Any())(x List(b)) List(b) { return x }"
	e := "SFoo(){}.foo()( Cons(){S(){}, Nil(){}} )"
	fggParseAndInferGood(t, Any, S, Function, List, MapImpl, SFoo, SFooM, e)
}

// De Cons(aa1){S(){}, Nil(){}}  vão sair constraints:
// S()      <= aa1
// Nil(aa2) <= List(aa1)

// E objetivo principal será ter que unificar
// Cons(S())  -- tipo que vai sair de Cons.Infer
// List(aaX)  -- tipo resultante de generalizaçao de signature