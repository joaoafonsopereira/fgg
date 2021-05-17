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

// MethodSet - aux type to represent the result of Methods.
// Makes it easier/more readable to check for superset relation
type MethodSet map[Name]Sig

func (m0 MethodSet) IsSupersetOf(m MethodSet) bool {
	for k, g := range m {
		g0, ok := m0[k]
		if !ok || !sigAlphaEquals(g0, g) {
			return false
		}
	}
	return true
}

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

// Pre: len(s.psi.as) == len (u_S.typs), where s is the STypeLit decl for u_S.t

// TODO IS THIS FUNCTION NEEDED? THE TYPE SUBS is already being applied in u_S.Underlying()
func fields(ds []Decl, u_S TNamed) []FieldDecl {
	s, ok := u_S.Underlying(ds).(STypeLit)
	if !ok {
		panic("Not a struct type: " + u_S.String())
	}

	decl := getTDecl(ds, u_S.t_name)
	subs := MakeTSubs(decl.Psi, u_S.u_args)
	fds := make([]FieldDecl, len(s.fDecls))
	for i := 0; i < len(s.fDecls); i++ {
		fds[i] = s.fDecls[i].TSubs(subs)
	}
	return fds
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
			td := getTDecl(ds, u_cast.t_name)
			subs := MakeTSubs(td.Psi, u_cast.u_args) // TODO problem: now this TSubs is already being done in Underlying. Where should it be??
			return methodsDelta(ds, delta, u_I.TSubs(subs))
		} else {
			res := make(MethodSet)
			for _, v := range ds {
				md, ok := v.(MethDecl)
				if ok && md.t_recv == u_cast.t_name {
					if ok, eta := MakeEtaDelta(ds, delta, md.Psi_recv, u_cast.u_args); ok {
						res[md.name] = md.ToSig().TSubs(eta)
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

// Pre: t_S is a struct type
// Submission version, m(~\rho) informal notation
//func body(ds []Decl, u_S TNamed, m Name, targs []Type) (Name, []Name, FGGExpr) {
func body(ds []Decl, u_S TNamed, m Name, targs []Type) (ParamDecl, []ParamDecl, FGGExpr) {
	for _, v := range ds {
		md, ok := v.(MethDecl)
		if ok && md.t_recv == u_S.t_name && md.name == m {
			subs := make(map[TParam]Type) // Cf. MakeEta
			for i := 0; i < len(md.Psi_recv.tFormals); i++ {
				subs[md.Psi_recv.tFormals[i].name] = u_S.u_args[i]
			}
			for i := 0; i < len(md.Psi_meth.tFormals); i++ {
				subs[md.Psi_meth.tFormals[i].name] = targs[i]
			}
			recv := ParamDecl{md.x_recv, u_S}
			pds := make([]ParamDecl, len(md.pDecls))
			for i := 0; i < len(md.pDecls); i++ {
				tmp := md.pDecls[i]
				pds[i] = ParamDecl{tmp.name, tmp.u.TSubs(subs)}
			}
			//return md.x_recv, xs, md.e_body.TSubs(subs)
			return recv, pds, md.e_body.TSubs(subs)
		}
	}
	panic("Method not found: " + u_S.String() + "." + m)
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

// !!! Sig in FGG includes ~a and ~x, which naively breaks "impls"
func sigAlphaEquals(g0 Sig, g Sig) bool {
	if len(g0.Psi.tFormals) != len(g.Psi.tFormals) || len(g0.pDecls) != len(g.pDecls) {
		return false
	}
	subs0 := makeParamIndexSubs(g0.Psi)
	subs := makeParamIndexSubs(g.Psi)
	for i := 0; i < len(g0.Psi.tFormals); i++ {
		if !g0.Psi.tFormals[i].u_I.TSubs(subs0).
			Equals(g.Psi.tFormals[i].u_I.TSubs(subs)) {
			//fmt.Println("z:")
			return false
		}
	}
	for i := 0; i < len(g0.pDecls); i++ {
		if !g0.pDecls[i].u.TSubs(subs0).Equals(g.pDecls[i].u.TSubs(subs)) {
			/*fmt.Println("w1: ", g0.pDecls[i].u, g0.pDecls[i].u.TSubs(subs0))
			fmt.Println("w2: ", g.pDecls[i].u, g.pDecls[i].u.TSubs(subs))
			fmt.Println("y:")*/
			return false
		}
	}
	/*fmt.Println("1:", g0)
	fmt.Println("2:", g)
	fmt.Println("3:", g0.meth == g.meth, g0.u_ret.Equals(g.u_ret))
	fmt.Println("4:", g0.u_ret.TSubs(subs0).Equals(g.u_ret.TSubs(subs)))*/
	return g0.meth == g.meth && g0.u_ret.TSubs(subs0).Equals(g.u_ret.TSubs(subs))
}

// CHECKME: Used by sigAlphaEquals, and MDecl.OK (for covariant receiver bounds)
func makeParamIndexSubs(Psi BigPsi) Delta {
	subs := make(Delta)
	for j := 0; j < len(Psi.tFormals); j++ {
		//subs[Psi.tFormals[j].name] = Psi.tFormals[j].name
		subs[Psi.tFormals[j].name] = TParam("α" + strconv.Itoa(j+1))
	}
	return subs
}