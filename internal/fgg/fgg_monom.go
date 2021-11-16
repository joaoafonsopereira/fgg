package fgg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rhu1/fgg/internal/fg"
)

var _ = fmt.Errorf

/**
 * Monomorph
 */

/* Export */

func ToMonomId(u Type) fg.Type {
	//return toMonomId(u.(TNamed))
	return monomType(u, nil, nil, Omega{})
}

/* Mu -- method instance sets */

// All m (MethInstan.meth) belong to the same t (MethInstan.u_recv.t_name)
type Mu map[string]MethInstan // Cf. Omega, toKey_Wm

func makeMu(t Type, omega Omega) Mu {
	mu := make(Mu)
	for k, m := range omega.ms {
		if m.u_recv.Equals(t) {
			mu[k] = m
		}
	}
	return mu
}

/* Dummy signatures/methods */

var (
	Top         = fg.NewTNamed("Top")
	Top_Lit     = fg.NewStructLit(Top, []fg.FGExpr{})
	Empty_SType = fg.NewSTypeLit([]fg.FieldDecl{})
	Empty_Pds   = []fg.ParamDecl{}
)

/* Monomorph: FGGProgram -> FGProgram */

func Monomorph(p FGGProgram) fg.FGProgram {
	ds_fgg := p.GetDecls()
	omega := GetOmega(ds_fgg, p.GetMain().(FGGExpr))
	return ApplyOmega(p, omega)
}

func ApplyOmega(p FGGProgram, omega Omega) fg.FGProgram {
	var ds_monom []Decl
	for _, decl := range p.decls {
		switch d := decl.(type) {
		case TypeDecl:
			tds_monom := monomTDecl1(omega, d)
			for _, v := range tds_monom {
				ds_monom = append(ds_monom, v)
			}
		case MethDecl:
			mds_monom := monomMDecl1(omega, d)
			for _, v := range mds_monom {
				ds_monom = append(ds_monom, v)
			}
		default:
			panic("Unknown Decl kind: " + reflect.TypeOf(d).String() +
				"\n\t" + d.String())
		}
	}
	ds_monom = append(ds_monom, fg.NewTypeDecl(Top.GetName(), Empty_SType))
	e_monom := monomExpr1(p.e_main, make(EtaClosed), omega)
	return fg.NewFGProgram(ds_monom, e_monom, p.printf)
}

func monomTDecl1(omega Omega, td TypeDecl) []fg.TypeDecl {
	var res []fg.TypeDecl
	t := td.GetName()
	for _, u := range omega.us {
		u_N, ok := u.(TNamed)
		if ok && u_N.t_name == t {
			eta := MakeEtaClosed(td.GetBigPsi(), u_N.u_args)
			mu := makeMu(u_N, omega)

			t_monom := toMonomId(u_N)
			src_monom := monomType(td.GetSourceType(), eta, mu, omega)
			td_monom := fg.NewTypeDecl(t_monom.GetName(), src_monom)
			res = append(res, td_monom)
		}
	}
	return res
}

func monomType(fgg_type Type, eta EtaClosed, mu Mu, omega Omega) fg.Type {
	switch t := fgg_type.(type) {
	case TParam:
		subs := t.SubsEtaClosed(eta)
		return monomType(subs, eta, nil, omega)
	case TPrimitive:
		return fg.NewTPrimitive(fg.Tag(t.tag))
	case UndefTPrimitive:
		return fg.NewUndefTPrimitive(fg.Tag(t.tag))
	case TNamed:
		return monomTNamed(t, eta)
	case STypeLit:
		return monomSTypeLit1(t, eta, omega)
	case ITypeLit:
		// convention: when this case is reached with mu == nil, it means
		// that monomType was applied to an 'anonymous' interface.
		// -> need to find the method instance set for that interface,
		//    as it may specify generic methods.
		if mu == nil {
			mu = makeMu(t, omega)
		}
		return monomITypeLit1(t, eta, mu, omega)
	default:
		panic("Couldn't monomorphise the type " + t.String())
	}
}

func monomTNamed(u TNamed, eta EtaClosed) fg.TNamed {
	t_subs := u.SubsEtaClosed(eta).(TNamed)
	return toMonomId(t_subs)
}

