package fg

import (
	"github.com/rhu1/fgg/internal/base"
	"reflect"
	"strings"
)

/* Export */

func NewTNamed(t Name) TNamed                  { return TNamed(t) }
func NewITypeLit(ss []Spec) ITypeLit           { return ITypeLit{ss} }
func NewSTypeLit(fds []FieldDecl) STypeLit     { return STypeLit{fds} }
func NewTPrimitive(t Tag) TPrimitive           { return TPrimitive{t} }
func NewUndefTPrimitive(t Tag) UndefTPrimitive { return UndefTPrimitive{t} }

// Factors t0 <: t_I for every Type t0, since the test is always the same.
// Pre: isInterfaceType(t_I)
//func Impls(ds []Decl, t0 Type, t_I Type) bool {
func Impls(ds []Decl, t0 Type, t_I ITypeLit) bool {
	ms0 := methods(ds, t0)
	msI := methods(ds, t_I)
	return ms0.IsSupersetOf(msI)
}

func EqualsOrImpls(ds []Decl, t0 Type, t Type) bool {
	if t0.Equals(t) {
		return true
	}
	if isInterfaceType(ds, t) {
		t_I := getInterface(ds, t)
		return Impls(ds, t0, t_I)
	}
	return false
}

/******************************************************************************/
/* Named (defined) types */

// Represents types declared/defined by the user
type TNamed Name

var _ Type = TNamed("")
var _ Spec = TNamed("")

func (t0 TNamed) GetName() Name { return Name(t0) }

func (t0 TNamed) AssignableTo(ds []Decl, t Type) (bool, Coercion) {

	if EqualsOrImpls(ds, t0, t) {
		return true, noOpCoercion
	}
	// if t is not a defined type
	if _, ok := t.(STypeLit); ok {
		if t0.Underlying(ds).Equals(t) {
			coercion := func(expr FGExpr) FGExpr {
				return Convert{t, expr}
			}
			return true, coercion
		}
	}
	return false, nil
}

func (t0 TNamed) Ok(ds []Decl) {
	getTDecl(ds, Name(t0)) // panics if decl not found
}

// t_I is a Spec, but not t_S -- this aspect is currently "dynamically typed"
// From Spec
func (t0 TNamed) GetSigs(ds []Decl) []Sig {
	t_I, ok := t0.Underlying(ds).(ITypeLit)
	if !ok {
		panic("Cannot use non-interface type as a Spec: " + t0.String())
	}
	var res []Sig
	for _, s := range t_I.specs {
		res = append(res, s.GetSigs(ds)...)
	}
	return res
}

func (t0 TNamed) Equals(t base.Type) bool {
	t_fg := asFGType(t)
	if _, ok := t_fg.(TNamed); !ok {
		return false
	}
	return t0 == t_fg.(TNamed)
}

func (t0 TNamed) String() string {
	return string(t0)
}

func (t0 TNamed) Underlying(ds []Decl) Type {
	td := getTDecl(ds, Name(t0))
	return td.GetSourceType().Underlying(ds)
}

/******************************************************************************/
/* Primitive types */

// Base interface for primitive (tagged) types,
// which may be defined or undefined.
type PrimType interface {
	Type
	Tag() Tag
}

/* Defined primitive types - int32, float32, string, etc. */

type TPrimitive struct {
	tag Tag
}

var _ Type = TPrimitive{}

func (t0 TPrimitive) Tag() Tag { return t0.tag }

func (t0 TPrimitive) Ok(ds []Decl) { /* nothing to check */ }

func (t0 TPrimitive) AssignableTo(ds []Decl, t Type) (bool, Coercion) {
	if EqualsOrImpls(ds, t0, t) {
		return true, noOpCoercion
	}
	return false, nil
}

func (t0 TPrimitive) Equals(t base.Type) bool {
	t_fg := asFGType(t)
	if _, ok := t_fg.(TPrimitive); !ok {
		return false
	}
	return t0 == t_fg.(TPrimitive)
}

func (t0 TPrimitive) String() string {
	return NameFromTag(t0.tag)
}

func (t0 TPrimitive) Underlying(ds []Decl) Type {
	return t0
}

/* Undefined primitive types - a type for Go's untyped constants */

type UndefTPrimitive struct {
	tag Tag
}

var _ Type = UndefTPrimitive{}

func (t0 UndefTPrimitive) Tag() Tag { return t0.tag }

func (t0 UndefTPrimitive) Ok(ds []Decl) { /* nothing to check */ }

func (t0 UndefTPrimitive) AssignableTo(ds []Decl, t Type) (bool, Coercion) {
	if EqualsOrImpls(ds, t0, t) {
		return true, noOpCoercion
	}
	if t0.RepresentableBy(ds, t) {
		coercion := func(expr FGExpr) FGExpr {
			return Convert{t, expr}
		}
		return true, coercion
	}
	return false, nil
}

