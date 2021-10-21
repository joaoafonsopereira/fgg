package fgg

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rhu1/fgg/internal/base"
)

var _ = fmt.Errorf
var _ = reflect.Append
var _ = strconv.AppendBool

/* Export */

func IsStructType(ds []Decl, u Type) bool      { return isStructType(ds, u) }
func IsIfaceType(ds []Decl, u Type) bool       { return isIfaceType(ds, u) }
func IsIfaceLikeType(ds []Decl, u Type) bool   { return isIfaceLikeType(ds, u) }
func NewTFormal(name TParam, u_I Type) TFormal { return TFormal{name, u_I} }
func NewBigPsi(tFormals []TFormal) BigPsi      { return BigPsi{tFormals} }

/* Aliases from base */

type Name = base.Name
type FGGNode = base.AstNode
type Decl = base.Decl

/* Name, Type, Type param, Type name -- !!! submission version, "Type name" overloaded */

// Name: see Aliases (at top)

type Type interface {
	base.Type
	ImplsDelta(ds []Decl, delta Delta, u Type) bool
	SubsEtaOpen(eta EtaOpen) Type
	SubsEtaClosed(eta EtaClosed) GroundType
	Ok(ds []Decl, delta Delta)
	ToGoString(ds []Decl) string
	Underlying(ds []Decl) Type
}

/* Type formals and actuals */

// Pre: len(as) == len(us)
// Wrapper for []TFormal (cf. e.g., FieldDecl), only because of "(type ...)" syntax
// Also ranged over by big phi
type BigPsi struct {
	tFormals []TFormal
}

func (Psi BigPsi) GetTFormals() []TFormal { return Psi.tFormals }

func (Psi BigPsi) Ok(ds []Decl, env Delta) {
	extendedEnv := env.Clone()
	for _, v := range Psi.tFormals {
		if _, ok := extendedEnv[v.name]; ok {
			panic("Duplicate param name " + string(v.name) + " under context: " +
				env.String() + "\n\t" + Psi.String())
		}
		extendedEnv[v.name] = v.u_I
	} // Delta built
	for _, v := range Psi.tFormals {
		if !isIfaceType(ds, v.u_I) {
			panic("Upper bound must be an interface type: not " + v.u_I.String() +
				"\n\t" + Psi.String())
		}
		v.u_I.Ok(ds, extendedEnv) // Checks params bound under env -- N.B. can forward ref (not restricted left-to-right)
	}
}

func (Psi BigPsi) ToDelta() Delta {
	delta := make(map[TParam]Type)
	for _, v := range Psi.tFormals {
		delta[v.name] = v.u_I
	}
	return delta
}

// The ordered value set of ToDelta
func (Psi BigPsi) Hat() SmallPsi {
	res := make(SmallPsi, len(Psi.tFormals))
	for i, v := range Psi.tFormals {
		res[i] = v.name
	}
	return res
}

func (Psi BigPsi) String() string {
	var b strings.Builder
	b.WriteString("(type ") // Includes "(...)" -- cf. e.g., writeFieldDecls
	if len(Psi.tFormals) > 0 {
		b.WriteString(Psi.tFormals[0].String())
		for _, v := range Psi.tFormals[1:] {
			b.WriteString(", ")
			b.WriteString(v.String())
		}
	}
	b.WriteString(")")
	return b.String()
}

type TFormal struct {
	name TParam
	u_I  Type
	// CHECKME: submission version, upper bound \tau_I is only "of the form t_I(~\tau)"? -- i.e., not \alpha?
	// ^If so, then can refine to TNamed
	//  -> It can also be an interface literal / an anonymous interface, hence
	//     it's not possible
}

func (tf TFormal) GetTParam() TParam   { return tf.name }
func (tf TFormal) GetUpperBound() Type { return tf.u_I }

func (tf TFormal) String() string {
	return string(tf.name) + " " + tf.u_I.String()
}

func (tf TFormal) SubsEtaOpen(eta EtaOpen) TFormal {
	return TFormal{tf.name.SubsEtaOpen(eta).(TParam), tf.u_I.SubsEtaOpen(eta)}
}

// Type actuals
// Also ranged over by small phi
type SmallPsi []Type // CHECKME: Currently only used in omega/monom, maybe deprecate?

func (x0 SmallPsi) SubsEtaOpen(eta EtaOpen) SmallPsi {
	res := make(SmallPsi, len(x0))
	for i, v := range x0 {
		res[i] = v.SubsEtaOpen(eta)
	}
	return res
}

