package fgg

import (
	"fmt"
	"reflect"
)

var _ = fmt.Errorf

/* GroundType, GroundEnv */

type GroundType interface { // TODO move??
	Type
	Ground()
}

// kind of a sum type
func (u TNamed) Ground()           {}
func (u STypeLit) Ground()         {}
func (u ITypeLit) Ground()         {}
func (t0 TPrimitive) Ground()      {}
func (t0 UndefTPrimitive) Ground() {}

// Basically a Gamma for only ground types
type GroundGamma map[Name]GroundType

/**
 * Build Omega -- (morally) a map from ground FGG types to Sigs of (potential)
 * calls on that receiver.  N.B., calls are recorded only as seen for each
 * specific receiver type -- i.e., omega does not attempt to "respect"
 * subtyping (cf. "zigzagging" in fgg_monom).
 */

// Pre: IsMonomOK
func GetOmega(ds []Decl, e_main FGGExpr) Omega {
	omega := Omega{make(map[string]GroundType), make(map[string]MethInstan)}
	collectExpr(ds, make(GroundGamma), omega, e_main)
	fixomega(ds, omega)
	//omega.Println()
	return omega
}

/* Omega, MethInstan */

type Omega struct {
	// Keys given by toKey_Wt, toKey_Wm
	us map[string]GroundType
	ms map[string]MethInstan
}

func (w Omega) addTInst(u GroundType) bool {
	key := toKey_Wt(u)
	if _, ok := w.us[key]; !ok {
		w.us[key] = u
		return true
	}
	return false
}

func (w Omega) addTInsts(insts map[string]GroundType) bool {
	res := false
	for k, u := range insts {
		if _, ok := w.us[k]; !ok {
			w.us[k] = u
			res = true
		}
	}
	return res
}

func (w Omega) addMInst(m MethInstan) bool {
	key := toKey_Wm(m)
	if _, ok := w.ms[key]; !ok {
		w.ms[key] = m
		return true
	}
	return false
}

func (w Omega) addMInsts(ms map[string]MethInstan) bool {
	res := false
	for k, m := range ms {
		if _, ok := w.ms[k]; !ok {
			w.ms[k] = m
			res = true
		}
	}
	return res
}

func (w Omega) clone() Omega {
	us := make(map[string]GroundType)
	ms := make(map[string]MethInstan)
	for k, v := range w.us {
		us[k] = v
	}
	for k, v := range w.ms {
		ms[k] = v
	}
	return Omega{us, ms}
}

func (w Omega) Println() {
	fmt.Println("=== Type instances:")
	for _, v := range w.us {
		fmt.Println(v)
	}
	fmt.Println("--- Method instances:")
	for _, v := range w.ms {
		fmt.Println(v.u_recv, v.meth, v.psi)
	}
	fmt.Println("===")
}

type MethInstan struct {
	// u_recv can't be refined into a TNamed: have to account for the case
	// where a variable (e.g. a struct field) has an interface literal as its
	// declared type. This 'anonymous' interface may itself have generic methods
	// for which we also have to collect the method instances -- cf. Test111
	// in fgg_prims_test.go.
	// N.B the anonymous interface itself is a ground type.
	u_recv GroundType
	//u_recv TNamed // Pre: isGround
	meth Name
	psi  SmallPsi // Pre: all isGround
}

func toKey_Wt(u_ground GroundType) string {
	return u_ground.String()
}

func toKey_Wm(x MethInstan) string {
	return x.u_recv.String() + "_" + x.meth + "_" + x.psi.String()
}

/* fixOmega */

func fixomega(ds []Decl, omega Omega) {
	/*fmt.Println("......initial.........", len(omega.us), len(omega.ms))
	omega.Println()
	fmt.Println(".............", len(omega.us), len(omega.ms))*/
	for auxG(ds, omega) {
		//omega.Println()
		//fmt.Println(".............", len(omega.us), len(omega.ms))
	}
}

/* Expressions */

