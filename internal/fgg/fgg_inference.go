package fgg

import (
	"strconv"
	"strings"
)

/* Subtype and Equality constraints - definition and basic methods */

// Constraint of the form u1 <: u2
type SubtypeConstr struct{
	u1, u2 Type
}

func NewSubtypeConstr(s, t Type) SubtypeConstr {
	return SubtypeConstr{s, t}
}
func (c SubtypeConstr) SubsEtaOpen(eta EtaOpen) SubtypeConstr {
	return SubtypeConstr{c.u1.SubsEtaOpen(eta), c.u2.SubsEtaOpen(eta)}
}

// Constraint of the form u1 == u2
type EqualityConstr struct{
	u1, u2 Type
}

func NewEqualityConstr(t1, t2 Type) EqualityConstr {
	return EqualityConstr{t1, t2}
}

func (c EqualityConstr) SubsEtaOpen(eta EtaOpen) EqualityConstr {
	return EqualityConstr{c.u1.SubsEtaOpen(eta), c.u2.SubsEtaOpen(eta)}
}

/* Sets of constraints - duplicated functionalities for lack of a better solution (e.g. generics) */

// Set of equality constraints
// note: assumes that the constraints are unifiable "1 by 1" -- C.f. UnifyAll and compose
type EqConstraintSet []EqualityConstr

func NewEqConstraintSet() EqConstraintSet {
	return make(EqConstraintSet, 0, 10)
}

func (cs EqConstraintSet) SubsEtaOpen(eta EtaOpen) EqConstraintSet {
	for i, c := range cs {
		cs[i] = c.SubsEtaOpen(eta)
	}
	return cs
}

func (cs EqConstraintSet) Add(c... EqualityConstr) EqConstraintSet {
	return append(cs, c...)
}

func (cs EqConstraintSet) UnifyAll(ds []Decl, delta Delta) EtaOpen {
	if len(cs) == 0 {
		return EtaOpen{}
	}
	c, cs_ := cs[0], cs[1:]
	eta := c.Unify(ds, delta)
	cs_ = cs_.SubsEtaOpen(eta)
	return compose(cs_.UnifyAll(ds, delta), eta)
	// se unifyAll retornasse (EtaOpen, error), o que quereria fazer é algo como
	// unifyAll(cs.SubsEtaOpen(subs), delta) >>= \s' -> compose(s', subs)
}

// Set of subtype/impls (<:) constraints
// note: assumes that the constraints are unifiable "1 by 1" -- C.f. UnifyAll and compose
type SubConstraintSet []SubtypeConstr

func NewSubConstraintSet() SubConstraintSet {
	return make(SubConstraintSet, 0, 10)
}

func (cs SubConstraintSet) SubsEtaOpen(eta EtaOpen) SubConstraintSet {
	for i, c := range cs {
		cs[i] = c.SubsEtaOpen(eta)
	}
	return cs
}

func (cs SubConstraintSet) Add(c... SubtypeConstr) SubConstraintSet {
	return append(cs, c...)
}

func (cs SubConstraintSet) UnifyAll(ds []Decl, delta Delta) EtaOpen {
	if len(cs) == 0 {
		return EtaOpen{}
	}
	c, cs_ := cs[0], cs[1:]
	eta := c.Unify(ds, delta)
	cs_ = cs_.SubsEtaOpen(eta)
	return compose(cs_.UnifyAll(ds, delta), eta)
	// se unifyAll retornasse (EtaOpen, error), o que quereria fazer é algo como
	// unifyAll(cs.SubsEtaOpen(subs), delta) >>= \s' -> compose(s', subs)
}

// s1 `compose` s2
func compose(s1, s2 EtaOpen) EtaOpen {
	res := make(EtaOpen)
	for tParam, u := range s2 {
		res[tParam] = u.SubsEtaOpen(s1)
	}
	// res `union` s1
	for tParam, u := range s1 {
		if u2, present := res[tParam]; present {
			// try to find mapping to most general type todo or should it be to the most specific type? e.g. MyInt <: Any()
			if u.ImplsDelta([]Decl{}, make(Delta), u2) {
				res[tParam] = u2
			} else if u2.ImplsDelta([]Decl{}, make(Delta), u) {
				res[tParam] = u
			} else {
				panic("Incompatible mappings for " + tParam.String() + ": " +
					u.String() + " and " + u2.String())
			}
		} else {
			res[tParam] = u
		}
	}
	// pode haver 2 mapeamentos diferentes para a mesma var?
	// i.e. há s1[T] e s2[T], mas s1[T] != s2[T]
	// se sim, quem tem prioridade?

	// ver https://www.cs.tufts.edu/comp/150FP/archive/benjamin-pierce/lti-toplas.pdf p. 32(final)-33
	// isto pode correr mal nesse caso?
	return res
}