func monomSTypeLit1(s STypeLit, eta EtaClosed, omega Omega) fg.STypeLit {
	fds := make([]fg.FieldDecl, len(s.fDecls))
	for i, fd := range s.fDecls {
		t_monom := monomType(fd.u, eta, nil, omega)
		fds[i] = fg.NewFieldDecl(fd.field, t_monom)
	}
	return fg.NewSTypeLit(fds)
}

func monomITypeLit1(c ITypeLit, eta EtaClosed, mu Mu, omega Omega) fg.ITypeLit {
	var ss []fg.Spec
	for _, spec := range c.specs {
		switch s := spec.(type) {
		case Sig: // !!! M contains Psi
			for _, m := range mu {
				if m.meth != s.meth {
					continue
				}
				theta := MakeEtaClosed(s.Psi, m.psi)
				for k, v := range eta {
					theta[k] = v
				}
				g_monom := monomSig1(s, m, theta, omega) // !!! small psi
				ss = append(ss, g_monom)
			}
			// dummy
			dummyName := toHashSig(s.SubsEtaOpen(eta.ToEtaOpen())) // eta is open in this context - doesn't include mappings for vars in s.Psi (contrasting with theta above)
			hash := fg.NewSig(dummyName, Empty_Pds, Top)
			ss = append(ss, hash)
		case TNamed: // Embedded
			ss = append(ss, monomTNamed(s, eta))
		default:
			panic("Unknown Spec kind: " + reflect.TypeOf(spec).String() +
				"\n\t" + spec.String())
		}
	}
	return fg.NewITypeLit(ss)
}

// Encapsulates 2 rules:
//  - M-MFormal (generate "monomorphised" method name)
//  - M-Sig  (monomorphise pDecls' & return's types)
func monomSig1(g Sig, m MethInstan, eta EtaClosed, omega Omega) fg.Sig {
	//getMonomMethName(omega Omega, m Name, targs []Type) Name {
	m_monom := toMonomMethName1(m.meth, m.psi, eta, omega) // !!! small psi
	pds_monom := make([]fg.ParamDecl, len(g.pDecls))
	for i, pd := range g.pDecls {
		t_monom := monomType(pd.u, eta, nil, omega)
		pds_monom[i] = fg.NewParamDecl(pd.name, t_monom)
	}
	ret_monom := monomType(g.u_ret, eta, nil, omega)
	return fg.NewSig(m_monom, pds_monom, ret_monom)
}

func monomMDecl1(omega Omega, md MethDecl) []fg.MethDecl {
	var res []fg.MethDecl
	// D
	for _, m := range omega.ms {
		u_recv, isTNamed := m.u_recv.(TNamed)
		if !(isTNamed && u_recv.t_name == md.t_recv && m.meth == md.name) {
			continue
		}
		theta := MakeEtaClosed(md.Psi_recv, u_recv.u_args)
		for i := 0; i < len(md.Psi_meth.tFormals); i++ {
			theta[md.Psi_meth.tFormals[i].name] = m.psi[i].(GroundType)
		}
		recv_monom := fg.NewParamDecl(md.x_recv, toMonomId(u_recv))                           // !!! t_S(phi) already ground receiver
		g_monom := monomSig1(Sig{md.name, md.Psi_meth, md.pDecls, md.u_ret}, m, theta, omega) // !!! small psi
		e_monom := monomExpr1(md.e_body, theta, omega)
		md_monom := fg.NewMDecl(recv_monom,
			g_monom.GetMethod(), g_monom.GetParamDecls(), g_monom.GetReturn(),
			e_monom)
		res = append(res, md_monom)
	}
	// D'
	for _, u := range omega.us {
		u_recv, isTNamed := u.(TNamed)
		if !isTNamed || u_recv.t_name != md.t_recv {
			continue
		}
		recv_monom := fg.NewParamDecl(md.x_recv, toMonomId(u_recv)) // !!! t_S(phi) already ground receiver
		// has to be open: doesn't include mappings for md.Psi
		// (contrasting with e.g. theta, above)
		eta := MakeEtaOpen(md.Psi_recv, u_recv.u_args)
		g := md.ToSig().SubsEtaOpen(eta)
		dummy := fg.NewMDecl(recv_monom, toHashSig(g), Empty_Pds, Top, Top_Lit)
		res = append(res, dummy)
	}
	return res
}

