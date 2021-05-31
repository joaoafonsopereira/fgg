/*
 * This file contains defs for "concrete" syntax w.r.t. programs and decls.
 * Base ("abstract") types, interfaces, etc. are in fgg.go.
 */

package fgg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rhu1/fgg/internal/base"
)

var _ = fmt.Errorf
var _ = reflect.Append

/* Public constructors */

// TODO: rename NewFGGProgram
func NewProgram(ds []Decl, e FGGExpr, printf bool) FGGProgram {
	return FGGProgram{ds, e, printf}
}

func NewTypeDecl(name Name, Psi BigPsi, srcType Type) TypeDecl {
	return TypeDecl{name, Psi, srcType}
}

func NewMethDecl(
	x_recv Name,
	t_recv Name,
	Psi_recv BigPsi,
	name Name,
	Psi_meth BigPsi,
	pDecls []ParamDecl,
	u_ret Type,
	e_body FGGExpr) MethDecl {
	return MethDecl{x_recv, t_recv, Psi_recv, name, Psi_meth, pDecls, u_ret, e_body}
}

// TODO: rename NewMethDecl
func NewMDecl(x_recv Name,
	t_recv Name,
	Psi_recv BigPsi,
	name Name,
	Psi_meth BigPsi,
	pDecls []ParamDecl,
	u_ret Type,
	e_body FGGExpr) MethDecl {
	return MethDecl{x_recv, t_recv, Psi_recv, name, Psi_meth, pDecls, u_ret, e_body}
}
func NewFieldDecl(f Name, t Type) FieldDecl                  { return FieldDecl{f, t} }
func NewParamDecl(x Name, t Type) ParamDecl                  { return ParamDecl{x, t} }     // For fgg_monom.MakeWMap
func NewSig(m Name, Psi BigPsi, pds []ParamDecl, t Type) Sig { return Sig{m, Psi, pds, t} } // For fgg_monom.MakeWMap

/* Program */

type FGGProgram struct {
	decls  []Decl
	e_main FGGExpr
	printf bool // false = "original" `_ = e_main` syntax; true = import-fmt/printf syntax
	// N.B. coincidentally "behaves" like an actual printf because interpreter prints out final eval result
}

var _ base.Program = FGGProgram{}
var _ FGGNode = FGGProgram{}

func (p FGGProgram) GetDecls() []Decl   { return p.decls } // Return a copy?
func (p FGGProgram) GetMain() base.Expr { return p.e_main }
func (p FGGProgram) IsPrintf() bool     { return p.printf } // HACK

func (p FGGProgram) Ok(allowStupid bool) (base.Type, base.Program) {
	tds := make(map[string]TypeDecl) // Type name
	mds := make(map[string]MethDecl) // Hack, string = md.recv.t + "." + md.name
	for _, v := range p.decls {
		switch d := v.(type) {
		case TypeDecl:
			d.Ok(p.decls)
			t := d.GetName()
			if _, ok := tds[t]; ok {
				panic("Multiple declarations of type name: " + t + "\n\t" +
					d.String())
			}
			tds[t] = d
		case MethDecl:
			d.Ok(p.decls)
			hash := string(d.t_recv) + "." + d.name
			if _, ok := mds[hash]; ok {
				panic("Multiple declarations for receiver " + string(d.t_recv) +
					" of the method name: " + d.name + "\n\t" + d.String())
			}
			mds[hash] = d
		default:
			panic("Unknown decl: " + reflect.TypeOf(v).String() + "\n\t" +
				v.String())
		}
	}
	// Empty envs for main
	var delta Delta
	var gamma Gamma
	typ, ast := p.e_main.Typing(p.decls, delta, gamma, allowStupid)
	return typ, FGGProgram{p.decls, ast, p.printf}
}

func (p FGGProgram) Eval() (base.Program, string) {
	e, rule := p.e_main.Eval(p.decls)
	return FGGProgram{p.decls, e.(FGGExpr), p.printf}, rule
}

func (p FGGProgram) String() string {
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
	x_recv   Name // CHECKME: better to be Variable?  (etc. for other such Names)
	t_recv   Name // N.B. t_S
	Psi_recv BigPsi
	// N.B. receiver elements "decomposed" because Psi (not TNamed, cf. fg.MDecl uses ParamDecl)
	name     Name // Refactor to embed Sig?
	Psi_meth BigPsi
	pDecls   []ParamDecl
	u_ret    Type // Return
	e_body   FGGExpr
}

var _ Decl = MethDecl{}

func (md MethDecl) GetRecvName() Name          { return md.x_recv }
func (md MethDecl) GetRecvTypeName() Name      { return md.t_recv }
func (md MethDecl) GetRecvPsi() BigPsi         { return md.Psi_recv }
func (md MethDecl) GetName() Name              { return md.name }
func (md MethDecl) GetMDeclPsi() BigPsi        { return md.Psi_meth } // MDecl in name to prevent false capture by TDecl interface
func (md MethDecl) GetParamDecls() []ParamDecl { return md.pDecls }
func (md MethDecl) GetReturn() Type            { return md.u_ret }
func (md MethDecl) GetBody() FGGExpr           { return md.e_body }

