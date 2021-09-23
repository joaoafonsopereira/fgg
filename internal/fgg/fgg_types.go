package fgg

import (
	"github.com/rhu1/fgg/internal/base"
	"reflect"
	"strings"
)

/* Exports */

func NewTParam(name Name) TParam                      { return TParam(name) }
func NewTNamed(t Name, us []Type) TNamed              { return TNamed{t, us} }
func NewTPrimitive(t Tag) TPrimitive                  { return TPrimitive{t, false} }
func NewUndefTPrimitive(t Tag) TPrimitive             { return TPrimitive{t, true} }
func NewSTypeLit(fds []FieldDecl) STypeLit            { return STypeLit{fds} }
func NewITypeLit(specs []Spec, tlist []Type) ITypeLit { return ITypeLit{specs, tlist} }

func NewTypeList(tlist []Type) TypeList { return tlist }

/* Type parameters */

type TParam Name

var _ Type = TParam("")

func (a TParam) SubsEtaOpen(eta EtaOpen) Type {
	res, ok := eta[a]
	if !ok {
		return a
	}
	return res
}

func (a TParam) SubsEtaClosed(eta EtaClosed) GroundType {
	res, ok := eta[a]
	if !ok {
		panic("Shouldn't get here: " + a)
	}
	return res
}

