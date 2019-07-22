//rhu@HZHL4 MINGW64 ~/code/go/src/github.com/rhu1/fgg
//$ go test github.com/rhu1/fgg/fgg
//$ go test github.com/rhu1/fgg/fgg -run Test001

package fgg_test // Separate package, can test "API"

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rhu1/fgg/fgg"
)

/* Harness funcs */

func parseAndCheckOk(prog string) fgg.FGGProgram {
	var adptr fgg.FGGAdaptor
	ast := adptr.Parse(true, prog)
	allowStupid := false
	ast.Ok(allowStupid)
	return ast
}

func parseAndOkGood(t *testing.T, elems ...string) fgg.FGGProgram {
	prog := fgg.MakeFggProgram(elems...)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: " + fmt.Sprintf("%v", r) + "\n" +
				prog)
		}
	}()
	return parseAndCheckOk(prog)
}

// N.B. do not use to check for bad *syntax* -- see the "[Parser]" panic check
func parseAndOkBad(t *testing.T, msg string, elems ...string) fgg.FGGProgram {
	prog := fgg.MakeFggProgram(elems...)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but none occurred: " + msg + "\n" +
				prog)
		} else {
			rec := fmt.Sprintf("%v", r)
			if strings.HasPrefix(rec, "[Parser]") {
				t.Errorf("Unexpected panic: " + rec + "\n" + prog)
			}
			// TODO FIXME: check panic more specifically
		}
	}()
	return parseAndCheckOk(prog)
}

//*
// Pre: parseAndOkGood
func evalAndOkGood(t *testing.T, p fgg.FGGProgram, steps int) fgg.FGGProgram {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: " + fmt.Sprintf("%v", r) + "\n" +
				p.String())
		}
	}()
	allowStupid := true
	for i := 0; i < steps; i++ {
		p, _ = p.Eval() // CHECKME: check rule names as part of test?
		p.Ok(allowStupid)
	}
	return p
}

// Pre: parseAndOkGood
func evalAndOkBad(t *testing.T, p fgg.FGGProgram, msg string, steps int) fgg.FGGProgram {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but none occurred: " + msg + "\n" +
				p.String())
		} else {
			// [Parser] panic should be already checked by parseAndOkGood
			// TODO FIXME: check panic more specifically
		}
	}()
	allowStupid := true
	for i := 0; i < steps; i++ {
		p, _ = p.Eval()
		p.Ok(allowStupid)
	}
	return p
}

/* Syntax and typing */

// TOOD: classify FG-compatible subset compare results to -fg

// Initial FGG test
func Test001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	e := "B(A()){A(){}}"
	//type IA(type ) interface { m1(type )() Any };
	//type A1(type ) struct { };
	parseAndOkGood(t, Any, A, B, e)
}

func Test001b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	A1 := "type A1(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	e := "B(A()){A1(){}}"
	parseAndOkBad(t, "A1() is not an A()", Any, A, A1, B, e)
}

// Testing StructLit typing, t_S OK
func Test002(t *testing.T) {
	IA := "type IA(type ) interface { m1(type )() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() A() { return x0 }"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A(){}}"
	parseAndOkGood(t, IA, A, Am1, B, e)
}

func Test002b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type )() A() }"
	A := "type A(type ) struct {}"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A(){}}"
	parseAndOkBad(t, "A() is not an A1()", IA, A, B, e)
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
	parseAndOkGood(t, Any, IA, A, Am1, A1, B, e)
}

func Test003b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type )() Any() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a IA()) struct { f a }"
	e := "B(A()){A1(){}}"
	parseAndOkBad(t, "A1() is not an A()", Any, IA, A, Am1, A1, B, e)
}

// Initial testing for select on parameterised struct
func Test004(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct { fA Any() }"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a Any()) struct { fB a }"
	e := "B(A()){A(){A1(){}}}.fB.fA"
	parseAndOkGood(t, Any, A, Am1, A1, B, e)
}

func Test004b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct { fA Any() }"
	Am1 := "func (x0 A(type )) m1(type )() Any() { return x0 }"
	A1 := "type A1(type ) struct { }"
	B := "type B(type a Any()) struct { fB a }"
	e := "B(A1()){A1(){}}.fB.fA"
	parseAndOkBad(t, "A1 has no field fA", Any, A, Am1, A1, B, e)
}

