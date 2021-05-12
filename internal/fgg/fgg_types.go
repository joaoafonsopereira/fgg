package fgg

import (
	"github.com/rhu1/fgg/internal/base"
	"reflect"
	"strings"
)

type STypeLit struct {
	//t_name Name
	//Psi    BigPsi
	fDecls []FieldDecl
}

var _ Type = STypeLit{}

func (s STypeLit) GetFieldDecls() []FieldDecl { return s.fDecls }

func (s STypeLit) TSubs(subs map[TParam]Type) Type {
	fds := make([]FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		fds[i] = fd.Subs(subs)
	}
	return STypeLit{fds}
}

func (s STypeLit) SubsEta(eta Eta) TNamed {
	panic("implement me")
}

func (s STypeLit) SubsEtaOpen(eta EtaOpen) Type {
	panic("implement me")
}

func (s STypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	panic("implement me")
}

func (s STypeLit) Impls(ds []base.Decl, t base.Type) bool {
	panic("implement me")
}

type S struct {
	struct{i int}
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
	panic("implement me")
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

func (fd FieldDecl) Subs(subs map[TParam]Type) FieldDecl {
	return FieldDecl{fd.field, fd.u.TSubs(subs)}
}

func (fd FieldDecl) String() string {
	return fd.field + " " + fd.u.String()
}

////////////////////////////////////////////////////////////////////////////////

type ITypeLit struct {
	//t_I Name
	//Psi    BigPsi
	specs []Spec
}

var _ Type = ITypeLit{}

func (c ITypeLit) GetSpecs() []Spec  { return c.specs }

func (c ITypeLit) TSubs(subs map[TParam]Type) Type {
	panic("implement me")
}

func (c ITypeLit) SubsEta(eta Eta) TNamed {
	panic("implement me")
}

func (c ITypeLit) SubsEtaOpen(eta EtaOpen) Type {
	panic("implement me")
}

func (c ITypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	panic("implement me")
}

func (c ITypeLit) Impls(ds []base.Decl, t base.Type) bool {
	panic("implement me")
}

func (c ITypeLit) Ok(ds []Decl, delta Delta) {
	c.Psi.Ok(ds, PRIMITIVE_PSI)
	root := makeRootPsi(c.Psi)
	delta := root.ToDelta()
	seen_g := make(map[Name]Sig)    // !!! unique(~S) more flexible
	seen_u := make(map[string]Type) // key is u.String()
	for _, v := range c.specs {
		switch s := v.(type) {
		case Sig:
			if _, ok := seen_g[s.meth]; ok {
				panic("Multiple sigs with name: " + s.meth + "\n\t" + c.String())
			}
			seen_g[s.meth] = s
			s.Ok(ds, root)
		case TNamed:
			k := s.String()
			if _, ok := seen_u[k]; ok {
				panic("Repeat embedding of type: " + k + "\n\t" + c.String())
			}
			seen_u[k] = s
			if !IsNamedIfaceType(ds, s) { // CHECKME: allow embed type param?
				panic("Embedded type must be a named interface, not: " + k + "\n\t" + c.String())
			}
			s.Ok(ds, delta)
			if isRecursiveInterfaceEmbedding(ds, make(map[string]TNamed), s) {
				panic("Invalid recursive interface embedding type:\n\t" + c.String())
			}
		default:
			panic("Unknown Spec kind: " + reflect.TypeOf(v).String() + "\n\t" +
				c.String())
		}
	}
}

func (c ITypeLit) Equals(t base.Type) bool {
	panic("implement me")
}

func (c ITypeLit) String() string {
	var b strings.Builder
	b.WriteString(" interface {")
	if len(c.specs) > 0 {
		b.WriteString(" ")
		b.WriteString(c.specs[0].String())
		for _, v := range c.specs[1:] {
			b.WriteString("; ")
			b.WriteString(v.String())
		}
		b.WriteString(" ")
	}
	b.WriteString("}")
	return b.String()
}

func (c ITypeLit) ToGoString(ds []Decl) string {
	panic("implement me")
}

func (c ITypeLit) Underlying(ds []Decl) Type {
	return c
}