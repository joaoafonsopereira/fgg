/*
 * This file contains defs for "concrete" syntax w.r.t. programs and decls.
 * Base ("abstract") types, interfaces, etc. are in fg.go.
 */

package fg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rhu1/fgg/internal/base"
)

var _ = fmt.Errorf

/* "Exported" constructors (e.g., for fgg_monom)*/

func NewFGProgram(ds []Decl, e FGExpr, printf bool) FGProgram {
	return FGProgram{ds, e, printf}
}

func NewTypeDecl(name Name, srcType Type) TypeDecl {
	return TypeDecl{name, srcType}
}

// TODO: NewMethDecl
func NewMDecl(recv ParamDecl, m Name, pds []ParamDecl, t Type, e FGExpr) MethDecl {
	return MethDecl{recv, m, pds, t, e}
}
func NewFieldDecl(f Name, t Type) FieldDecl      { return FieldDecl{f, t} }
func NewParamDecl(x Name, t Type) ParamDecl      { return ParamDecl{x, t} } // For fgg_monom.MakeWMap
func NewSig(m Name, pds []ParamDecl, t Type) Sig { return Sig{m, pds, t} }  // For fgg_monom.MakeWMap

/* Program */

type FGProgram struct {
	decls  []Decl
	e_main FGExpr
	printf bool // false = "original" `_ = e_main` syntax; true = import-fmt/printf syntax
	// N.B. coincidentally "behaves" like an actual printf because interpreter prints out final eval result
}

var _ base.Program = FGProgram{}
var _ FGNode = FGProgram{}

// From base.Program
func (p FGProgram) GetDecls() []Decl   { return p.decls } // Return a copy?
func (p FGProgram) GetMain() base.Expr { return p.e_main }
func (p FGProgram) IsPrintf() bool     { return p.printf } // HACK

// From base.Program
func (p FGProgram) Ok(allowStupid bool, _mode base.TypingMode) (base.Type, base.Program) {
	tds := make(map[string]TypeDecl) // Type name
	mds := make(map[string]MethDecl) // Hack, string = string(md.recv.t) + "." + md.name
	for _, v := range p.decls {
		switch d := v.(type) {
		case TypeDecl:
			d.Ok(p.decls) // Currently empty -- TODO: check, e.g., unique field names -- cf., above [Warning]
			// N.B. checks also omitted from submission version
			t := d.GetName()
			if _, ok := tds[t]; ok {
				panic("Multiple declarations of type name: " + t + "\n\t" +
					d.String())
			}
			tds[t] = d
		case MethDecl:
			d.Ok(p.decls)
			hash := d.recv.t.String() + "." + d.name
			if _, ok := mds[hash]; ok {
				panic("Multiple declarations for receiver " + d.recv.t.String() +
					" of the method name: " + d.name + "\n\t" + d.String())
			}
			mds[hash] = d
		default:
			panic("Unknown decl: " + reflect.TypeOf(v).String() + "\n\t" +
				v.String())
		}
	}
	var gamma Gamma // Empty env for main
	typ, ast := p.e_main.Typing(p.decls, gamma, allowStupid)
	return typ, FGProgram{p.decls, ast, p.printf}
}

// CHECKME: resulting FGProgram is not parsed from source, OK? -- cf. Expr.Eval
// But doesn't affect FGPprogam.Ok() (i.e., Expr.Typing)
// From base.Program
func (p FGProgram) Eval() (base.Program, string) {
	e, rule := p.e_main.Eval(p.decls)
	return FGProgram{p.decls, e.(FGExpr), p.printf}, rule
}

func (p FGProgram) String() string {
	var b strings.Builder
	b.WriteString("package main;\n")
	if p.printf {
		b.WriteString("import \"fmt\";\n")
	}
	for _, v := range p.decls {
		b.WriteString(v.String())
		b.WriteString(";\n")
	}
	b.WriteString("func main() { ")
	if p.printf {
		b.WriteString("fmt.Printf(\"%#v\", ")
		b.WriteString(p.e_main.String())
		b.WriteString(")")
	} else {
		b.WriteString("_ = ")
		b.WriteString(p.e_main.String())
	}
	b.WriteString(" }")
	return b.String()
}