// Initial testing for call
func Test005(t *testing.T) {
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type )() A() { return x0 }"
	e := "A(){}.m1()()"
	parseAndOkGood(t, A, Am1, e)
}

func Test006(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	e := "A(){}"
	parseAndOkGood(t, IA, A, e)
}

func Test006b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a A())() A() }"
	A := "type A(type ) struct {}"
	e := "A(){}"
	parseAndOkBad(t, "A() invalid upper bound", IA, A, e)
}

func Test007(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A())()"
	parseAndOkGood(t, Any, IA, A, Am1, A1, e)
}

func Test007b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1()()"
	parseAndOkBad(t, "Missing type actual", Any, IA, A, Am1, A1, e)
}

func Test007c(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() A() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() A() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A1())()"
	parseAndOkBad(t, "A1() is not an IA()", Any, IA, A, Am1, A1, e)
}

func Test007d(t *testing.T) {
	Any := "type Any(type ) interface {}"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "A(){}.m1(A())()"
	parseAndOkGood(t, Any, IA, A, Am1, A1, e)
}

// Testing Sig parsing
func Test008(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) B(a) { return x0 }"
	e := "A(){}"
	parseAndOkGood(t, IA, A, Am1, B, Bm2, e)
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
	parseAndOkGood(t, Any, IA, A, Am1, B, Bm2, e)
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
	parseAndOkBad(t, "A1() is not an A()", Any, IA, A, Am1, A1, B, Bm2, e)
}

// Initial test for generic type assertion
func Test0010(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	B := "type B(type a IA()) struct {}"
	Bm2 := "func (x0 B(type a IA())) m2(type )(x1 a) Any() { return x1 }" // Unnecessary
	e := "ToAny(){B(A()){}}.any.(B(A()))"
	parseAndOkGood(t, Any, ToAny, IA, A, Am1, B, Bm2, e)
}

func Test0011(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	ToIA := "type ToIA(type ) struct { upcast IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	e := "ToIA(){A(){}}.upcast.(A())"
	parseAndOkGood(t, IA, ToIA, A, Am1, e)
}

func Test0011b(t *testing.T) {
	IA := "type IA(type ) interface { m1(type a IA())() IA() }"
	ToIA := "type ToIA(type ) struct { upcast IA() }"
	A := "type A(type ) struct {}"
	Am1 := "func (x0 A(type )) m1(type a IA())() IA() { return x0 }"
	A1 := "type A1(type ) struct {}"
	e := "ToIA(){A(){}}.upcast.(A1())"
	parseAndOkBad(t, "A1() is not an IA", IA, ToIA, A, Am1, A1, e)
}

func Test0011c(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	B := "type B(type ) struct {}"
	Bm3 := "func (x0 B(type )) m3(type b Any())(x1 b) Any() { return x1 }"
	e := "ToAny(){B(){}}"
	parseAndOkGood(t, Any, ToAny, B, Bm3, e)
}

// Testing parsing for Call with both targ and arg
func Test0012(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type ) struct {}"
	Bm := "func (x0 B(type )) m(type a Any())(x1 a) a { return x1 }"
	e := "B(){}.m(A())(A(){})"
	parseAndOkGood(t, Any, A, B, Bm, e)
}

// Testing Call typing, meth-tparam TSubs of result
func Test0013(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type )) m(type a Any())(x1 a) a { return x1 }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f"
	parseAndOkGood(t, Any, A, B, Bm, e)
}

// Testing u <: a, i.e., upper is open type param
func Test0014(t *testing.T) {
	Any := "type Any(type ) interface {}"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type )) m(type a Any())() a { return A(){} }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f" // Eval would break type preservation, see TestEval001
	parseAndOkBad(t, Any, A, B, Bm, e)
}

/* Eval */

func TestEval001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	ToAny := "type ToAny(type ) struct { any Any() }"
	A := "type A(type ) struct {}"
	B := "type B(type a Any()) struct { f a }"
	Bm := "func (x0 B(type )) m(type a Any())(x1 a) a { return ToAny(){A(){}}.any.(a) }"
	e := "B(A()){A(){}}.m(B(A()))(B(A()){A(){}}).f"
	prog := parseAndOkGood(t, Any, ToAny, A, B, Bm, e)
	evalAndOkBad(t, prog, "Cannot caast A() to B(A())", 3)
}

// TOOD: classify FG-compatible subset compare results to -fg