func monomExpr1(e1 FGGExpr, eta EtaClosed, omega Omega) fg.FGExpr {
	switch e := e1.(type) {
	case Variable:
		return fg.NewVariable(e.name)
	case StructLit:
		es_monom := make([]fg.FGExpr, len(e.elems))
		for i := 0; i < len(e.elems); i++ {
			es_monom[i] = monomExpr1(e.elems[i], eta, omega)
		}
		t_monom := monomType(e.u_S, eta, nil, omega)
		return fg.NewStructLit(t_monom, es_monom)
	case Select:
		return fg.NewSelect(monomExpr1(e.e_S, eta, omega), e.field)
	case Call:
		e_monom := monomExpr1(e.e_recv, eta, omega)
		m_monom := toMonomMethName1(e.meth, e.t_args, eta, omega)
		es_monom := make([]fg.FGExpr, len(e.args))
		for i := 0; i < len(e.args); i++ {
			es_monom[i] = monomExpr1(e.args[i], eta, omega)
		}
		return fg.NewCall(e_monom, m_monom, es_monom)
	case Assert:
		e_monom := monomExpr1(e.e_I, eta, omega)
		t_monom := monomType(e.u_cast, eta, nil, omega)
		return fg.NewAssert(e_monom, t_monom)

	case Convert:
		t_monom := monomType(e.typ, eta, nil, omega)
		e_monom := monomExpr1(e.expr, eta, omega)
		return fg.NewConvert(t_monom, e_monom)

	case Sprintf:
		args := make([]fg.FGExpr, len(e.args))
		for i := 0; i < len(e.args); i++ {
			args[i] = monomExpr1(e.args[i], eta, omega)
		}
		return fg.NewSprintf(e.format, args)

	// todo factor this exactly-the-same code
	case BinaryOperation:
		left_monom := monomExpr1(e.left, eta, omega)
		right_monom := monomExpr1(e.right, eta, omega)
		return fg.NewBinaryOp(left_monom, right_monom, fg.Operator(e.op))
	case Comparison:
		left_monom := monomExpr1(e.left, eta, omega)
		right_monom := monomExpr1(e.right, eta, omega)
		return fg.NewBinaryOp(left_monom, right_monom, fg.Operator(e.op))

	case PrimitiveLiteral:
		return fg.NewPrimitiveLiteral(e.payload, fg.Tag(e.tag))

	case TypedPrimitiveValue:
		lit_monom := fg.NewPrimitiveLiteral(e.lit.payload, fg.Tag(e.lit.tag))
		t_monom := monomType(e.typ, eta, nil, omega)
		return fg.NewTypedPrimitiveValue(lit_monom, t_monom)

	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e1).String() + "\n\t" +
			e1.String())
	}
}

/* Helpers */

/*// Pre: len(targs) > 0
func getMonomMethName(omega Omega, m Name, targs []Type) Name {
	first := toMonomId(omega[toWKey(targs[0].(TNamed))].u_ground)
	res := m + "<" + first.String()
	for _, v := range targs[1:] {
		next := toMonomId(omega[toWKey(v.(TNamed))].u_ground)
		res = res + "," + next.String()
	}
	res = res + ">"
	return Name(res)
}*/

func toMonomId(u TNamed) fg.TNamed {
	res := u.String()
	res = strings.Replace(res, ",", ",,", -1) // TODO: refactor, cf. Frontend.monomOutputHack
	res = strings.Replace(res, "(", "<", -1)
	res = strings.Replace(res, ")", ">", -1)
	res = strings.Replace(res, " ", "", -1)
	return fg.NewTNamed(res)
}

// !!! CHECKME: psi should already be grounded, eta unnecessary?
// N.B. this method is kind of overloaded,
// as it represents both M-METHOD & M-MFORMAL (fig. 21 of paper)
func toMonomMethName1(m Name, psi SmallPsi, eta EtaClosed, omega Omega) Name {
	if len(psi) == 0 {
		return m + "<>"
	}
	//first := toMonomId(psi[0].SubsEta(eta))
	first := monomType(psi[0], eta, nil, omega)
	res := m + "<" + first.String()
	for _, v := range psi[1:] {
		next := monomType(v, eta, nil, omega)
		res = res + ",," + next.String() // Cf. Frontend.monomOutputHack -- TODO: factor out
	}
	res = res + ">"
	return Name(res)
}