func (x0 SmallPsi) String() string {
	var b strings.Builder
	for _, v := range x0 {
		b.WriteString(v.String())
	}
	return b.String()
}

func (x0 SmallPsi) Equals(x SmallPsi) bool {
	if len(x0) != len(x) {
		return false
	}
	for i := 0; i < len(x0); i++ {
		if !x0[i].Equals(x[i]) {
			return false
		}
	}
	return true
}

/* Context, Type context, Substitutions */

//type Gamma map[Variable]Type
type Gamma map[Name]Type
type Delta map[TParam]Type // Type intended to be an upper bound

type EtaOpen map[TParam]Type         // cf. Delta todo would it make more sense to be just Eta? contrasting with EtaClosed
type EtaClosed map[TParam]GroundType // TNamed intended to be a ground

func (delta Delta) String() string {
	res := "["
	first := true
	for k, v := range delta {
		if first {
			first = false
		} else {
			res = res + ", "
		}
		res = k.String() + ":" + v.String()
	}
	return res + "]"
}

// Deltas should be treated immutably, hence one must clone the current
// environment (Delta) in order to add new mappings to it.
func (delta Delta) Clone() Delta {
	clone := make(Delta)
	for param, bound := range delta {
		clone[param] = bound
	}
	return clone
}

// Pre: len(psi) == len(Psi.GetTFormals()); psi all ground
func MakeEtaClosed(Psi BigPsi, psi SmallPsi) EtaClosed {
	eta := make(EtaClosed)
	tfs := Psi.tFormals
	for i := 0; i < len(tfs); i++ {
		eta[tfs[i].name] = psi[i].(GroundType) // TODO maybe SmallPsi could be []GroundType ?
	}
	return eta
}

// Although the result only contains mappings to ground types,
// this conversion is necessary in the cases where there is no mapping
// for a given type variable. Cf. Sig.SubsEtaOpen
func (eta EtaClosed) ToEtaOpen() EtaOpen {
	res := make(EtaOpen)
	for k, v := range eta {
		res[k] = v
	}
	return res
}

func MakeEtaDelta(ds []Decl, delta Delta, Psi BigPsi, psi SmallPsi) (bool, EtaOpen) {
	eta := MakeEtaOpen(Psi, psi)
	for _, v := range Psi.tFormals {
		a := v.name.SubsEtaOpen(eta)
		u_I := v.u_I.SubsEtaOpen(eta)
		if !a.ImplsDelta(ds, delta, u_I) {
			return false, eta
		}
	}
	return true, eta
}

func MakeEtaOpen(Psi BigPsi, psi SmallPsi) EtaOpen {
	eta := make(EtaOpen)
	tfs := Psi.tFormals
	for i := 0; i < len(tfs); i++ {
		eta[tfs[i].name] = psi[i]
	}
	return eta
}

/* AST base intefaces: FGGNode, Decl, Spec, Expr */

// FGGNode, Name: see Aliases (at top)

type Spec interface {
	FGGNode
	GetSigs(ds []Decl) []Sig
}

type FGGExpr interface {
	base.Expr
	Subs(subs map[Variable]FGGExpr) FGGExpr
	TSubs(subs EtaOpen) FGGExpr // TODO maybe rename to SubsEtaOpen (future SubsEta), for consistency?
	// gamma and delta should be treated immutably
	Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr)
	Eval(ds []Decl) (FGGExpr, string)

	Infer(ds []Decl, delta Delta, gamma Gamma) Type // todo retornar um Delta que representa bounds de TVar ainda nao unificada

}

/* Helpers */

// Check if u is a \tau_S
func isStructType(ds []Decl, u Type) bool {
	_, ok := u.Underlying(ds).(STypeLit)
	return ok
}

func isIfaceType(ds []Decl, u Type) bool {
	_, ok := u.Underlying(ds).(ITypeLit)
	return ok
}

// checks if u is a u_J
func isIfaceLikeType(ds []Decl, u Type) bool {
	isIface := isIfaceType(ds, u)
	_, isTParam := u.(TParam)
	return isIface || isTParam
}

func writeTypes(b *strings.Builder, us []Type) {
	if len(us) > 0 {
		b.WriteString(us[0].String())
		for _, v := range us[1:] {
			b.WriteString(", " + v.String())
		}
	}
}

func writeToGoTypes(ds []Decl, b *strings.Builder, us []Type) {
	if len(us) > 0 {
		b.WriteString(us[0].ToGoString(ds))
		for _, v := range us[1:] {
			b.WriteString(", " + v.ToGoString(ds))
		}
	}
}