func (md MethDecl) Ok(ds []Decl) {
	// (type t_S(Phi') T ) ∈ D
	recv_decl := getTDecl(ds, md.t_recv) // panics if not found
	if isIfaceType(ds, recv_decl.GetSourceType()) {
		panic("Invalid receiver type: " + md.t_recv +
			"\n\t" + md.String())
	}
	// Phi_md <: Phi_td
	tfs_md := md.Psi_recv.tFormals
	tfs_td := recv_decl.GetBigPsi().tFormals
	if len(tfs_td) != len(tfs_md) {
		panic("Receiver type parameter arity mismatch:\n\tmdecl=" + md.t_recv +
			md.Psi_recv.String() + ", tdecl=" + recv_decl.GetName() +
			"\n\t" + recv_decl.GetBigPsi().String())
	}
	subs_md := makeParamIndexSubs(md.Psi_recv)
	subs_td := makeParamIndexSubs(recv_decl.GetBigPsi())
	for i := 0; i < len(tfs_td); i++ {
		if !tfs_md[i].u_I.SubsEtaOpen(subs_md).Impls(ds, tfs_td[i].u_I.SubsEtaOpen(subs_td)) { // Canonicalised
			panic("Receiver parameter upperbound not a subtype of type decl upperbound:" +
				"\n\tmdecl=" + tfs_md[i].String() +
				", tdecl=" + tfs_td[i].String())
		}
	}
	// Phi, Psi ok
	md.Psi_recv.Ok(ds, make(Delta))
	delta := md.Psi_recv.ToDelta()
	md.Psi_meth.Ok(ds, delta)
	for _, v := range md.Psi_meth.tFormals {
		delta[v.name] = v.u_I
	}
	// distinct; params ok; construct gamma for body typing
	as := md.Psi_recv.Hat()                          // !!! submission version, x:t_S(a) => x:t_S(~a)
	gamma := Gamma{md.x_recv: TNamed{md.t_recv, as}} // CHECKME: can we give the bounds directly here instead of 'as'?
	seen := make(map[Name]Name)
	seen[md.x_recv] = md.x_recv
	for _, v := range md.pDecls {
		if _, ok := seen[v.name]; ok {
			panic("Duplicate receiver/param name: " + v.name + "\n\t" + md.String())
		}
		seen[v.name] = v.name
		v.u.Ok(ds, delta)
		gamma[v.name] = v.u
	}
	md.u_ret.Ok(ds, delta)
	allowStupid := false
	// don't care about 'ast' returned from typing of method body -- only from method Call
	u, _ := md.e_body.Typing(ds, delta, gamma, allowStupid)

	/*fmt.Println("a:", u)
	fmt.Println("b:", md.u_ret)
	fmt.Println("c:", u.ImplsDelta(ds, delta, md.u_ret))*/

	if !u.ImplsDelta(ds, delta, md.u_ret) {
		panic("Method body type must implement declared return type: found=" +
			u.String() + ", expected=" + md.u_ret.String() + "\n\t" + md.String())
	}
}

func (md MethDecl) ToSig() Sig {
	return Sig{md.name, md.Psi_meth, md.pDecls, md.u_ret}
}

func (md MethDecl) String() string {
	var b strings.Builder
	b.WriteString("func (")
	//b.WriteString(md.recv.String())
	b.WriteString(md.x_recv)
	b.WriteString(" ")
	b.WriteString(md.t_recv)
	b.WriteString(md.Psi_recv.String())
	b.WriteString(") ")
	b.WriteString(md.name)
	b.WriteString(md.Psi_meth.String())
	b.WriteString("(")
	writeParamDecls(&b, md.pDecls)
	b.WriteString(") ")
	b.WriteString(md.u_ret.String())
	b.WriteString(" { return ")
	b.WriteString(md.e_body.String())
	b.WriteString(" }")
	return b.String()
}

// Cf. FieldDecl
type ParamDecl struct {
	name Name // CHECKME: Variable?
	u    Type
}

var _ FGGNode = ParamDecl{}

func (pd ParamDecl) GetName() Name { return pd.name }
func (pd ParamDecl) GetType() Type { return pd.u }

func (pd ParamDecl) String() string {
	return pd.name + " " + pd.u.String()
}

/* Sig */

type Sig struct {
	meth   Name
	Psi    BigPsi // Add-meth-tparams
	pDecls []ParamDecl
	u_ret  Type
}

var _ Spec = Sig{}

func (g Sig) GetMethod() Name            { return g.meth }
func (g Sig) GetPsi() BigPsi             { return g.Psi }
func (g Sig) GetParamDecls() []ParamDecl { return g.pDecls }
func (g Sig) GetReturn() Type            { return g.u_ret }