/* Works because duck typing uses nominal method sets, cf.
type Any1 interface {};
type Any2 interface {};
type A struct {};
func (x0 A) foo() Any1 { return x0 };
type IB interface { foo() Any2 };
type toAny1 struct { any Any1 };
func main() { _ = toAny1{A{}}.any.(IB) } // assertion failure */
func toHashSig(g Sig) string {
	/*subs := make(Delta)
	for i := 0; i < len(g.Psi.tFormals); i++ {
		subs[g.Psi.tFormals[i].name] = TParam("α" + strconv.Itoa(i+1))
	}
	g1 := g.TSubs(subs)*/
	g1 := g
	var b strings.Builder
	b.WriteString(g.meth)
	b.WriteString("_")
	for _, v := range g1.Psi.tFormals {
		b.WriteString("_")
		b.WriteString(v.name.String())
		b.WriteString("_")
		b.WriteString(v.u_I.String())
	}
	b.WriteString("_")
	for _, v := range g1.pDecls { // Formal param names discarded
		b.WriteString("_")
		b.WriteString(v.u.String())
	}
	b.WriteString("_")
	b.WriteString(g1.u_ret.String()) // TODO upon generating dummy sigs this shouldn't be done, as the ret type will be Top
	res := b.String()
	res = strings.Replace(res, "(", "_", -1) // TODO
	res = strings.Replace(res, ")", "_", -1)
	res = strings.Replace(res, ",", "_", -1)
	res = strings.Replace(res, " ", "", -1)
	return res
}









































/* Deprecated -- Simplistic isMonom check:
   no typeparam nested in a named type in typeargs of StructLit/Call exprs */

/*
func IsMonomable(p FGGProgram) (FGGExpr, bool) {
	for _, v := range p.decls {
		switch d := v.(type) {
		case STypeLit:
		case ITypeLit:
		case MethDecl:
			if e, ok := isMonomableMDecl(d); !ok {
				return e, false
			}
		default:
			panic("Unknown Decl kind: " + reflect.TypeOf(v).String() + "\n\t" +
				v.String())
		}
	}
	return isMonomableExpr(p.e_main)
}

func isMonomableMDecl(d MethDecl) (FGGExpr, bool) {
	return isMonomableExpr(d.e_body)
}

// Post: if bool is true, Expr is the offender; o/w disregard Expr
func isMonomableExpr(e FGGExpr) (FGGExpr, bool) {
	switch e1 := e.(type) {
	case Variable:
		return e1, true
	case StructLit:
		for _, v := range e1.u_S.u_args {
			if u1, ok := v.(TNamed); ok {
				if isOrContainsTParam(u1) {
					return e1, false
				}
			}
		}
		for _, v := range e1.elems {
			if e2, ok := isMonomableExpr(v); !ok {
				return e2, false
			}
		}
		return e1, true
	case Select:
		return isMonomableExpr(e1.e_S)
	case Call:
		for _, v := range e1.t_args {
			if u1, ok := v.(TNamed); ok {
				if isOrContainsTParam(u1) {
					return e1, false
				}
			}
		}
		if e2, ok := isMonomableExpr(e1.e_recv); !ok {
			return e2, false
		}
		for _, v := range e1.args {
			if e2, ok := isMonomableExpr(v); !ok {
				return e2, false
			}
		}
		return e1, true
	case Assert:
		if u1, ok := e1.u_cast.(TNamed); ok {
			if isOrContainsTParam(u1) {
				return e1, false
			}
		}
		return isMonomableExpr(e1.e_I)
	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e).String() + "\n\t" +
			e.String())
	}
}
*/

// returns true iff u is a TParam or contains a TParam
func isOrContainsTParam(u Type) bool {
	if _, ok := u.(TParam); ok {
		return true
	}
	u1 := u.(TNamed)
	for _, v := range u1.u_args {
		if isOrContainsTParam(v) {
			return true
		}
	}
	return false
}
