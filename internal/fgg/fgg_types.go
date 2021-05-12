package fgg

import (
	"github.com/rhu1/fgg/internal/base"
	"reflect"
	"strings"
)

type STypeLit struct {
	fDecls []FieldDecl
}

var _ Type = STypeLit{}

func (s STypeLit) GetFieldDecls() []FieldDecl { return s.fDecls }

func (s STypeLit) TSubs(subs map[TParam]Type) Type {
	fds := make([]FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		fds[i] = fd.TSubs(subs)
	}
	return STypeLit{fds}
}

func (s STypeLit) SubsEta(eta Eta) TNamed {
	//fds := make([]FieldDecl, len(s.fDecls))
	//for i, fd := range s.fDecls {
	//	fds[i] = fd.SubsEta(eta)
	//}
	//return STypeLit{fds}
	panic("STypeLit.SubsEta") // TODO how to return a TNamed?? Maybe should return some kind of GroundType ??
}

func (s STypeLit) SubsEtaOpen(eta EtaOpen) Type {
	fds := make([]FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		fds[i] = fd.SubsEtaOpen(eta)
	}
	return STypeLit{fds}
}

func (s STypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	return s.Equals(u.Underlying(ds))
}

func (s STypeLit) Impls(ds []base.Decl, t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGG type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	return s.ImplsDelta(ds, make(Delta), t.(Type))
}

func (s STypeLit) Ok(ds []Decl, delta Delta) {
	seen := make(map[Name]FieldDecl)
	for _, v := range s.fDecls {
		if _, ok := seen[v.field]; ok {
			panic("Duplicate field name: " + v.field + "\n\t" + s.String())
		}
		seen[v.field] = v
		v.u.Ok(ds, delta)
	}
	u_S := TNamed{s.t_name, s.Psi.Hat()}
	if isRecursiveFieldType(ds, make(map[string]TNamed), u_S) {
		panic("Invalid recursive struct type:\n\t" + s.String())
	}
}

func (s STypeLit) Equals(t base.Type) bool {
	other, ok := t.(STypeLit)
	if !ok {
		return false
	}
	if len(s.fDecls) != len(other.fDecls) {
		return false
	}
	for i, fd := range s.fDecls {
		if !fd.Equals(other.fDecls[i]) {
			return false
		}
	}
	return true
}

func (s STypeLit) String() string {
	var b strings.Builder
	b.WriteString(" struct {")
	if len(s.fDecls) > 0 {
		b.WriteString(" ")
		writeFieldDecls(&b, s.fDecls)
		b.WriteString(" ")
	}
	b.WriteString("}")
	return b.String()
}

func (s STypeLit) ToGoString(ds []Decl) string {
	panic("implement me")
}

func (s STypeLit) Underlying(ds []Decl) Type {
	return s
}


type FieldDecl struct {
	field Name
	u     Type // u=tau
}

var _ FGGNode = FieldDecl{}

func (fd FieldDecl) GetName() Name { return fd.field }
func (fd FieldDecl) GetType() Type { return fd.u }

// TODO these 3 look too similar -- possible refactor?
func (fd FieldDecl) TSubs(subs map[TParam]Type) FieldDecl {
	return FieldDecl{fd.field, fd.u.TSubs(subs)}
}
func (fd FieldDecl) SubsEta(eta Eta) FieldDecl {
	return FieldDecl{fd.field, fd.u.SubsEta(eta)}
}
func (fd FieldDecl) SubsEtaOpen(eta EtaOpen) FieldDecl {
	return FieldDecl{fd.field, fd.u.SubsEtaOpen(eta)}
}

func (fd FieldDecl) Equals(other FieldDecl) bool {
	return fd.field == other.field && fd.u.Equals(other.u)
}

func (fd FieldDecl) String() string {
	return fd.field + " " + fd.u.String()
}

////////////////////////////////////////////////////////////////////////////////

type ITypeLit struct {
	specs []Spec
}

var _ Type = ITypeLit{}

func (i ITypeLit) GetSpecs() []Spec { return i.specs }

func (i ITypeLit) TSubs(subs map[TParam]Type) Type {
	specs := make([]Spec, len(i.specs))
	for i, spec := range i.specs {
		switch spec := spec.(type) {
		case Sig:
			specs[i] = spec.TSubs(subs)
		case TNamed:
			specs[i] = spec.TSubs(subs).(TNamed)
		}
	}
	return ITypeLit{specs}
}

func (i ITypeLit) SubsEta(eta Eta) TNamed {
	panic("implement me")

}

func (i ITypeLit) SubsEtaOpen(eta EtaOpen) Type {
	panic("implement me ITypeLit.SubsEtaOpen")
	//specs := make([]Spec, len(i.specs))
	//for i, spec := range i.specs {
	//	switch spec := spec.(type) {
	//	case Sig:
	//		specs[i] = spec.SubsE(subs)
	//	case TNamed:
	//		specs[i] = spec.SubsEtaOpen(eta).(TNamed)
	//	}
	//}
	//return ITypeLit{specs}
}

func (i ITypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	if isIfaceType(ds, u) {
		return false
	}

	gs := methodsDelta(ds, delta, u)   // u is a t_I
	gs0 := methodsDelta(ds, delta, i)
	for k, g := range gs {
		g0, ok := gs0[k]
		if !ok || !sigAlphaEquals(g0, g) {
			return false
		}
	}
	return true
}

func (i ITypeLit) Impls(ds []base.Decl, t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGG type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	return i.ImplsDelta(ds, make(Delta), t.(Type))
}

// Pre: delta.Ok
func (i ITypeLit) Ok(ds []Decl, delta Delta) {
	seen_g := make(map[Name]Sig)    // !!! unique(~S) more flexible
	seen_u := make(map[string]Type) // key is u.String()
	for _, v := range i.specs {
		switch s := v.(type) {
		case Sig:
			if _, ok := seen_g[s.meth]; ok {
				panic("Multiple sigs with name: " + s.meth + "\n\t" + i.String())
			}
			seen_g[s.meth] = s
			s.Ok(ds, root)
		case TNamed:
			k := s.String()
			if _, ok := seen_u[k]; ok {
				panic("Repeat embedding of type: " + k + "\n\t" + i.String())
			}
			seen_u[k] = s
			if !IsNamedIfaceType(ds, s) { // CHECKME: allow embed type param?
				panic("Embedded type must be a named interface, not: " + k + "\n\t" + i.String())
			}
			s.Ok(ds, delta)
			if isRecursiveInterfaceEmbedding(ds, make(map[string]TNamed), s) {
				panic("Invalid recursive interface embedding type:\n\t" + c.String())
			}
		default:
			panic("Unknown Spec kind: " + reflect.TypeOf(v).String() + "\n\t" +
				i.String())
		}
	}
}

func (i ITypeLit) Equals(t base.Type) bool {
	panic("implement me")
}

func (i ITypeLit) String() string {
	var b strings.Builder
	b.WriteString(" interface {")
	if len(i.specs) > 0 {
		b.WriteString(" ")
		b.WriteString(i.specs[0].String())
		for _, v := range i.specs[1:] {
			b.WriteString("; ")
			b.WriteString(v.String())
		}
		b.WriteString(" ")
	}
	b.WriteString("}")
	return b.String()
}

func (i ITypeLit) ToGoString(ds []Decl) string {
	panic("implement me")
}

func (i ITypeLit) Underlying(ds []Decl) Type {
	return i
}