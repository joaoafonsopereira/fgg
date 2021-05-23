package fgg

import (
	"fmt"
	"reflect"
)

var _ = fmt.Errorf

/* GroundType, GroundEnv */

type GroundType interface { // TODO move??
	Type
	isGround() bool
}

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
	collectExpr(ds, make(GroundGamma), e_main, omega)
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
	u_recv GroundType // Pre: isGround
	meth   Name
	psi    SmallPsi // Pre: all isGround
}

// Pre: isGround(u_ground)
func toKey_Wt(u_ground GroundType) string {
	return u_ground.String()
}

// Pre: isGround(x.u_ground)
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
func collectExpr(ds []Decl, gamma GroundGamma, e FGGExpr, omega Omega) bool {
	res := false
	switch e1 := e.(type) {
	case Variable:
		return res
	case StructLit:
		for _, elem := range e1.elems {
			res = collectExpr(ds, gamma, elem, omega) || res
		}
		k := toKey_Wt(e1.u_S)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = e1.u_S
			res = true
		}
	case Select:
		return collectExpr(ds, gamma, e1.e_S, omega)
	case Call:
		res = collectExpr(ds, gamma, e1.e_recv, omega) || res
		for _, e_arg := range e1.args {
			res = collectExpr(ds, gamma, e_arg, omega) || res
		}
		gamma1 := make(Gamma)
		for k, v := range gamma {
			gamma1[k] = v
		}
		// TODO check _ assign
		tmp, _ := e1.e_recv.Typing(ds, make(Delta), gamma1, false)
		ground_recv := tmp.(GroundType)
		k_t := toKey_Wt(ground_recv)
		if _, ok := omega.us[k_t]; !ok {
			omega.us[k_t] = ground_recv
			res = true
		}
		m := MethInstan{ground_recv, e1.meth, e1.GetTArgs()} // N.B. type/method instans recorded separately
		k_m := toKey_Wm(m)
		if _, ok := omega.ms[k_m]; !ok {
			omega.ms[k_m] = m
			res = true
		}
	case Assert:
		res = collectExpr(ds, gamma, e1.e_I, omega) || res
		ground := e1.u_cast.(GroundType)
		k := toKey_Wt(ground)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = ground
			res = true
		}
	case Sprintf:
		for _, arg := range e1.args {
			res = collectExpr(ds, gamma, arg, omega) || res
		}

	case BinaryOperation: // TODO is it possible to factor out the repeated code?? <<<<<<<-----------
		res = collectExpr(ds, gamma, e1.left, omega) || res
		res = collectExpr(ds, gamma, e1.right, omega) || res
	case Comparison:
		res = collectExpr(ds, gamma, e1.left, omega) || res
		res = collectExpr(ds, gamma, e1.right, omega) || res

	case PrimitiveLiteral, BoolVal, Int32Val, Int64Val, Float32Val, Float64Val, StringVal: // TODO maybe create an interface to represent all primitives <<<<<<<-----------
		// Do nothing -- these nodes are leafs of the Ast, hence there is no
		// new type instantiations to be found underneath them.
		// Besides, there's no reason to collect primitive type 'instances', as
		// there is only 1 possible 'instance' and it has no methods.

	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e).String() + "\n\t" +
			e.String())
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
	//res = auxP(ds, omega) || res
	return res
}