/* Unification of a Subtype/Equality constraint */

func bindSub(x FreshTVar, u Type) EtaOpen {
	if x.Equals(u) {
		return EtaOpen{}
	} else if occursCheck(x, u) {
		panic("Cannot construct infinite type: " + x.String() + " ~ " + u.String())
	} else {
		// if u1.ImplsDelta( delta[u2_cast] ) ... todo check bounds
		return EtaOpen{x.TParam: u}
	}
}

// todo are these 2 binds different? If they are, it is due to the way the bounds are checked
func bindEq(x FreshTVar, u Type) EtaOpen {
	if occursCheck(x, u) {
		panic("Cannot construct infinite type: " + x.String() + " ~ " + u.String())
	} else {
		// if u1.ImplsDelta( delta[u2_cast] ) ... todo check bounds
		return EtaOpen{x.TParam: u}
	}
}

// the goal is to find a substitution eta s.t. u1[eta] <: u2[eta]
func (c SubtypeConstr) Unify(ds []Decl, delta Delta) EtaOpen {
	u1 := c.u1
	u2 := c.u2

	if u1_cast, ok := u1.(FreshTVar); ok {
		return bindSub(u1_cast, u2)
	}
	if u2_cast, ok := u2.(FreshTVar); ok {
		return bindSub(u2_cast, u1) // todo when considering bounds, will this test that bound(u2) <: bound(u1) ??
	}

	u1_named, ok1 := u1.(TNamed)
	u2_named, ok2 := u2.(TNamed)
	if ok1 && ok2 {
		if u1_named.t_name == u2_named.t_name {

			// TODO aqui nao devia recolher também equality constraints??
			//  ou sera que e.g. S <: T  =>  List(S) <: List(T)  ---> para qualquer "tipo" List?
			// return NewEqualityConstr(u1, u2).Unify(ds, delta)   <---- s

			// unify vars of u2 with args of u1
			constrs := NewSubConstraintSet()
			for i, u_arg := range u2_named.u_args {
				if hasFreshTVars(u_arg) {
					c := NewSubtypeConstr(u1_named.u_args[i], u_arg) // TODO should I be collecting constraints inside unify?
					constrs = constrs.Add(c)                        //   Or maybe add that logic to a method AddConstraints that searches for name-matching TNameds?
				} else if !u1_named.u_args[i].ImplsDelta(ds, delta, u_arg) {
					panic("")
				}
			}
			return constrs.UnifyAll(ds, delta)

		} else {
			ms_t1 := methodsDelta(ds, delta, u1)
			ms_t2 := methodsDelta(ds, delta, u2)
			return unifyMethods(ds, delta, ms_t1, ms_t2)
		}
	}
	// either u1 or u2 not a TNamed

	if u1.ImplsDelta(ds, delta, u2) {
		return EtaOpen{}
	} else {
		panic("Can't unify (<:) types " + u1.String() + " and " + u2.String())
	}

}

func (c EqualityConstr) Unify(ds []Decl, delta Delta) EtaOpen {
	u1 := c.u1
	u2 := c.u2

	if u1_cast, ok := u1.(FreshTVar); ok {
		return bindEq(u1_cast, u2)
	}
	if u2_cast, ok := u2.(FreshTVar); ok {
		return bindEq(u2_cast, u1)
	}

	u1_cast, ok1 := u1.(TNamed)
	u2_cast, ok2 := u2.(TNamed)
	if ok1 && ok2 && u1_cast.t_name == u2_cast.t_name {
		constrs := NewEqConstraintSet()
		for i, u_arg := range u2_cast.u_args {
			c := NewEqualityConstr(u1_cast.u_args[i], u_arg) // TODO should I be collecting constraints inside unify?
			constrs = constrs.Add(c)

			//if hasFreshTVars(u_arg) {
			//	c := NewEqualityConstr(u1_cast.u_args[i], u_arg)
			//	constrs = constrs.Add(c)
			//} else if !u_arg.Equals(u1_cast.u_args[i]) {
			//	panic("")
			//}
		}
		return constrs.UnifyAll(ds, delta)
	}
	// either u1 or u2 not a TNamed
	// TODO consider untyped constants here
	if u1.Equals(u2) {
		return EtaOpen{}
	} else {
		panic("Can't unify (==) types " + u1.String() + " and " + u2.String())
	}
}

