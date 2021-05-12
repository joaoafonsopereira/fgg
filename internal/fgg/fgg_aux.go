package fgg

import (
	"fmt"
	"reflect"
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

// Pre: len(s.psi.as) == len (u_S.typs), where s is the STypeLit decl for u_S.t
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
func methods(ds []Decl, u Type) map[Name]Sig { // CHECKME: deprecate?
	return methodsDelta(ds, make(Delta), u)
}

// TODO FIXME refactor
func MethodsDelta1(ds []Decl, delta Delta, u Type) map[Name]Sig {
	return methodsDelta(ds, delta, u)
}

func methodsDelta(ds []Decl, delta Delta, u Type) map[Name]Sig {
	switch u_cast := u.(type) {
	case ITypeLit:
		res := make(map[Name]Sig)
		for _, s := range u_cast.specs {
			switch s1 := s.(type) {
			case Sig:
				res[s1.meth] = s1
			case TNamed: // Embedded u_I
				for k, v := range methodsDelta(ds, delta, s1) { // cycles? (cf. submission version)
					res[k] = v
				}
			default:
				panic("Unknown Spec kind: " + reflect.TypeOf(s).String())
			}
		}
		return res

	case TNamed:
		// The method set of an interface type is its interface.
		// The method set of any other TNamed T consists of all methods
		// declared with receiver type T
		if u_I, ok := u_cast.Underlying(ds).(ITypeLit); ok {
			td := getTDecl(ds, u_cast.t_name)
			subs := MakeTSubs(td.Psi, u_cast.u_args)
			return methodsDelta(ds, delta, u_I.TSubs(subs))
		} else {
			res := make(map[Name]Sig)
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
		return map[Name]Sig{} // primitives don't implement any methods

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
