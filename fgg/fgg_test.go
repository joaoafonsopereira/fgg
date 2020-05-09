//rhu@HZHL4 MINGW64 ~/code/go/src/github.com/rhu1/fgg
//$ go test github.com/rhu1/fgg/fgg
//$ go test github.com/rhu1/fgg/fgg -run Test001

package fgg_test // Separate package, can test "API"

import (
	"testing"

	"github.com/rhu1/fgg/base"
	"github.com/rhu1/fgg/base/testutils"
	"github.com/rhu1/fgg/fgg"
)

/* Harness funcs */

func fggParseAndOkGood(t *testing.T, elems ...string) base.Program {
	var adptr fgg.FGGAdaptor
	p := testutils.ParseAndOkGood(t, &adptr,
		fgg.MakeFggProgram(elems...)).(fgg.FGGProgram)
	fgg.Monomorph(p)
	return p
}

// N.B. do not use to check for bad *syntax* -- see the PARSER_PANIC_PREFIX panic check in base.ParseAndOkBad
func fggParseAndOkBad(t *testing.T, msg string, elems ...string) base.Program {
	var adptr fgg.FGGAdaptor
	return testutils.ParseAndOkBad(t, msg, &adptr, fgg.MakeFggProgram(elems...))
	// Not attempting monom on bad program
}

/* Syntax and typing */

// TOOD: classify FG-compatible subset compare results to -fg

// Initial FGG test

// Initial FGG test
func Test001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	e := "B(A()){A(){}}"
	//type IA(type ) interface { m1(type )() Any };
	//type A1(type ) struct { };
	fggParseAndOkGood(t, Any, A, B, e)
}

func Test001b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	A1 := "type A1(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	e := "B(A()){A1(){}}"
	fggParseAndOkBad(t, "A1() is not an A()", Any, A, A1, B, e)
}

// Testing StructLit typing, t_S OK
func Test002(t *testing.T) {
	IA := "type IA(type ) interface { m1(type )() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() A() { return x0 }"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A(){}}"
	fggParseAndOkGood(t, IA, A, Am1, B, e)
}

func Test002b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type )() A() }"
	A := "type A(type ) struct {}"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A(){}}"
	fggParseAndOkBad(t, "A() is not an A1()", IA, A, B, e)
}

// Testing fields (and t-args subs)
func Test003(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type )() Any() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A(){}}"
	fggParseAndOkGood(t, Any, IA, A, Am1, A1, B, e)
}

func Test003b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type )() Any() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A1(){}}"
	fggParseAndOkBad(t, "A1() is not an A()", Any, IA, A, Am1, A1, B, e)
}

// Initial testing for select on parameterised struct
func Test004(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct { fA Any() }"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a Any()) struct { fB a }"
	e := "B(A()){A(){A1(){}}}.fB.fA"
	fggParseAndOkGood(t, Any, A, Am1, A1, B, e)
}

func Test004b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct { fA Any() }"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a Any()) struct { fB a }"
	e := "B(A1()){A1(){}}.fB.fA"
	fggParseAndOkBad(t, "A1 has no field fA", Any, A, Am1, A1, B, e)
}

// Initial testing for call
func Test005(t *testing.T) {
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() A() { return x0 }"
	e := "A(){}.m1()()"
	fggParseAndOkGood(t, A, Am1, e)
}

func Test006(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	e := "A(){}"
	fggParseAndOkGood(t, IA, A, e)
}

func Test006b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a A())() A() }"
	A := "type A(type ) struct {}"
	e := "A(){}"
	fggParseAndOkBad(t, "A() invalid upper bound", IA, A, e)
}

func Test007(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A())()"
	fggParseAndOkGood(t, Any, IA, A, Am1, A1, e)
}

func Test007b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1()()"
	fggParseAndOkBad(t, "Missing type actual", Any, IA, A, Am1, A1, e)
}

func Test007c(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A1())()"
	fggParseAndOkBad(t, "A1() is not an IA()", Any, IA, A, Am1, A1, e)
}

func Test007d(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A())()"
	fggParseAndOkGood(t, Any, IA, A, Am1, A1, e)
}

// Testing Sig parsing
func Test008(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) B(a) { return x0 }"
	e := "A(){}"
	fggParseAndOkGood(t, IA, A, Am1, B, Bm2, e)
}

// Testing calls on parameterised struct
func Test009(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) B(a) { return x0 }"
	e := "B(A()){}.m2()(A(){})"
	fggParseAndOkGood(t, Any, IA, A, Am1, B, Bm2, e)
}

func Test009b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	A1 := "type A1(type ) struct {}"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) B(a) { return x0 }"
	e := "B(A()){}.m2()(A1(){})"
	fggParseAndOkBad(t, "A1() is not an A()", Any, IA, A, Am1, A1, B, Bm2, e)
}