// On successful unification, returns eta
// s.t. ms1 [is a superset of/at least equal to] ms2[eta]
func unifyMethods(ds []Decl, delta Delta, ms1, ms2 MethodSet) EtaOpen {
	constrs := NewEqConstraintSet()
	for name, sig2 := range ms2 {
		sig1, ok := ms1[name]
		if !ok {
			panic("Cant make ms2 a subset of ms1 -- method " + name + " of ms2 not present in ms1")
		}
		if len(sig1.Psi.tFormals) != len(sig2.Psi.tFormals) || len(sig1.pDecls) != len(sig2.pDecls) {
			panic("Can't unify signatures:\n\tsig1 = " + sig1.String() + "\n\tsig2 = " + sig2.String())
		}
		sigConstrs := collectSigConstrs(sig1, sig2)
		constrs = constrs.Add(sigConstrs...)
	}
	return constrs.UnifyAll(ds, delta)
}

func collectSigConstrs(g1, g2 Sig) EqConstraintSet {
	subs1 := makeParamIndexSubs(g1.Psi) // todo maybe factor this + subsEtaOpen as e.g. canonicalizeSig
	subs2 := makeParamIndexSubs(g2.Psi)
	sig1 := g1.SubsEtaOpen(subs1)
	sig2 := g2.SubsEtaOpen(subs2)

	constrs := NewEqConstraintSet()
	for i, tf1 := range sig1.Psi.tFormals {
		c := NewEqualityConstr(tf1.u_I, sig2.Psi.tFormals[i].u_I)
		constrs = constrs.Add(c)
	}
	for i, pd1 := range sig1.pDecls {
		c := NewEqualityConstr(pd1.u, sig2.pDecls[i].u)
		constrs = constrs.Add(c)
	}
	return constrs.Add(NewEqualityConstr(sig1.u_ret, sig2.u_ret))
}

/******************************************************************************/
/* Inference of expressions' types */

func (x Variable) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	res, ok := gamma[x.name]
	if !ok {
		panic("Var not in env: " + x.String())
	}
	return res
}

func (s StructLit) Infer(ds []Decl, delta Delta, gamma Gamma) Type {

	td := getTDecl(ds, s.u_S.GetName()) // panics if not found
	u_S := instantiateType(s.u_S.GetName(), td.GetBigPsi())

	fs := fields(ds, u_S)
	if len(s.elems) != len(fs) {
		var b strings.Builder
		b.WriteString("Arity mismatch: args=[")
		writeExprs(&b, s.elems)
		b.WriteString("], fields=[")
		writeFieldDecls(&b, fs)
		b.WriteString("]\n\t")
		b.WriteString(s.String())
		panic(b.String())
	}
	//elems := make([]FGGExpr, len(s.elems))
	constraints := NewSubConstraintSet()
	for i := 0; i < len(s.elems); i++ {
		u := s.elems[i].Infer(ds, delta, gamma)
		constr := NewSubtypeConstr(u, fs[i].u)
		constraints = constraints.Add(constr)

		//elems[i] = newSubtree
		//// if newSubtree is a PrimitiveLiteral node, convert it to the Ast node
		//// corresponding to a value of the expected type (r)
		//if lit, ok := newSubtree.(PrimitiveLiteral); ok {
		//	elems[i] = ConvertLitNode(lit, u_decl)
		//}
	}
	subs := constraints.UnifyAll(ds, delta)

	return u_S.SubsEtaOpen(subs) //, StructLit{s.u_S, elems}
}

func (c Call) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	u_recv := c.e_recv.Infer(ds, delta, gamma)
	var g Sig
	//if tmp, ok := methodsDelta(ds, delta, bounds(delta, u_recv))[c.meth]; !ok { // !!! submission version had "methods(m)"
	if tmp, ok := methods(ds, u_recv)[c.meth]; !ok { // !!! submission version had "methods(m)"
		panic("Method not found: " + c.meth + " in " + u_recv.String())
	} else {
		g = tmp
	}
	if len(c.args) != len(g.pDecls) {
		var b strings.Builder
		b.WriteString("Arity mismatch: args=[")
		writeExprs(&b, c.args)
		b.WriteString("], params=[")
		writeParamDecls(&b, g.pDecls)
		b.WriteString("]\n\t")
		b.WriteString(c.String())
		panic(b.String())
	}
	//args := make([]FGGExpr, len(c.args))

	sigInst := instantiateSig(g)

	constraints := NewSubConstraintSet()
	for i := 0; i < len(c.args); i++ {
		u_a := c.args[i].Infer(ds, delta, gamma)

		constr := NewSubtypeConstr(u_a, sigInst.pDecls[i].u)
		constraints = constraints.Add(constr)

		//args[i] = newSubtree
		//// if newSubtree is a PrimitiveLiteral node, convert it to the Ast node
		//// corresponding to a value of the expected type (u_p)
		//if lit, ok := newSubtree.(PrimitiveLiteral); ok {
		//	args[i] = ConvertLitNode(lit, u_p)
		//}
	}
	subs := constraints.UnifyAll(ds, delta)
	return sigInst.u_ret.SubsEtaOpen(subs) //, Call{e_recv, c.meth, c.t_args, args}
}