// u0 <: u
func (a TParam) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	if a1, ok := u.(TParam); ok {
		return a == a1
	} else if isIfaceLikeType(ds, u) { // todo review this -- which types can a TParam implement?
		//return bounds(delta, a).ImplsDelta(ds, delta, u) // !!! more efficient?
		gs0 := methodsDelta(ds, delta, a)
		gs := methodsDelta(ds, delta, u)
		for k, g := range gs {
			g0, ok := gs0[k]
			//if !ok || !sigAlphaEquals(g0, g) {
			if !ok || g0.String() != g.String() {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

// Cf. base.Type
func (a TParam) Impls(ds []Decl, t base.Type) bool {
	u := asFGGType(t)
	return a.ImplsDelta(ds, make(Delta), u)
}

func (a TParam) Ok(ds []Decl, delta Delta) {
	if _, ok := delta[a]; !ok {
		panic("Type param " + a.String() + " unknown in context: " + delta.String())
	}
}

func (a TParam) Equals(t base.Type) bool {
	u := asFGGType(t)
	if b, ok := u.(TParam); ok {
		return a == b
	}
	return false
}

func (a TParam) String() string {
	return string(a)
}

func (a TParam) ToGoString(ds []Decl) string {
	return string(a)
}

func (a TParam) Underlying(ds []Decl) Type {
	return a
}

/******************************************************************************/
/* Named (defined) types -- the only 'kind' that may take type parameters */

// Convention: t=type name (t), u=FGG type (tau)
type TNamed struct {
	t_name Name
	u_args []Type // SmallPsi
}

var _ Type = TNamed{}
var _ Spec = TNamed{}

func (u0 TNamed) GetName() Name    { return u0.t_name }
func (u0 TNamed) GetTArgs() []Type { return u0.u_args } // SmallPsi

func (u0 TNamed) SubsEtaOpen(eta EtaOpen) Type {
	//fmt.Println("555:", u0, eta)
	us := make([]Type, len(u0.u_args))
	for i := 0; i < len(us); i++ {
		us[i] = u0.u_args[i].SubsEtaOpen(eta)
	}
	return TNamed{u0.t_name, us}
}

func (u0 TNamed) SubsEtaClosed(eta EtaClosed) GroundType {
	//fmt.Println("555:", u0, eta)
	us := make([]Type, len(u0.u_args))
	for i := 0; i < len(us); i++ {
		us[i] = u0.u_args[i].SubsEtaClosed(eta)
	}
	return TNamed{u0.t_name, us}
}

// u0 <: u
// delta unused here (cf. TParam.ImplsDelta)
func (u0 TNamed) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	// TODO Impls seems to be overloaded: it acts both as Implements and IsAssignableTo
	switch u := u.(type) {
	case TParam: // e.g., fgg_test.go, Test014 TODO revise this
		panic("Type name does not implement open type param: found=" +
			u0.String() + ", expected=" + u.String())
	case TPrimitive:
		return false
	case STypeLit: // or any other composite type literal, if there were more
		return u0.Underlying(ds).Equals(u)

	case TNamed:
		if u0.Equals(u) {
			return true
		} else if u_I, ok := u.Underlying(ds).(ITypeLit); ok {
			return u0.ImplsDelta(ds, delta, u_I) // falls into the case below (ITypeLit)
		} else {
			return false
		}
	case ITypeLit:
		if u.HasTList() {
			tlist := u.GetTList(ds)
			if !(tlist.Contains(u0) || tlist.Contains(u0.Underlying(ds))) { // TODO initial version; not accounting for https://github.com/golang/go/issues/45346
				return false
			}
		}
		gs := methodsDelta(ds, delta, u)   // u is a t_I
		gs0 := methodsDelta(ds, delta, u0) // t0 may be any
		return gs0.IsSupersetOf(gs)
	default:
		panic("Unknown type: " + u.String())
	}
}

// Cf. base.Type
func (u0 TNamed) Impls(ds []Decl, u base.Type) bool {
	u_fgg := asFGGType(u)
	return u0.ImplsDelta(ds, make(Delta), u_fgg)
}

func (u0 TNamed) Ok(ds []Decl, delta Delta) {
	td := getTDecl(ds, u0.t_name) // Panics if type not found
	Psi := td.GetBigPsi()
	if len(Psi.tFormals) != len(u0.u_args) {
		var b strings.Builder
		b.WriteString("Arity mismatch between type formals and actuals: formals=")
		b.WriteString(Psi.String())
		b.WriteString(" actuals=")
		writeTypes(&b, u0.u_args)
		b.WriteString("\n\t")
		b.WriteString(u0.String())
		panic(b.String())
	}
	for _, v := range u0.u_args {
		v.Ok(ds, delta)
	}
	// Duplicates MakeEtaDelta so as to pinpoint wrong (\alpha, u_I) pair
	eta := MakeEtaOpen(Psi, u0.u_args)
	for _, tf := range Psi.tFormals {
		actual := tf.name.SubsEtaOpen(eta)
		// CHECKME: submission T-Named, subs applied to Delta? -- already applied, Delta is coming from the subs context
		formal := tf.u_I.SubsEtaOpen(eta)
		if !actual.ImplsDelta(ds, delta, formal) { // formal is a \tau_I, checked by TDecl.Ok
			panic("Type actual must implement type formal: actual=" +
				actual.String() + " formal=" + formal.String())
		}
	}
}

// \tau_I is a Spec, but not \tau_S -- this aspect is currently "dynamically typed"
// From Spec
func (u TNamed) GetSigs(ds []Decl) []Sig {
	u_I, ok := u.Underlying(ds).(ITypeLit)
	if !ok {
		panic("Cannot use non-interface type as a Spec: " + u.String() +
			" is a " + reflect.TypeOf(u).String())
	}
	var res []Sig
	for _, s := range u_I.specs {
		res = append(res, s.GetSigs(ds)...)
	}
	return res
}

func (u0 TNamed) Equals(t base.Type) bool {
	u := asFGGType(t)
	if _, ok := u.(TNamed); !ok {
		return false
	}
	u1 := u.(TNamed)
	if u0.t_name != u1.t_name || len(u0.u_args) != len(u1.u_args) {
		return false
	}
	for i := 0; i < len(u0.u_args); i++ {
		if !u0.u_args[i].Equals(u1.u_args[i]) { // Asks equality of nested TParam
			return false
		}
	}
	return true
}

func (u TNamed) String() string {
	var b strings.Builder
	b.WriteString(string(u.t_name))
	b.WriteString("(")
	writeTypes(&b, u.u_args)
	b.WriteString(")")
	return b.String()
}

func (u TNamed) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString("main.")
	b.WriteString(string(u.t_name))
	b.WriteString("(")
	writeToGoTypes(ds, &b, u.u_args)
	b.WriteString(")")
	return b.String()
}

func (u TNamed) Underlying(ds []Decl) Type {
	decl := getTDecl(ds, u.t_name)
	under := decl.GetSourceType().Underlying(ds)
	// the underlying type itself may have type variables, as in e.g.
	// type S[T any] struct { x T }
	//  -> the underlying of S[int] is struct { x int }, hence the TSubs
	subs := MakeEtaOpen(decl.Psi, u.u_args) // EtaOpen <-> Eta (may be closed or not, we don't know)
	return under.SubsEtaOpen(subs)
}

/******************************************************************************/
/* Primitive types */

type TPrimitive struct {
	tag       Tag
	undefined bool
}

var _ Type = TPrimitive{}

func (t TPrimitive) Tag() Tag        { return t.tag }
func (t TPrimitive) Undefined() bool { return t.undefined }

func (t TPrimitive) SubsEtaOpen(eta EtaOpen) Type {
	return t
}

func (t TPrimitive) SubsEtaClosed(eta EtaClosed) GroundType {
	return t
}

func (t0 TPrimitive) ImplsDelta(ds []Decl, delta Delta, u Type) bool {

	if t0.Undefined() {
		return t0.canConvertTo(ds, delta, u)
	}

	switch u_cast := u.(type) {
	case TPrimitive:
		return t0.Equals(u_cast)
	case TNamed:
		return isIfaceType(ds, u) && t0.ImplsDelta(ds, delta, u.Underlying(ds))
	case ITypeLit:
		if u_cast.HasTList() {
			if !(u_cast.GetTList(ds).Contains(t0)) {
				return false
			}
		}
		return len(methods(ds, u_cast)) == 0
	default: // TParam, STypeLit, ...
		return false
	}
}

func (t0 TPrimitive) canConvertTo(ds []Decl, delta Delta, u Type) bool {
	if !t0.Undefined() {
		panic("shouldn't get here")
	}
	switch under := u.Underlying(ds).(type) {
	case TPrimitive:
		return t0.fitsIn(under)
	case TParam:
		constraint := bounds(delta, u)
		return t0.canConvertTo(ds, delta, constraint) // falls into case below
	case ITypeLit:
		if under.HasTList() {
			for _, u2 := range under.GetTList(ds) {
				if !t0.canConvertTo(ds, delta, u2) {
					return false
				}
			}
		}
		return len(methods(ds, under)) == 0
	}
	return false
}

func (t0 TPrimitive) fitsIn(t TPrimitive) bool {
	if !t0.Undefined() {
		panic("shouldn't get here")
	}
	if t0.tag > t.tag { // necessary check??
		return false
	}
	switch t0.tag {
	case BOOL:
		return t.tag == BOOL
	case STRING:
		return t.tag == STRING
	case INT32, INT64:
		return INT32 <= t.tag && t.tag <= FLOAT64 // kind of ad-hoc
	case FLOAT32, FLOAT64:
		return FLOAT32 <= t.tag && t.tag <= FLOAT64
	default:
		panic("FitsIn: t has unsupported type: " + t.String())
	}
}

func (t0 TPrimitive) Impls(ds []base.Decl, t base.Type) bool {
	t_fgg := asFGGType(t)
	return t0.ImplsDelta(ds, make(Delta), t_fgg)
}

func (t TPrimitive) Ok(ds []Decl, delta Delta) {
	// do nothing -- a primitive type is always Ok
}

func (t0 TPrimitive) Equals(t base.Type) bool {
	u := asFGGType(t)
	if _, ok := u.(TPrimitive); !ok {
		return false
	}
	return t0 == u.(TPrimitive)
}

func (t TPrimitive) String() string {
	undef := ""
	if t.undefined {
		undef = "(undefined)"
	}
	return NameFromTag(t.tag) + undef
}

func (t TPrimitive) ToGoString(ds []Decl) string {
	return t.String()
}

func (t TPrimitive) Underlying(ds []Decl) Type {
	return t
}

/******************************************************************************/
/* Struct literal */

type STypeLit struct {
	fDecls []FieldDecl
}

var _ Type = STypeLit{}

func (s STypeLit) GetFieldDecls() []FieldDecl { return s.fDecls }

func (s STypeLit) SubsEtaOpen(eta EtaOpen) Type {
	fds := make([]FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		fds[i] = fd.SubsEtaOpen(eta)
	}
	return STypeLit{fds}
}

func (s STypeLit) SubsEtaClosed(eta EtaClosed) GroundType {
	fds := make([]FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		fds[i] = fd.SubsEtaClosed(eta)
	}
	return STypeLit{fds}
}

func (s STypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	return s.Equals(u.Underlying(ds))
}

func (s STypeLit) Impls(ds []base.Decl, t base.Type) bool {
	u := asFGGType(t)
	return s.ImplsDelta(ds, make(Delta), u)
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

func (fd FieldDecl) SubsEtaOpen(eta EtaOpen) FieldDecl {
	return FieldDecl{fd.field, fd.u.SubsEtaOpen(eta)}
}

func (fd FieldDecl) SubsEtaClosed(eta EtaClosed) FieldDecl {
	return FieldDecl{fd.field, fd.u.SubsEtaClosed(eta)}
}

func (fd FieldDecl) Equals(other FieldDecl) bool {
	return fd.field == other.field && fd.u.Equals(other.u)
}

func (fd FieldDecl) String() string {
	return fd.field + " " + fd.u.String()
}

/******************************************************************************/
/* Interface literal */

type ITypeLit struct {
	specs []Spec
	tlist TypeList
}

var _ Type = ITypeLit{}

func (i ITypeLit) GetSpecs() []Spec { return i.specs }
func (i ITypeLit) HasTList() bool {
	return i.tlist != nil && len(i.tlist) > 0
}

// When a constraint embeds another constraint, the type list of
// the final constraint is the intersection of all the type lists involved.
// If there are multiple embedded types, intersection preserves the property
// that any type argument must satisfy the requirements of all embedded types.
func (i ITypeLit) GetTList(ds []Decl) TypeList {
	res := i.tlist
	for _, spec := range i.specs {
		if emb, ok := spec.(TNamed); ok {
			emb_under := emb.Underlying(ds).(ITypeLit) // checked in ok
			if emb_under.HasTList() {
				res = res.intersect(emb_under.GetTList(ds))
			}
		}
	}
	return res
}

func (i ITypeLit) SubsEtaClosed(eta EtaClosed) GroundType {
	specs := make([]Spec, len(i.specs))
	for i, spec := range i.specs {
		switch s := spec.(type) {
		case Sig:
			// eta won't contain mappings for the type vars of this Sig,
			// hence it isn't really 'closed' in this context
			subs := eta.ToEtaOpen()
			specs[i] = s.SubsEtaOpen(subs)
		case TNamed:
			specs[i] = s.SubsEtaClosed(eta).(TNamed)
		}
	}
	return ITypeLit{specs, i.tlist}
}

func (i ITypeLit) SubsEtaOpen(eta EtaOpen) Type {
	specs := make([]Spec, len(i.specs))
	for i, spec := range i.specs {
		switch s := spec.(type) {
		case Sig:
			specs[i] = s.SubsEtaOpen(eta)
		case TNamed:
			specs[i] = s.SubsEtaOpen(eta).(TNamed)
		}
	}
	return ITypeLit{specs, i.tlist}
}

func (i ITypeLit) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	if isIfaceType(ds, u) {
		return false
	}
	gs := methodsDelta(ds, delta, u) // u is a t_I
	gs0 := methodsDelta(ds, delta, i)
	return gs0.IsSupersetOf(gs)
}

func (i ITypeLit) Impls(ds []base.Decl, t base.Type) bool {
	u := asFGGType(t)
	return i.ImplsDelta(ds, make(Delta), u)
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
			s.Ok(ds, delta)
		case TNamed:
			k := s.String()
			if _, ok := seen_u[k]; ok {
				panic("Repeat embedding of type: " + k + "\n\t" + i.String())
			}
			seen_u[k] = s
			if !IsIfaceType(ds, s) { // CHECKME: allow embed type param?
				panic("Embedded type must be a named interface, not: " + k + "\n\t" + i.String())
			}
			s.Ok(ds, delta)
		default:
			panic("Unknown Spec kind: " + reflect.TypeOf(v).String() + "\n\t" +
				i.String())
		}
	}
	if i.HasTList() {
		i.tlist.Ok(ds, delta)
	}
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
	// todo compare type lists?
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
			return sigAlphaEquals(s1, g2)
		}
	}
	return false
}