// gamma used to type Call receiver
func collectExpr(ds []Decl, gamma GroundGamma, omega Omega, e FGGExpr) bool {
	res := false
	switch e1 := e.(type) {
	case Variable:
		return res
	case StructLit:
		res = collectExprs(ds, gamma, omega, e1.elems...)
		u_S := e1.u_S.(GroundType)
		res = omega.addTInst(u_S) || res
	case Select:
		return collectExpr(ds, gamma, omega, e1.e_S)
	case Call:
		res = collectExpr(ds, gamma, omega, e1.e_recv)
		res = collectExprs(ds, gamma, omega, e1.args...) || res
		gamma1 := make(Gamma)
		for k, v := range gamma {
			gamma1[k] = v
		}

		tmp, _ := e1.e_recv.Typing(ds, make(Delta), gamma1, false)
		ground_recv := tmp.(GroundType)
		res = omega.addTInst(ground_recv) || res

		m := MethInstan{ground_recv, e1.meth, e1.GetTArgs()} // N.B. type/method instans recorded separately
		res = omega.addMInst(m) || res

	case Assert:
		res = collectExpr(ds, gamma, omega, e1.e_I)
		ground := e1.u_cast.(GroundType)
		res = omega.addTInst(ground) || res
	case Convert:
		res = collectExpr(ds, gamma, omega, e1.expr)
		ground := e1.typ.(GroundType)
		res = omega.addTInst(ground) || res
	case Sprintf:
		res = collectExprs(ds, gamma, omega, e1.args...)

	case BinaryOperation:
		res = collectExprs(ds, gamma, omega, e1.left, e1.right)
	case Comparison:
		res = collectExprs(ds, gamma, omega, e1.left, e1.right)

	// This may be needed to handle instances of types declared like:
	// type MyInt[T any] int32
	case TypedPrimitiveValue:
		if u_N, ok := e1.typ.(TNamed); ok {
			res = omega.addTInst(u_N)
		}
	case PrimitiveLiteral:
		return res

	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e).String() + "\n\t" +
			e.String())
	}
	return res
}
func collectExprs(ds []Decl, gamma GroundGamma, omega Omega, es ...FGGExpr) bool {
	res := false
	for _, arg := range es {
		res = collectExpr(ds, gamma, omega, arg) || res
	}
	return res
}

/* Aux */

// Return true if omega has changed
// N.B. mutating omega in each sub-step -- can produce many levels of nesting within one G step
// ^also non-deterministic progress, because mutating maps while ranging; also side-effect results may depend on iteration order over maps
// N.B. no closure over types occurring in bounds, or *interface decl* method sigs
func auxG(ds []Decl, omega Omega) bool {
	res := false
	res = auxF(ds, omega) || res
	res = auxI(ds, omega) || res
	res = auxM(ds, omega) || res
	res = auxS(ds, make(Delta), omega) || res
	// I/face embeddings
	res = auxE1(ds, omega) || res
	res = auxE2(ds, omega) || res
	// type declarations
	res = auxT(ds, omega) || res
	return res
}

//Pair := "type Pair(type X Any(), Y Any()) struct { x X; y Y}"
//PairInt := "type PairInt(type ) Pair(int32, int32)"
//
//Pair := "type Pair(type X Any(), Y Any()) struct { x X; y Y}"
//PairEq := "type PairEq(type T Any()) Pair(T, T)"
//PairInt := "type PairInt(type ) PairEq(int32)"
func auxT(ds []Decl, omega Omega) bool {
	tmp := make(map[string]GroundType)
	for _, u := range omega.us {
		u_N, isTNamed := u.(TNamed)
		if !isTNamed {
			continue
		}
		td := getTDecl(ds, u_N.t_name)
		if src, ok := td.GetSourceType().(TNamed); ok {
			eta := MakeEtaClosed(td.GetBigPsi(), u_N.GetTArgs())
			ground_src := src.SubsEtaClosed(eta)
			tmp[toKey_Wt(ground_src)] = ground_src
		}
	}
	return omega.addTInsts(tmp)
}

func auxF(ds []Decl, omega Omega) bool {
	tmp := make(map[string]GroundType)
	for _, u := range omega.us {
		// underlying of a ground type is itself ground
		if u_S, ok := u.Underlying(ds).(STypeLit); ok {
			for _, fd := range u_S.GetFieldDecls() {
				// in a ground struct, every field has ground type
				ground := fd.u.(GroundType)
				tmp[toKey_Wt(ground)] = ground
			}
		}
	}
	return omega.addTInsts(tmp)
}