// Initial test for generic type assertion
func Test010(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) Any() { return x1 }" // Unnecessary
	e := "ToAny(){B(A()){}}.any.(B(A()))"
	fggParseAndOkGood(t, Any, ToAny, IA, A, Am1, B, Bm2, e)
}

func Test011(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	ToIA := "type ToIA(type ) struct { upcast IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	e := "ToIA(){A(){}}.upcast.(A())"
	fggParseAndOkGood(t, IA, ToIA, A, Am1, e)
}

func Test011b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	ToIA := "type ToIA(type ) struct { upcast IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "ToIA(){A(){}}.upcast.(A1())"
	fggParseAndOkBad(t, "A1() is not an IA", IA, ToIA, A, Am1, A1, e)
}

func Test011c(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	B := "type B(type ) struct {}"
	Bm3 := "func (x0 B(type )) m3(type b Any())(x1 b) Any() { return x1 }"
	e := "ToAny(){B(){}}"
	fggParseAndOkGood(t, Any, ToAny, B, Bm3, e)
}

// Testing parsing for Call with both targ and arg
func Test012(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type ) struct {}"
	Bm := "func (x0 B(type )) m(type a Any())(x1 a) a { return x1 }"
	e := "B(){}.m(A())(A(){})"
	fggParseAndOkGood(t, Any, A, B, Bm, e)
}

// Testing Call typing, meth-tparam TSubs of result
func Test013(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type a Any())) m(type b Any())(x1 b) b { return x1 }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f"
	fggParseAndOkGood(t, Any, A, B, Bm, e)
}

// Testing u <: a, i.e., upper is open type param
func Test014(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type a Any())) m(type b Any())() b { return A(){} }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f" // Eval would break type preservation, see TestEval001
	fggParseAndOkBad(t, Any, A, B, Bm, e)
}

// testing sigAlphaEquals
func Test015(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) interface { m(type a Any())(x a) Any() }"
	B := "type B(type ) interface { m(type b Any())(x b) Any() }"
	C := "type C(type ) struct {}"
	Cm := "func (x0 C(type )) m(type b Any())(x b) Any() { return x0 }"
	D := "type D(type ) struct {}"
	Dm := "func (x0 D(type )) foo(type )(x A()) Any() { return x0 }"
	e := "D(){}.foo()(C(){})"
	fggParseAndOkBad(t, Any, A, B, C, Cm, D, Dm, e)
}

// testing covariant receiver bounds (MDecl.OK) -- cf. map.fgg (memberBr)
func Test016(t *testing.T) {
	Any := "Any(type ) interface {}"
	A := "type A(type a Any()) interface { m(type )(x a) Any() }" // param must occur in a meth sig
	B := "type B(type a A(a)) struct {}"                          // must have recursive param
	Bm := "func (x0 B(type b A(b))) m(type )(x b) Any() { return x0 }"
	D := "type D(type ) struct{}"
	e := "D(){}"
	fggParseAndOkBad(t, Any, A, B, Bm, D, e)
}

/* Monom */

// TODO: isMonomorphisable -- should fail that check
/*
func TestMonom001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type a Any())) m(type )() Any() { return B(B(a)){x0}.m()() }"
	e := "B(A()){A(){}}.m()()"
	parseAndOkBad(t, "Polymorphic recursion on the receiver type", Any, A, B, Bm, e)
}
//*/

//TODO: add -monom compose.fgg bug -- missing field type collection when visiting struct lits (e.g., Compose f, g types)
//TODO: add -monom map.fgg bug -- missing add-meth-param instans collection for interface type receivers (e.g., Bool().Cond(Bool())(...))

/* Eval */

// TOOD: classify FG-compatible subset compare results to -fg

func TestEval001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type a Any())) m(type b Any())(x1 b) b { return ToAny(){A(){}}.any.(b) }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f"
	prog := fggParseAndOkGood(t, Any, ToAny, A, B, Bm, e)
	testutils.EvalAndOkBad(t, prog, "Cannot cast A() to B(A())", 3)
}

/* fmt.Sprintf */

func TestEval002(t *testing.T) {
	imp := "import \"fmt\""
	A := "type A(type ) struct {}"
	e := "fmt.Sprintf(\"\")"
	prog := fggParseAndOkGood(t, imp, A, e)
	testutils.EvalAndOkGood(t, prog, 1)
}

func TestEval003(t *testing.T) {
	imp := "import \"fmt\""
	A := "type A(type ) struct {}"
	e := "fmt.Sprintf(\"%v ,_()+- %v\", A(){}, A(){})"
	prog := fggParseAndOkGood(t, imp, A, e)
	testutils.EvalAndOkGood(t, prog, 1)
}
