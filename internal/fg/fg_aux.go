package fg

import "fmt"

var _ = fmt.Errorf

// MethodSet - aux type to represent the result of Methods.
// Makes it easier/more readable to check for superset relation
type MethodSet map[Name]Sig

func (m0 MethodSet) IsSupersetOf(m MethodSet) bool {
	for k, sig := range m {
		sig0, ok := m0[k]
		if !ok || !sig.EqExceptVars(sig0) {
			return false
		}
	}
	return true
}

/* fields(t_S), methods(t), body(t_S, m) */

// Pre: t_S is a struct type
func fields(ds []Decl, t_S Type) []FieldDecl {
	s, ok := t_S.Underlying(ds).(STypeLit)
	if !ok {
		panic("Not a struct type: " + t_S.String())
	}
	return s.fDecls
}

// Go has no overloading, meth names are a unique key
func methods(ds []Decl, t Type) MethodSet {
	switch t_cast := t.(type) {
	case ITypeLit:
		res := make(MethodSet)
		for _, s := range t_cast.specs {
			for _, v := range s.GetSigs(ds) { // cycles? (cf. submission version, recursive "methods")
				res[v.meth] = v
			}
		}
		return res

	case TNamed:
		// The method set of an interface type is its interface.
		// The method set of any other TNamed T consists of all methods
		// declared with receiver type T
		if t_I, ok := t_cast.Underlying(ds).(ITypeLit); ok {
			return methods(ds, t_I)
		} else {
			res := make(MethodSet)
			for _, v := range ds {
				md, ok := v.(MethDecl)
				if ok && md.recv.t == t_cast {
					res[md.name] = md.ToSig()
				}
			}
			return res
		}
	case TPrimitive, UndefTPrimitive, STypeLit:
		return MethodSet{} // primitives don't implement any methods
	default:
		panic("Unknown type: " + t.String()) // Perhaps redundant if all TDecl OK checked first
	}
}

// Pre: t_S is a concrete type
func body(ds []Decl, t_S TNamed, m Name) (Name, []Name, FGExpr) {
	md := getMethDecl(ds, t_S, m) // panics if not found
	xs := make([]Name, len(md.pDecls))
	for i := 0; i < len(md.pDecls); i++ {
		xs[i] = md.pDecls[i].name
	}
	return md.recv.name, xs, md.e_body
}

// Represents the aux function type() defined in fig.16 of the paper.
// Returns the exact run-time type of a value expression.
func concreteType(e FGExpr) Type {
	switch e1 := e.(type) {
	case StructLit:
		return e1.t_S
	case TypedPrimitiveValue:
		return e1.typ
	case PrimitiveLiteral:
		return UndefTPrimitive{e1.tag}
	}
	panic("concreteType: expression is not a value: " + e.String())
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

func getMethDecl(ds []Decl, recv Type, m Name) MethDecl {
	for _, d := range ds {
		md, ok := d.(MethDecl)
		if ok && md.recv.t == recv && md.name == m {
			return md
		}
	}
	panic("Method not found: " + recv.String() + "." + m)
}