func auxI(ds []Decl, omega Omega) bool {
	tmp := make(map[string]MethInstan)
	for _, m := range omega.ms {
		if !isIfaceType(ds, m.u_recv) {
			continue
		}
		// more precise to iterate us: might have iface types that were only
		// added as type instances, without "mapping" to a method inst
		// (e.g. only appearing in a struct field)
		for _, u := range omega.us {
			if !isIfaceType(ds, u) || u.Equals(m.u_recv) {
				continue
			}
			if ImplsDelta(ds, Delta{}, u, getInterface(ds, m.u_recv)) {
				mm := MethInstan{u, m.meth, m.psi}
				tmp[toKey_Wm(mm)] = mm
			}
		}
	}
	return omega.addMInsts(tmp)
}

func auxM(ds []Decl, omega Omega) bool {
	tmp := make(map[string]GroundType)
	for _, m := range omega.ms {
		sig := methods(ds, m.u_recv)[m.meth]
		eta := MakeEtaClosed(sig.Psi, m.psi)
		for _, pd := range sig.pDecls {
			u_pd := pd.u.SubsEtaClosed(eta) // HERE: need receiver subs also? cf. map.fgg "type b Eq(b)" -- methods should be ok? -> no: subs performed in methods(m.u_recv)
			tmp[toKey_Wt(u_pd)] = u_pd
		}
		u_ret := sig.u_ret.SubsEtaClosed(eta)
		tmp[toKey_Wt(u_ret)] = u_ret
	}
	return omega.addTInsts(tmp)
}

func auxS(ds []Decl, delta Delta, omega Omega) bool {
	res := false
	tmp := make(map[string]MethInstan)
	clone := omega.clone()
	for _, m := range clone.ms {
		for _, u := range clone.us {
			u_N, ok := u.(TNamed)
			if !ok || isIfaceType(ds, u_N) {
				continue
			}
			assignable, _ := u_N.AssignableToDelta(ds, delta, m.u_recv)
			if !assignable {
				continue
			}

			x0, xs, e := body(ds, u_N, m.meth, m.psi)
			gamma := make(GroundGamma)
			gamma[x0.name] = x0.u.(GroundType)
			for _, pd := range xs {
				gamma[pd.name] = pd.u.(GroundType)
			}
			m1 := MethInstan{u, m.meth, m.psi}
			k := toKey_Wm(m1)
			//if _, ok := omega.ms[k]; !ok { // No: initial collectExpr already adds to omega.ms
			tmp[k] = m1
			res = collectExpr(ds, gamma, omega, e) || res
			//}
		}
	}
	return omega.addMInsts(tmp) || res
}

// Add embedded types
func auxE1(ds []Decl, omega Omega) bool {
	tmp := make(map[string]GroundType)
	for _, u := range omega.us {
		if u_I, ok := u.Underlying(ds).(ITypeLit); ok {
			for _, s := range u_I.specs {
				if u_emb, ok := s.(TNamed); ok {
					// omega.us contains only ground types -> their underlying
					// types are also ground types, thus no need for
					// explicit substitutions over u_emb's
					tmp[toKey_Wt(u_emb)] = u_emb
				}
			}
		}
	}
	return omega.addTInsts(tmp)
}

// Propagate method instances up to embedded supertypes
func auxE2(ds []Decl, omega Omega) bool {
	tmp := make(map[string]MethInstan)
	for _, m := range omega.ms {
		if !isIfaceType(ds, m.u_recv) {
			continue
		}
		u_I := getInterface(ds, m.u_recv)
		for _, s := range u_I.GetSpecs() {
			if u_emb, ok := s.(TNamed); ok {
				if _, hasMeth := methods(ds, u_emb)[m.meth]; hasMeth {
					m_emb := MethInstan{u_emb, m.meth, m.psi}
					tmp[toKey_Wm(m_emb)] = m_emb
				}
			}
		}
	}
	return omega.addMInsts(tmp)
}

/* Helpers */

func isGround(u TNamed) bool {
	for _, v := range u.u_args {
		if u1, ok := v.(TNamed); !ok {
			return false
		} else if !isGround(u1) {
			return false
		}
	}
	return true
}
