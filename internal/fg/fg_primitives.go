package fg

import (
	"github.com/rhu1/fgg/internal/base"
	"strconv"
)

// constants
// CHECKME: maybe these "representations" will change

type Tag int

const (
	INVALID Tag = iota
	INT8
	INT32
	INT64
)

var precisions = map[int]Tag {
	8 : INT8,
	32 : INT32,
	64 : INT64,
}

var NameToTags = map[string]Tag {
	"int8" : INT8,
	"int32" : INT32,
	"int64" : INT64,
}

/* "Exported" constructors (( ? for fgg (monomorph) ? )) */

func TagFromName(name string) Tag {
	//tag, ok := NameToTags[name]
	return NameToTags[name]
}


func NewIntLit(lit string) PrimitiveLiteral {
	for prec, tag := range precisions {
		if i, err := strconv.ParseInt(lit, 10, prec); err == nil {
			return PrimitiveLiteral{i, tag}
		}
	}
	return PrimitiveLiteral{tag: INVALID} // where to check for such invalid literals? how to propagate this info?
}

/* PrimitiveLiteral -- represents primitive type literals */

type PrimitiveLiteral struct {
	payload interface{}
	tag     Tag
}

var _ FGExpr = PrimitiveLiteral{}

//func (b PrimitiveLiteral) GetValue() interface{} {
//	return b.payload
//}

func (b PrimitiveLiteral) Subs(subs map[Variable]FGExpr) FGExpr {
	return b
}

func (b PrimitiveLiteral) Eval(ds []Decl) (FGExpr, string) {
	panic("Cannot reduce: " + b.String())
}

func (b PrimitiveLiteral) Typing(ds []Decl, gamma Gamma, allowStupid bool) Type {
	return TPrimitive{
		//name:      "",
		tag:       b.tag,
		undefined: true,
	}
}

func (b PrimitiveLiteral) IsValue() bool {
	return true
}

func (b PrimitiveLiteral) CanEval(ds []base.Decl) bool {
	return false
}

func (b PrimitiveLiteral) String() string {
	//return string(b.payload)
	panic("implement me PrimitiveLiteral.String")
}


func (b PrimitiveLiteral) ToGoString(ds []base.Decl) string {
	panic("implement me PrimitiveLiteral.ToGoString")
}