func (i ITypeLit) String() string {
	var b strings.Builder
	b.WriteString(" interface { ")
	if i.HasTList() {
		b.WriteString(i.tlist.String())
	}
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

/******************************************************************************/
/* Type lists (cf. ITypeLit) -- not a type itself, just a helper */

type TypeList []Type

func (tlist0 TypeList) intersect(tlist TypeList) TypeList {
	inter := TypeList{}
	for _, t := range tlist0 {
		if tlist.Contains(t) {
			inter = append(inter, t)
		}
	}
	return inter
}

func (tlist0 TypeList) Contains(t Type) bool {
	for _, t2 := range tlist0 {
		if t2.Equals(t) {
			return true
		}
	}
	return false
}

func (tlist0 TypeList) Ok(ds []Decl, delta Delta) {
	seen_tl := make(map[string]Type) // key is u.String()
	for _, u := range tlist0 {
		k := u.String()
		if _, ok := seen_tl[k]; ok {
			panic("Duplicate type: " + k + " in type list\n\t" + tlist0.String())
		}
		seen_tl[k] = u

		if isIfaceType(ds, u) {
			under := u.Underlying(ds).(ITypeLit)
			if under.HasTList() {
				panic("") // todo   "interface Contains type constraints", according to go2goplay
			}
		}
		u.Ok(ds, delta)
	}
}

func (tlist0 TypeList) String() string {
	var b strings.Builder
	b.WriteString("type ")
	writeTypes(&b, tlist0)
	return b.String()
}

/******************************************************************************/
/* Aux */

// Cast to Type as defined in fgg.go.
// Panics if cast fails.
func asFGGType(t base.Type) Type {
	u, ok := t.(Type)
	if !ok {
		panic("Expected FGG type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	return u
}