/* MethDecl, ParamDecl */

type MethDecl struct {
	recv   ParamDecl
	name   Name // Not embedding Sig because Sig doesn't take xs
	pDecls []ParamDecl
	t_ret  Type // Return
	e_body FGExpr
}

var _ Decl = MethDecl{}

func (md MethDecl) GetReceiver() ParamDecl     { return md.recv }
func (md MethDecl) GetName() Name              { return md.name }   // From Decl
func (md MethDecl) GetParamDecls() []ParamDecl { return md.pDecls } // Returns non-receiver params
func (md MethDecl) GetReturn() Type            { return md.t_ret }
func (md MethDecl) GetBody() FGExpr            { return md.e_body }

func (md MethDecl) Ok(ds []Decl) {
	md.recv.t.Ok(ds)

	if isInterfaceType(ds, md.recv.t) {
		panic("Invalid receiver type: " + md.recv.t.String() +
			"\n\t" + md.String())
	}
	// distinct, params ok
	env := Gamma{md.recv.name: md.recv.t}
	for _, v := range md.pDecls {
		if _, ok := env[v.name]; ok {
			panic("Multiple receiver/parameters with name " + v.name + "\n\t" +
				md.String())
		}
		v.t.Ok(ds)
		env[v.name] = v.t
	}
	md.t_ret.Ok(ds)
	allowStupid := false
	// don't care about 'ast' returned from typing of method body -- only from method Call
	t, _ := md.e_body.Typing(ds, env, allowStupid)
	if !t.Impls(ds, md.t_ret) {
		panic("Method body type must implement declared return type: found=" +
			t.String() + ", expected=" + md.t_ret.String() + "\n\t" + md.String())
	}
}

func (md MethDecl) ToSig() Sig {
	return Sig{md.name, md.pDecls, md.t_ret}
}

func (md MethDecl) String() string {
	var b strings.Builder
	b.WriteString("func (")
	b.WriteString(md.recv.String())
	b.WriteString(") ")
	b.WriteString(md.name)
	b.WriteString("(")
	writeParamDecls(&b, md.pDecls)
	b.WriteString(") ")
	b.WriteString(md.t_ret.String())
	b.WriteString(" { return ")
	b.WriteString(md.e_body.String())
	b.WriteString(" }")
	return b.String()
}

// Cf. FieldDecl  // Rename PDecl?
type ParamDecl struct {
	name Name // CHECKME: Variable? (also Env key)
	t    Type
}

var _ FGNode = ParamDecl{}

func (pd ParamDecl) GetName() Name { return pd.name } // From Decl
func (pd ParamDecl) GetType() Type { return pd.t }

func (pd ParamDecl) String() string {
	var b strings.Builder
	b.WriteString(pd.name)
	b.WriteString(" ")
	b.WriteString(pd.t.String())
	return b.String()
}

/* Sig */

type Sig struct {
	meth   Name
	pDecls []ParamDecl
	t_ret  Type
}

var _ Spec = Sig{}

func (g Sig) GetMethod() Name            { return g.meth }
func (g Sig) GetParamDecls() []ParamDecl { return g.pDecls }
func (g Sig) GetReturn() Type            { return g.t_ret }

func (g0 Sig) Ok(ds []Decl) {
	seen := make(map[Type]ParamDecl)
	for _, v := range g0.pDecls {
		if _, ok := seen[v.t]; ok {
			panic("Multiple parameters with same name: " + v.name +
				"\n\t" + g0.String())
		}
		v.t.Ok(ds)
	}
	g0.t_ret.Ok(ds)
}

