package fgg_test

import "testing"

// Infer method's type parameter
func TestInfer001(t *testing.T) {
	Any := "type Any(type ) interface {}"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "S(){}.id()(1)"

	fggParseAndOkGood(t, Any, S, Sm, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

func TestInfer001b(t *testing.T) {
	Any := "type Any(type ) interface {}"
	S := "type S(type ) struct {}"
	Sm := "func (s S(type )) id(type T Any())(x T) T { return x }"
	e := "S(){}.id()(S(){}).id()(1)" // somewhat more complicated that in test 001

	fggParseAndOkGood(t, Any, S, Sm, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}

// Infer type's type parameter (TODO should this be possible?)
func TestInfer002(t *testing.T) {
	Any := "type Any(type ) interface {}"
	S := "type Box(type T Any()) struct {x T}"
	e := "Box()(false)"

	fggParseAndOkGood(t, Any, S, e)
	//prog := fggParseAndOkMonomGood(t, A, S, Sm, MyInt, MyIntStr, e)
	//testutils.EvalAndOkGood(t, prog, 3)
}