func (t0 UndefTPrimitive) RepresentableBy(ds []Decl, t Type) bool {
	if t_P, ok := t.Underlying(ds).(PrimType); ok {
		return t0.fitsIn(t_P)
	}
	return false
}

func (t0 UndefTPrimitive) fitsIn(t_P PrimType) bool {
	switch {
	case isBool(t0) && isBool(t_P):
		return true
	case isString(t0) && isString(t_P):
		return true
	case isNumeric(t0) && isNumeric(t_P):
		return t0.Tag() <= t_P.Tag()
	default:
		return false
	}
}

func (t0 UndefTPrimitive) Equals(t base.Type) bool {
	t_fg := asFGType(t)
	if _, ok := t_fg.(UndefTPrimitive); !ok {
		return false
	}
	return t0 == t_fg.(UndefTPrimitive)
}

func (t0 UndefTPrimitive) String() string {
	return NameFromTag(t0.tag) + "(undefined)"
}

func (t0 UndefTPrimitive) Underlying(ds []Decl) Type {
	return t0
}

/******************************************************************************/
/* Struct literal */

type STypeLit struct {
	fDecls []FieldDecl
}

var _ Type = STypeLit{}

func (s STypeLit) GetFieldDecls() []FieldDecl { return s.fDecls }

func (s STypeLit) Ok(ds []Decl) {
	fs := make(map[Name]FieldDecl)
	for _, v := range s.fDecls {
		if _, ok := fs[v.name]; ok {
			panic("Multiple fields with name: " + v.name + "\n\t" + s.String())
		}
		fs[v.name] = v
		v.t.Ok(ds)
	}
}

func (s STypeLit) AssignableTo(ds []Decl, t Type) (bool, Coercion) {
	if EqualsOrImpls(ds, s, t) {
		return true, noOpCoercion
	}
	if s.Equals(t.Underlying(ds)) {
		coercion := func(expr FGExpr) FGExpr {
			return Convert{t, expr}
		}
		return true, coercion
	}
	return false, nil
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

func (s STypeLit) Underlying(ds []Decl) Type {
	return s
}

// Rename FDecl?
type FieldDecl struct {
	name Name
	t    Type
}

var _ FGNode = FieldDecl{}

func (f FieldDecl) GetType() Type { return f.t }

// From Decl
func (f FieldDecl) GetName() Name { return f.name }

func (fd FieldDecl) Equals(other FieldDecl) bool {
	return fd.name == other.name && fd.t.Equals(other.t)
}

func (fd FieldDecl) String() string {
	return fd.name + " " + fd.t.String()
}

/******************************************************************************/
/* Interface literal */

type ITypeLit struct {
	specs []Spec
}

var _ Type = ITypeLit{}

func (i ITypeLit) GetSpecs() []Spec { return i.specs }

func (i ITypeLit) Ok(ds []Decl) {
	seen := make(map[Name]Sig)
	for _, v := range i.specs {
		switch s := v.(type) {
		case Sig:
			if _, ok := seen[s.meth]; ok {
				panic("Multiple sigs with name: " + s.meth + "\n\t" + i.String())
			}
			seen[s.meth] = s
		case Type:
			if !isInterfaceType(ds, s) {
				panic("Embedded type must be an interface, not: " + s.String() +
					"\n\t" + s.String())
			}
		}
	}
}

func (i ITypeLit) AssignableTo(ds []Decl, t Type) (bool, Coercion) {
	if isInterfaceType(ds, t) {
		t_I := getInterface(ds, t)
		if Impls(ds, i, t_I) {
			return true, noOpCoercion
		}
	}
	return false, nil
}

func (i ITypeLit) Equals(t base.Type) bool {
	other, ok := t.(ITypeLit)
	if !ok {
		return false
	}
	// goal: methodSet(i) == methodSet(other), regardless of order
	// > this version is still sensible to order (todo)
	for idx, spec := range i.specs {
		if !specEquals(spec, other.specs[idx]) {
			return false
		}
	}
	return true
}

func specEquals(s1, s2 Spec) bool {
	switch s1 := s1.(type) {
	case TNamed:
		if named, ok := s2.(TNamed); ok {
			return s1.Equals(named)
		}
	case Sig:
		if g2, ok := s2.(Sig); ok {
			return s1.EqExceptVars(g2)
		}
	}
	return false
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

func (i ITypeLit) Underlying(ds []Decl) Type {
	return i
}

/******************************************************************************/
/* Aux */

// Cast to Type as defined in fg.go.
// Panics if cast fails.
func asFGType(t base.Type) Type {
	t_fg, ok := t.(Type)
	if !ok {
		panic("Expected FGR type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	return t_fg
}