// !!! Sig in FG (also, Go spec) includes ~x, which naively breaks "impls"
func (g0 Sig) EqExceptVars(g Sig) bool {
	if len(g0.pDecls) != len(g.pDecls) {
		return false
	}
	for i := 0; i < len(g0.pDecls); i++ {
		if g0.pDecls[i].t != g.pDecls[i].t {
			return false
		}
	}
	return g0.meth == g.meth && g0.t_ret == g.t_ret
}

// From Spec
func (g Sig) GetSigs(_ []Decl) []Sig {
	return []Sig{g}
}

func (g Sig) String() string {
	var b strings.Builder
	b.WriteString(g.meth)
	b.WriteString("(")
	writeParamDecls(&b, g.pDecls)
	b.WriteString(") ")
	b.WriteString(g.t_ret.String())
	return b.String()
}

type TypeDecl struct {
	name    Name
	srcType Type
}

var _ Decl = TypeDecl{}

func (t TypeDecl) GetName() Name       { return t.name }
func (t TypeDecl) GetSourceType() Type { return t.srcType }

func (t TypeDecl) Ok(ds []base.Decl) {
	t.srcType.Ok(ds)
	checkCyclicTypeDecl(ds, t, t.srcType)
}

func (t TypeDecl) String() string {
	var b strings.Builder
	b.WriteString("type ")
	b.WriteString(string(t.name))
	b.WriteString(" ")
	b.WriteString(t.srcType.String())
	return b.String()
}

/* Helpers */

// For a type declaration decl, searches for any occurrence
// of decl.GetName() in the target type, recursively
func checkCyclicTypeDecl(ds []Decl, decl TypeDecl, target Type) {
	switch target := target.(type) {
	case TPrimitive:
		return
	case TNamed:
		if target.GetName() == decl.GetName() {
			panic("Invalid cyclic declaration: " + decl.String())
		}
		targetDecl := getTDecl(ds, target.GetName())
		checkCyclicTypeDecl(ds, decl, targetDecl.GetSourceType())

	case STypeLit:
		for _, f := range target.GetFieldDecls() {
			if u, ok := f.t.(TNamed); ok {
				//if isStructType(ds, u) // CHECKME: without this check, the next call may be needlessly checking for cycles in u_I's -- cf. commented checkCyclicTypeDecl
				checkCyclicTypeDecl(ds, decl, u)
			}
		}
	case ITypeLit:
		for _, s := range target.GetSpecs() {
			if u, ok := s.(TNamed); ok {
				// u is a u_I, checked in Ok
				checkCyclicTypeDecl(ds, decl, u)
			}
		}
	}
}

// Doesn't include "(...)" -- slightly more convenient for debug messages
func writeFieldDecls(b *strings.Builder, fds []FieldDecl) {
	if len(fds) > 0 {
		b.WriteString(fds[0].String())
		for _, v := range fds[1:] {
			b.WriteString("; " + v.String())
		}
	}
}

func writeParamDecls(b *strings.Builder, pds []ParamDecl) {
	if len(pds) > 0 {
		b.WriteString(pds[0].String())
		for _, v := range pds[1:] {
			b.WriteString(", " + v.String())
		}
	}
}

/* Old */

//*/

// RH: Possibly refactor aspects of this and related as "Decl.Wf()" -- the parts of "Ok()" omitted from the paper
func isDistinctDecl(decl Decl, ds []Decl) bool {
	var count int
	for _, d := range ds {
		switch d := d.(type) {
		case TypeDecl:
			// checks that type-name is unique regardless of definition
			// RH: Refactor as a single global pass (use a temp map), or into a TDecl.Wf() -- done: currently integrated into FGProgram.Ok for now (to avoid a second iteration)
			if td, ok := decl.(TypeDecl); ok && d.GetName() == td.GetName() {
				count++
			}
		case MethDecl:
			// checks that (method-type, method-name) is unique
			// RH: CHECKME: this would allow (bad) "return overloading"? -- note, d.t is the method return type
			if md, ok := decl.(MethDecl); ok && d.t_ret.String() == md.t_ret.String() && d.GetName() == md.GetName() {
				count++
			}
		}
	}
	return count == 1
}

//*/