func (s Select) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	u := s.e_S.Infer(ds, delta, gamma)
	if !IsStructType(ds, u) {
		panic("Illegal select on expr of non-struct type: " + u.String() +
			"\n\t" + s.String())
	}
	u_S := u.Underlying(ds).(STypeLit)
	for _, fd := range u_S.GetFieldDecls() {
		if fd.field == s.field {
			return fd.u //, Select{e_S, s.field}
		}
	}
	panic("Field " + s.field + " not found in type: " + u.String() +
		"\n\t" + s.String())
}

func (a Assert) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	return a.u_cast // TODO
}

func (x Sprintf) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	// todo type the arguments, return String type
	return TPrimitive{tag: STRING}
}

func (b BinaryOperation) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	ltype := b.left.Infer(ds, delta, gamma)
	rtype := b.right.Infer(ds, delta, gamma)

	var pred PrimtPredicate
	switch b.op {
	case ADD:
		pred = Or(isNumeric, isString)
	case SUB:
		pred = isNumeric
	case LAND, LOR:
		pred = isBool
	}
	if ok := evalPrimtPredicate(ds, delta, pred, ltype); !ok {
		panic("operator " + string(b.op) + " not defined for type: " + ltype.String())
	}
	if ok := evalPrimtPredicate(ds, delta, pred, rtype); !ok {
		panic("operator " + string(b.op) + " not defined for type: " + rtype.String())
	}

	// verify that ltype and rtype are compatible;
	// if they are, return the most general type
	if ltype.ImplsDelta(ds, delta, rtype) {
		return rtype //, newTree
	}
	if rtype.ImplsDelta(ds, delta, ltype) {
		return ltype //, newTree
	}
	panic("mismatched types " + ltype.String() + " and " + rtype.String())
}

func (c Comparison) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	ltype := c.left.Infer(ds, delta, gamma)
	rtype := c.right.Infer(ds, delta, gamma)

	if ok := evalPrimtPredicate(ds, delta, isComparable, ltype); !ok {
		panic("operator " + string(c.op) + " not defined for type: " + ltype.String())
	}
	if ok := evalPrimtPredicate(ds, delta, isComparable, rtype); !ok {
		panic("operator " + string(c.op) + " not defined for type: " + rtype.String())
	}

	if !ltype.ImplsDelta(ds, delta, rtype) && !rtype.ImplsDelta(ds, delta, ltype) {
		panic("mismatched types " + ltype.String() + " and " + rtype.String())
	}

	return TPrimitive{tag: BOOL} //, NewBinaryOp(ltree, rtree, c.op)
}

func (x NamedPrimitiveLiteral) Infer(ds []Decl, delta Delta, gamma Gamma) Type { return x.typ }
func (x PrimitiveLiteral) Infer(ds []Decl, delta Delta, gamma Gamma) Type {
	return NewUndefTPrimitive(x.tag)
}
func (x BoolVal) Infer(ds []Decl, delta Delta, gamma Gamma) Type    { return TPrimitive{tag: BOOL} }
func (x Int32Val) Infer(ds []Decl, delta Delta, gamma Gamma) Type   { return TPrimitive{tag: INT32} }
func (x Int64Val) Infer(ds []Decl, delta Delta, gamma Gamma) Type   { return TPrimitive{tag: INT64} }
func (x Float32Val) Infer(ds []Decl, delta Delta, gamma Gamma) Type { return TPrimitive{tag: FLOAT32} }
func (x Float64Val) Infer(ds []Decl, delta Delta, gamma Gamma) Type { return TPrimitive{tag: FLOAT64} }
func (x StringVal) Infer(ds []Decl, delta Delta, gamma Gamma) Type  { return TPrimitive{tag: STRING} }

/******************************************************************************/
/* Fresh type variables */