//func (g Sig) TSubs(subs map[TParam]Type) Sig {
// Only makes sense to have SubsEtaOpen, as eta will never contain mappings
// for the type vars belonging to g.Psi. TODO why is that?
// The parameters are only fully instantiated in monomSig1 [fgg_monom.go]
func (g Sig) SubsEtaOpen(eta EtaOpen) Sig {
	tfs := make([]TFormal, len(g.Psi.tFormals))
	for i, tf := range g.Psi.tFormals {
		tfs[i] = TFormal{tf.name, tf.u_I.SubsEtaOpen(eta)}
	}
	ps := make([]ParamDecl, len(g.pDecls))
	for i, pd := range g.pDecls {
		ps[i] = ParamDecl{pd.name, pd.u.SubsEtaOpen(eta)}
	}
	u := g.u_ret.SubsEtaOpen(eta)
	return Sig{g.meth, BigPsi{tfs}, ps, u}
}

func (g Sig) Ok(ds []Decl, env Delta) {
	g.Psi.Ok(ds, env)
	for _, v := range g.Psi.tFormals {
		env[v.name] = v.u_I
	}
	seen := make(map[Name]ParamDecl)
	for _, v := range g.pDecls {
		if _, ok := seen[v.name]; ok {
			panic("Duplicate variable name " + v.name + ":\n\t" + g.String())
		}
		seen[v.name] = v
		v.u.Ok(ds, env)
	}
	g.u_ret.Ok(ds, env)
}

func (g Sig) GetSigs(_ []Decl) []Sig {
	return []Sig{g}
}

func (g Sig) String() string {
	var b strings.Builder
	b.WriteString(g.meth)
	b.WriteString(g.Psi.String())
	b.WriteString("(")
	writeParamDecls(&b, g.pDecls)
	b.WriteString(") ")
	b.WriteString(g.u_ret.String())
	return b.String()
}

/* Type Declaration */

type TypeDecl struct {
	name    Name
	Psi     BigPsi
	srcType Type
}

var _ Decl = TypeDecl{}

func (t TypeDecl) GetName() Name       { return t.name }
func (t TypeDecl) GetBigPsi() BigPsi   { return t.Psi }
func (t TypeDecl) GetSourceType() Type { return t.srcType }

func (t TypeDecl) Ok(ds []base.Decl) {
	// check type formals
	t.Psi.Ok(ds, make(Delta))
	delta := t.Psi.ToDelta()
	// check source type
	t.srcType.Ok(ds, delta)

	// check if the declaration has cycles (e.g. type A B; type B A)
	checkCyclicTypeDecl(ds, t, t.srcType)
}

func (t TypeDecl) String() string {
	var b strings.Builder
	b.WriteString("type ")
	b.WriteString(string(t.name))
	b.WriteString(t.Psi.String())
	b.WriteString(" ")
	b.WriteString(t.srcType.String())
	return b.String()
}

/* Aux, helpers */

/*func BigPsiOk(ds []Decl, env BigPsi, Psi BigPsi) {
	Psi.Ok(ds)
	delta := Psi.ToDelta()
	for _, v := range Psi.tFormals {
		u_I, _ := v.u_I.(TNamed) // \tau_I, already checked by psi.Ok
		u_I.Ok(ds, delta)        // !!! Submission version T-Type, t_i => t_I
	}
}*/

// For a type declaration decl, searches for any occurrence
// of decl.GetName() in the target type.
// If the target type is a struct/interface type,
// checks if its underlying type embeds decl.GetName().
func checkCyclicTypeDecl(ds []Decl, decl TypeDecl, target Type) {
	switch target := target.(type) {
	case TParam, TPrimitive:
		return
	case TNamed:
		if target.GetName() == decl.GetName() {
			panic("Invalid cyclic declaration: " + decl.String())
		}
		targetDecl := getTDecl(ds, target.GetName())
		checkCyclicTypeDecl(ds, decl, targetDecl.GetSourceType())

	case STypeLit:
		for _, f := range target.GetFieldDecls() {
			if u, ok := f.u.(TNamed); ok {
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

/* Version with fromStruct "optimization" */
//func checkCyclicTypeDecl(ds []Decl, decl TypeDecl, target Type, fromStruct bool) {
//	switch target := target.(type) {
//	case TParam, TPrimitive:
//		return
//	case TNamed:
//		if target.GetName() == decl.GetName() {
//			panic("Invalid cyclic declaration: " + decl.String())
//		}
//		targetDecl := getTDecl(ds, target.GetName())
//		checkCyclicTypeDecl(ds, decl, targetDecl.GetSourceType(), fromStruct)
//
//	case STypeLit:
//		for _, f := range target.GetFieldDecls() {
//			if u, ok := f.u.(TNamed); ok {
//				//if isStructType(ds, u) // CHECKME: without this check, the next call may be needlessly checking for cycles in u_I's
//				checkCyclicTypeDecl(ds, decl, u, true)
//			}
//		}
//	case ITypeLit:
//		if fromStruct {
//			return
//		}
//		for _, s := range target.GetSpecs() {
//			if u, ok := s.(TNamed); ok {
//				// u is a u_I, checked in Ok
//				checkCyclicTypeDecl(ds, decl, u, fromStruct)
//			}
//		}
//	}
//}

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
