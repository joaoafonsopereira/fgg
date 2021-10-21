package fgg

import (
	"fmt"
	"strconv"
)

var _ = fmt.Errorf

/* Export */

func Bounds(delta Delta, u Type) Type          { return bounds(delta, u) }
func Fields(ds []Decl, u_S TNamed) []FieldDecl { return fields(ds, u_S) }
func Methods(ds []Decl, u Type) map[Name]Sig   { return methods(ds, u) }
func GetTDecl(ds []Decl, t Name) TypeDecl      { return getTDecl(ds, t) }

/* bounds(delta, u), fields(u_S), methods(u), body(u_S, m) */

// return type TName?
func bounds(delta Delta, u Type) Type {
	if a, ok := u.(TParam); ok {
		if res, ok := delta[a]; ok {
			return res
		}
	}
	return u // CHECKME: submission version, includes when TParam 'a' not in delta, correct?
}

func fields(ds []Decl, u_S TNamed) []FieldDecl {
	s, ok := u_S.Underlying(ds).(STypeLit)
	if !ok {
		panic("Not a struct type: " + u_S.String())
	}
	return s.GetFieldDecls()
}

// Go has no overloading, meth names are a unique key
func methods(ds []Decl, u Type) MethodSet { // CHECKME: deprecate?
	return methodsDelta(ds, make(Delta), u)
}

// TODO FIXME refactor
func MethodsDelta1(ds []Decl, delta Delta, u Type) MethodSet {
	return methodsDelta(ds, delta, u)
}

func methodsDelta(ds []Decl, delta Delta, u Type) MethodSet {
	switch u_cast := u.(type) {
	case ITypeLit:
		res := make(MethodSet)
		for _, s := range u_cast.specs {
			for _, v := range s.GetSigs(ds) {
				res[v.meth] = v
			}
		}
		return res

	case TNamed:
		// The method set of an interface type is its interface.
		// The method set of any other TNamed T consists of all methods
		// declared with receiver type T
		if u_I, ok := u_cast.Underlying(ds).(ITypeLit); ok {
			return methodsDelta(ds, delta, u_I)
		} else {
			res := make(MethodSet)
			for _, v := range ds {
				md, ok := v.(MethDecl)
				if ok && md.t_recv == u_cast.t_name {
					if ok, eta := MakeEtaDelta(ds, delta, md.Psi_recv, u_cast.u_args); ok {
						res[md.name] = md.ToSig().SubsEtaOpen(eta)
					}
				}
			}
			return res
		}

	case TParam:
		upper, ok := delta[u_cast]
		if !ok {
			panic("TParam: " + u.String() + " not in env: " + delta.String())
		}
		//return methodsDelta(ds, delta, bounds(delta, u_cast)) // !!! delegate to bounds
		return methodsDelta(ds, delta, upper)

	case TPrimitive, STypeLit:
		return MethodSet{} // primitives don't implement any methods

	default:
		panic("Unknown type: " + u.String()) // Perhaps redundant if all TDecl OK checked first
	}
}

// Pre: t_S is a concrete type
// Submission version, m(~\rho) informal notation
//func body(ds []Decl, u_S TNamed, m Name, targs []Type) (Name, []Name, FGGExpr) {
func body(ds []Decl, u_S TNamed, m Name, targs []Type) (ParamDecl, []ParamDecl, FGGExpr) {
	md := getMethDecl(ds, u_S.t_name, m) // panics if not found
	theta := MakeEtaOpen(md.Psi_recv, u_S.u_args) // cf MakeEta
	for i := 0; i < len(md.Psi_meth.tFormals); i++ { //TODO TSubs.add() ..., TSubs vs SubsEtaOpen
		theta[md.Psi_meth.tFormals[i].name] = targs[i]
	}
	recv := ParamDecl{md.x_recv, u_S}
	pds := make([]ParamDecl, len(md.pDecls))
	for i := 0; i < len(md.pDecls); i++ {
		tmp := md.pDecls[i]
		pds[i] = ParamDecl{tmp.name, tmp.u.SubsEtaOpen(theta)}
	}
	//return md.x_recv, xs, md.e_body.TSubs(subs)
	return recv, pds, md.e_body.TSubs(theta)
}

// Represents the aux function type() defined in fig.16 of the paper.
// Returns the exact run-time type of a value expression.
func dynamicType(e FGGExpr) Type {
	switch e1 := e.(type) {
	case StructLit:
		return e1.u_S
	case NamedPrimitiveLiteral:
		return e1.typ
	case PrimtValue:
		panic("dynamicType(PrimtValue) not defined") // todo <<<<--------------------
	}
	panic("dynamicType: expression is not a value: " + e.String())
}

/* MethodSet; alpha-equality of signatures */

// MethodSet - aux type to represent the result of Methods.
// Makes it easier/more readable to check for superset relation
type MethodSet map[Name]Sig

func (m0 MethodSet) IsSupersetOf(m MethodSet) bool {
	for name, g := range m {
		g0, ok := m0[name]
		if !ok || !sigAlphaEquals(g0, g) {
			return false
		}
	}
	return true
}

// !!! Sig in FGG includes ~a and ~x, which naively breaks "impls"
func sigAlphaEquals(g0 Sig, g1 Sig) bool {
	if len(g0.Psi.tFormals) != len(g1.Psi.tFormals) || len(g0.pDecls) != len(g1.pDecls) {
		return false
	}
	subs0 := makeParamIndexSubs(g0.Psi)
	subs1 := makeParamIndexSubs(g1.Psi)
	sig0 := g0.SubsEtaOpen(subs0)
	sig1 := g1.SubsEtaOpen(subs1)

	for i := 0; i < len(sig0.Psi.tFormals); i++ {
		if !sig0.Psi.tFormals[i].u_I.Equals(sig1.Psi.tFormals[i].u_I) {
			return false
		}
	}
	for i := 0; i < len(sig0.pDecls); i++ {
		if !sig0.pDecls[i].u.Equals(sig1.pDecls[i].u) {
			return false
		}
	}
	return sig0.meth == sig1.meth && sig0.u_ret.Equals(sig1.u_ret)
}

// CHECKME: Used by sigAlphaEquals, and MDecl.OK (for covariant receiver bounds)
func makeParamIndexSubs(Psi BigPsi) EtaOpen {
	subs := make(EtaOpen)
	for j := 0; j < len(Psi.tFormals); j++ {
		//subs[Psi.tFormals[j].name] = Psi.tFormals[j].name
		subs[Psi.tFormals[j].name] = TParam("α" + strconv.Itoa(j+1))
	}
	return subs
}

/* Additional */

func getTDecl(ds []Decl, t Name) TypeDecl {
	for _, v := range ds {
		td, ok := v.(TypeDecl)
		if ok && td.GetName() == t {
			return td
		}
	}
	panic("Type not found: " + t)
}

func getMethDecl(ds []Decl, recv Name, m Name) MethDecl {
	for _, d := range ds {
		md, ok := d.(MethDecl)
		if ok && md.t_recv == recv && md.name == m {
			return md
		}
	}
	panic("Method not found: " + recv + "." + m)
}