// +- same distinction as between type vars in a Type scheme (TParam)
// and fresh/free Type variables
type FreshTVar struct {
	//x TParam
	TParam
	bound Type
}

var _ Type = FreshTVar{}

func newFreshTVar(name TParam, bound Type) FreshTVar {
	return FreshTVar{name, bound}
}

func (tv FreshTVar) SubsEtaOpen(eta EtaOpen) Type {
	res, ok := eta[tv.TParam]
	if !ok {
		return tv
	}
	return res
}

func (tv FreshTVar) Underlying(ds []Decl) Type {
	return tv
}

// Adds recorded bound to Delta before calling the normal TParam.ImplsDelta.
// Needed because the bound for a fresh type var may not be in context
// e.g., inferring the type of the empty List results in a Nil(ααX),
//       but since Infer doesn't return a Delta, the context for ααX is lost.
func (tv FreshTVar) ImplsDelta(ds []Decl, delta Delta, u Type) bool {
	extendedDelta := make(Delta)
	for param, bound := range delta {
		extendedDelta[param] = bound
	}
	extendedDelta[tv.TParam] = tv.bound
	return tv.TParam.ImplsDelta(ds, extendedDelta, u)
}

const FreshPrefix = "αα"
var freshCount = 0 // global var
func freshName() TParam {
	res := TParam(FreshPrefix + strconv.Itoa(freshCount+1))
	freshCount++
	return res
}

func isFreshTVar(u Type) bool {
	_, ok := u.(FreshTVar)
	return ok
}

func hasFreshTVars(u Type) bool {
	if isFreshTVar(u) {
		return true
	}
	if u_cast, ok := u.(TNamed); ok {
		for _, u_arg := range u_cast.u_args {
			if hasFreshTVars(u_arg) {
				return true
			}
		}
	}
	return false
}

func instantiatePsi(Psi BigPsi) map[TParam]FreshTVar {
	subs := make(map[TParam]Type)
	for _, tf := range Psi.GetTFormals() {
		subs[tf.name] = freshName()
	}
	resSubs := make(map[TParam]FreshTVar)
	for _, tf := range Psi.GetTFormals() {
		x := subs[tf.name].(TParam)
		bound := tf.u_I.SubsEtaOpen(subs)
		resSubs[tf.name] = newFreshTVar(x, bound)
	}
	return resSubs
}

func instantiateType(tname Name, Psi BigPsi) TNamed {
	//subs := make(EtaOpen)
	//for _, tf := range Psi.GetTFormals() {
	//	subs[tf.name] = freshName()
	//}
	insts := instantiatePsi(Psi)

	args := make([]Type, len(Psi.GetTFormals()))
	for i, tf := range Psi.GetTFormals() {
		args[i] = insts[tf.name]
	}
	return NewTNamed(tname, args)
}

func instantiateSig(sig Sig) Sig {
	//subs := make(map[TParam]Type)
	//for _, tf := range sig.Psi.tFormals {
	//	subs[tf.name] = freshName()
	//}
	//delta := make(Delta)
	//for _, tf := range sig.Psi.tFormals {
	//	x := subs[tf.name].(TParam)
	//	delta[x] = tf.u_I.SubsEtaOpen(subs)
	//}

	insts := instantiatePsi(sig.Psi)
	// "cast" insts to EtaOpen to take advantage of existing .SubsEtaOpen "infrastructure"
	eta := make(EtaOpen)
	for param, tVar := range insts {
		eta[param] = tVar
	}

	// subs in param declarations' types
	ps := make([]ParamDecl, len(sig.pDecls))
	for i, pd := range sig.pDecls {
		ps[i] = ParamDecl{pd.name, pd.u.SubsEtaOpen(eta)}
	}
	// subs in u_ret
	u_ret := sig.u_ret.SubsEtaOpen(eta)
	return Sig{sig.meth,
		BigPsi{}, // todo all the type parameters were instantiated (?)
		ps, u_ret}
}

func occursCheck(a FreshTVar, u Type) bool {
	for _, fvar := range ftvs(u) {
		if a.Equals(fvar) {
			return true
		}
	}
	return false
}

// duplicates fv in fgg_ismonom, only changes the return type
func ftvs(u Type) []FreshTVar {
	if cast, ok := u.(FreshTVar); ok {
		return []FreshTVar{cast}
	}
	res := []FreshTVar{}
	if cast, ok := u.(TNamed); ok {
		for _, v := range cast.u_args {
			res = append(res, ftvs(v)...)
		}
	}
	return res
}