func auxF(ds []Decl, omega Omega) bool {
	res := false
	tmp := make(map[string]GroundType)
	for _, u := range omega.us {
		if !isStructType(ds, u) { //|| u.Equals(STRING_TYPE) { // CHECKME
			continue
		}
		for _, u_f := range Fields(ds, u) {
			cast := u_f.u.(TNamed)
			tmp[toKey_Wt(cast)] = cast
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

func auxI(ds []Decl, omega Omega) bool {
	res := false
	tmp := make(map[string]MethInstan)
	for _, m := range omega.ms {
		if !IsNamedIfaceType(ds, m.u_recv) {
			continue
		}
		for _, m1 := range omega.ms {
			if !IsNamedIfaceType(ds, m1.u_recv) {
				continue
			}
			if m1.u_recv.Impls(ds, m.u_recv) {
				mm := MethInstan{m1.u_recv, m.meth, m.psi}
				tmp[toKey_Wm(mm)] = mm
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
}

func auxM(ds []Decl, omega Omega) bool {
	res := false
	tmp := make(map[string]GroundType)
	for _, m := range omega.ms {
		gs := methods(ds, m.u_recv)
		for _, g := range gs { // Should be only g s.t. g.meth == m.meth
			if g.meth != m.meth {
				continue
			}
			eta := MakeEta(g.Psi, m.psi)
			for _, pd := range g.pDecls {
				u_pd := pd.u.SubsEta(eta) // HERE: need receiver subs also? cf. map.fgg "type b Eq(b)" -- methods should be ok?
				tmp[toKey_Wt(u_pd)] = u_pd
			}
			u_ret := g.u_ret.SubsEta(eta)
			tmp[toKey_Wt(u_ret)] = u_ret
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

func auxS(ds []Decl, delta Delta, omega Omega) bool {
	res := false
	tmp := make(map[string]MethInstan)
	clone := omega.clone()
	for _, m := range clone.ms {
		for _, u := range clone.us {
			if !isStructType(ds, u) || !u.ImplsDelta(ds, delta, m.u_recv) {
				continue
			}
			//x0, xs, e := body(ds, u, m.meth, m.psi)
			x0, xs, e := body(ds, u.(TNamed), m.meth, m.psi) // TODO CHECK THIS CASTTTTTT <-------------------------------------------
			gamma := make(GroundGamma)
			gamma[x0.name] = x0.u.(GroundType)
			for _, pd := range xs {
				gamma[pd.name] = pd.u.(GroundType)
			}
			m1 := MethInstan{u, m.meth, m.psi}
			k := toKey_Wm(m1)
			//if _, ok := omega.ms[k]; !ok { // No: initial collectExpr already adds to omega.ms
			tmp[k] = m1
			res = collectExpr(ds, gamma, e, omega) || res
			//}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
}

// Add embedded types
func auxE1(ds []Decl, omega Omega) bool {
	res := false
	tmp := make(map[string]GroundType)
	for _, u := range omega.us {
		if !isNamedIfaceType(ds, u) {  // TODO pôr função que faz esta verificação a retornar logo u_I, td_I
			continue
		}
		u_I := u.Underlying(ds).(ITypeLit)
		td_I := getTDecl(ds, u.t_name)
		eta := MakeEta(td_I.Psi, u.u_args) //  TODO IS THIS SUBS NECESSARY AFTER Underlying() ?
		for _, s := range u_I.specs {
			if u_emb, ok := s.(TNamed); ok {
				u_sub := u_emb.SubsEta(eta)
				tmp[toKey_Wt(u_sub)] = u_sub
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

// Propagate method instances up to embedded supertypes
func auxE2(ds []Decl, omega Omega) bool {
	res := false
	tmp := make(map[string]MethInstan)
	for _, m := range omega.ms {
		if !isNamedIfaceType(ds, m.u_recv) {
			continue
		}
		u_I := m.u_recv.Underlying(ds).(ITypeLit) // TODO IS THIS SUBS NEEDED? OR IS IT ALREADY BEING APPLIED IN Underlying()?
		td_I := getTDecl(ds, m.u_recv.t_name)
		eta := MakeEta(td_I.Psi, m.u_recv.u_args)
		for _, s := range u_I.specs {
			if u_emb, ok := s.(TNamed); ok {
				u_sub := u_emb.SubsEta(eta)
				gs := methods(ds, u_sub)
				for _, g := range gs {
					if m.meth == g.meth {
						m_emb := MethInstan{u_sub, m.meth, m.psi}
						tmp[toKey_Wm(m_emb)] = m_emb
					}
				}
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
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
