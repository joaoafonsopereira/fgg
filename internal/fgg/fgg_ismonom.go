package fgg

import (
	"fmt"
	"reflect"
	"strings"
)

var _ = fmt.Errorf
var _ = reflect.Append
var _ = strings.Compare

// Return true if *not* nomono
func IsMonomOK(p FGGProgram) (bool, string) {
	ds := p.GetDecls()
	for _, v := range ds {
		if md, ok := v.(MethDecl); ok {
			omega := Nomega{make(map[string]Type), make(map[string]MethInstanOpen)}
			delta := md.Psi_recv.ToDelta()
			for _, tf := range md.Psi_meth.tFormals {
				delta[tf.name] = tf.u_I
			}
			gamma := make(Gamma)
			psi_recv := make(SmallPsi, len(md.Psi_recv.tFormals))
			for i, tf := range md.Psi_recv.tFormals {
				psi_recv[i] = tf.name
			}
			//psi_recv = md.Psi_recv.Hat()
			u_recv := TNamed{md.t_recv, psi_recv}
			gamma[md.x_recv] = u_recv
			omega.us[tokeyWtOpen(u_recv)] = u_recv
			for _, pd := range md.pDecls { // TODO: factor out
				gamma[pd.name] = pd.u
			}
			collectExprOpen(ds, delta, gamma, omega, md.e_body)
			if ok, msg := nomonoOmega(ds, delta, md, omega); ok {
				return false, msg
			}
		}
	}
	return true, ""
}

// Return true if nomono
func nomonoOmega(ds []Decl, delta Delta, md MethDecl, omega Nomega) (bool, string) {
	for auxGOpen(ds, delta, omega) {
		for _, v := range omega.ms {
			u_S, ok :=  v.u_recv.(TNamed)
			if !ok || isIfaceType(ds, u_S) {
				continue
			}
			if u_S.t_name == md.t_recv && v.meth == md.name {
				if occurs(md.Psi_recv, u_S.u_args) {
					return true, md.t_recv + md.Psi_recv.String() + " ->* " + md.t_recv +
						"(" + SmallPsi(u_S.u_args).String() + ")"
				}
				if occurs(md.Psi_meth, v.psi) {
					return true, md.t_recv + md.Psi_recv.String() + "." + md.name +
						md.Psi_meth.String() + " ->* " + md.name + "(" + v.psi.String() + ")"
				}
			}
		}
	}
	return false, ""
}

// Pre: len(Psi) == len(psi)
func occurs(Psi BigPsi, psi SmallPsi) bool {
	for i, v := range Psi.tFormals {
		if cast, ok := psi[i].(TNamed); ok { // !!! simplified
			for _, x := range fv(cast) {
				if x.Equals(v.name) {
					return true
				}
			}
		}
	}
	return false
}

func fv(u Type) []TParam {
	if cast, ok := u.(TParam); ok {
		return []TParam{cast}
	}
	res := []TParam{}
	cast := u.(TNamed)
	for _, v := range cast.u_args {
		res = append(res, fv(v)...)
	}
	return res
}

/* !!! Duplication of Omega (etc.) for non-ground types -- if only Go had generics! */

type Nomega struct {
	us map[string]Type
	ms map[string]MethInstanOpen
}

func (w Nomega) clone() Nomega {
	us := make(map[string]Type)
	ms := make(map[string]MethInstanOpen)
	for k, v := range w.us {
		us[k] = v
	}
	for k, v := range w.ms {
		ms[k] = v
	}
	return Nomega{us, ms}
}

func (w Nomega) Println() {
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

// Factor out with MethInstan
type MethInstanOpen struct {
	u_recv Type
	meth   Name
	psi    SmallPsi
}

func tokeyWtOpen(u Type) string {
	return u.String()
}

func tokeyWmOpen(x MethInstanOpen) string {
	return x.u_recv.String() + "_" + x.meth + "_" + x.psi.String()
}

func collectExprOpen(ds []Decl, delta Delta, gamma Gamma, omega Nomega, e FGGExpr) bool {
	res := false
	switch e1 := e.(type) {
	case Variable:
		return res
	case StructLit:
		res = collectExprsOpen(ds, delta, gamma, omega, e1.elems...)
		k := tokeyWtOpen(e1.u_S)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = e1.u_S
			res = true
		}
	case Select:
		return collectExprOpen(ds, delta, gamma, omega, e1.e_S)
	case Call:
		res = collectExprOpen(ds, delta, gamma, omega, e1.e_recv)
		res = collectExprsOpen(ds, delta, gamma, omega, e1.args...) || res
		gamma1 := make(Gamma)
		for k, v := range gamma {
			gamma1[k] = v
		}
		u_recv, _ := e1.e_recv.Typing(ds, delta, gamma1, false)
		k_t := tokeyWtOpen(u_recv)
		if _, ok := omega.us[k_t]; !ok {
			omega.us[k_t] = u_recv
			res = true
		}
		m := MethInstanOpen{u_recv, e1.meth, e1.GetTArgs()} // CHECKME: why add u_recv separately?
		k_m := tokeyWmOpen(m)
		if _, ok := omega.ms[k_m]; !ok {
			omega.ms[k_m] = m
			res = true
		}
	case Assert:
		res = collectExprOpen(ds, delta, gamma, omega, e1.e_I)
		u := e1.u_cast
		k := tokeyWtOpen(u)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = u
			res = true
		}
	case Convert:
		res = collectExprOpen(ds, delta, gamma, omega, e1.expr)
		u := e1.typ
		k := tokeyWtOpen(u)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = u
			res = true
		}
	case Sprintf:
		res = collectExprsOpen(ds, delta, gamma, omega, e1.args...)

	case BinaryOperation:
		res = collectExprsOpen(ds, delta, gamma, omega, e1.left, e1.right)
	case Comparison:
		res = collectExprsOpen(ds, delta, gamma, omega, e1.left, e1.right)

	// This may be needed to handle instances of types declared like:
	// type MyInt[T any] int32
	case TypedPrimitiveValue:
		if u_N, ok := e1.typ.(TNamed); ok {
			k := tokeyWtOpen(u_N)
			if _, ok := omega.us[k]; !ok {
				omega.us[k] = u_N
				res = true
			}
		}
	case PrimitiveLiteral:
		return res

	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e).String() + "\n\t" +
			e.String())
	}
	return res
}

func collectExprsOpen(ds []Decl, delta Delta, gamma Gamma, omega Nomega, es ...FGGExpr) bool {
	res := false
	for _, arg := range es {
		res = collectExprOpen(ds, delta, gamma, omega, arg) || res
	}
	return res
}

/* Aux */

// Return true if omega has changed
// N.B. no closure over types occurring in bounds, or *interface decl* method sigs
func auxGOpen(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	res = auxFOpen(ds, omega) || res
	res = auxIOpen(ds, delta, omega) || res
	res = auxMOpen(ds, delta, omega) || res
	res = auxSOpen(ds, delta, omega) || res
	// I/face embeddings
	res = auxE1Open(ds, omega) || res
	res = auxE2Open(ds, omega) || res
	// type declarations
	res = auxTOpen(ds, omega) || res
	return res
}

func auxTOpen(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]Type)
	for _, u := range omega.us {
		u_N, isTNamed := u.(TNamed)
		if !isTNamed {
			continue
		}
		td := getTDecl(ds, u_N.t_name)
		if src, ok := td.GetSourceType().(TNamed); ok {
			eta := MakeEtaOpen(td.GetBigPsi(), u_N.GetTArgs())
			srcInst := src.SubsEtaOpen(eta)
			tmp[tokeyWtOpen(srcInst)] = srcInst
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

func auxFOpen(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]Type)
	for _, u := range omega.us {
		if u_S, ok := u.Underlying(ds).(STypeLit); ok {
			for _, fd := range u_S.GetFieldDecls() {
				tmp[tokeyWtOpen(fd.u)] = fd.u
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

func auxIOpen(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstanOpen)
	for _, m := range omega.ms {
		if !isIfaceType(ds, m.u_recv) { // want both named and anonymous ifaces
			continue
		}
		// more precise to iterate us: might have iface types that were only
		// added as type instances, without "mapping" to a method inst
		// (e.g. only appearing in a struct field)
		for _, u := range omega.us {
			if !isIfaceType(ds, u) || u.Equals(m.u_recv) {
				continue
			}
			if ImplsDelta(ds, delta, u, getInterface(ds, m.u_recv)) {
				mm := MethInstanOpen{u, m.meth, m.psi}
				tmp[tokeyWmOpen(mm)] = mm
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

func auxMOpen(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]Type)
	for _, m := range omega.ms {
		sig := methodsDelta(ds, delta, m.u_recv)[m.meth]
		eta := MakeEtaOpen(sig.Psi, m.psi)
		for _, pd := range sig.pDecls {
			u_pd := pd.u.SubsEtaOpen(eta)
			tmp[tokeyWtOpen(u_pd)] = u_pd
		}
		u_ret := sig.u_ret.SubsEtaOpen(eta)
		tmp[tokeyWtOpen(u_ret)] = u_ret
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

func auxSOpen(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstanOpen)
	clone := omega.clone()
	for _, m := range clone.ms {
		for _, u := range clone.us {
			u_recv := bounds(delta, m.u_recv) // !!! cf. plain type param
			u_N, ok := u.(TNamed)
			if !ok || isIfaceType(ds, u_N) {
				continue
			}
			assignable, _ := u_N.AssignableToDelta(ds, delta, u_recv)
			if !assignable {
				continue
			}

			x0, xs, e := body(ds, u_N, m.meth, m.psi)
			gamma := make(Gamma)
			gamma[x0.name] = x0.u
			for _, pd := range xs {
				gamma[pd.name] = pd.u
			}
			m1 := MethInstanOpen{u_N, m.meth, m.psi}
			k := tokeyWmOpen(m1)
			//if _, ok := omega.ms[k]; !ok { // No: initial collectExpr already adds to omega.ms
			tmp[k] = m1
			res = collectExprOpen(ds, delta, gamma, omega, e) || res
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
func auxE1Open(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]TNamed)
	for _, u := range omega.us {
		if u_I, ok := u.Underlying(ds).(ITypeLit); ok {
			for _, s := range u_I.specs {
				if u_emb, ok := s.(TNamed); ok {
					tmp[tokeyWtOpen(u_emb)] = u_emb
				}
			}
		} //else if todo CHECKME: type param -------------------------------------------

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
func auxE2Open(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstanOpen)
	for _, m := range omega.ms {
		if !isIfaceType(ds, m.u_recv) {
			continue
		}
		u_I := getInterface(ds, m.u_recv)
		for _, s := range u_I.GetSpecs() {
			if u_emb, ok := s.(TNamed); ok {
				if _, hasMeth := methods(ds, u_emb)[m.meth]; hasMeth {
					m_emb := MethInstanOpen{u_emb, m.meth, m.psi}
					tmp[tokeyWmOpen(m_emb)] = m_emb
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

/*

















































 */

/*
 * Deprecated: old CFG-based test
 */

//type RecvMethPair struct {
//	t_recv Name // Pre: t_S
//	m      Name // TODO rename
//}
//
//func (x0 RecvMethPair) equals(x RecvMethPair) bool {
//	return x0.t_recv == x.t_recv && x0.m == x.m
//}
//
//type cTypeArgs struct {
//	psi_recv SmallPsi
//	psi_meth SmallPsi
//}
//
//func (x0 cTypeArgs) equals(x cTypeArgs) bool {
//	return x0.psi_recv.Equals(x.psi_recv) && x0.psi_meth.Equals(x.psi_meth)
//}
//
//// Static call graph, agnostic of specific type args (cf. MethInstan)
//// N.B. nodes are for struct types
//type cgraph struct {
//	edges map[RecvMethPair]map[RecvMethPair]([]cTypeArgs)
//}
//
//func (x0 cgraph) String() string {
//	var b strings.Builder
//	for k, v := range x0.edges {
//		b.WriteString(k.t_recv)
//		b.WriteString(".")
//		b.WriteString(k.m)
//		b.WriteString(": ")
//		b.WriteString(fmt.Sprintf("%v", v))
//		b.WriteString("\n")
//	}
//	return b.String()
//}
//
//// CFG-based occurs check -- needs method set unification ("open" impls)
//// CHECKME: generally, covariant receiver bounds specialisation
//func IsMonomOK_CFG(p FGGProgram) bool {
//	ds := p.GetDecls()
//	graph := cgraph{make(map[RecvMethPair]map[RecvMethPair]([]cTypeArgs))}
//	for _, v := range ds {
//		if md, ok := v.(MethDecl); ok {
//			buildGraph(ds, md, graph)
//		}
//	}
//	//buildGraphExpr(ds, make(Delta), make(Gamma), ...)  // visit main unnecessary -- CHECKME: all type instans seen?
//	//fmt.Println("111:\n", graph.String(), "---")
//	cycles := make(map[string]cycle)
//	findCycles(graph, cycles)
//	/*for _, v := range cycles {
//		fmt.Println("aaa:", v)
//	}*/
//	for _, v := range cycles {
//		//fmt.Println("bbb:", v)
//		if isNomonoCycle(ds, graph, v) {
//			return false
//		}
//		return true
//	}
//	return true
//}
//
//// Occurs check -- N.B. conservative w.r.t. whether type params actually used
//func isNomonoCycle(ds []Decl, graph cgraph, c cycle) bool {
//	for _, tArgs := range graph.edges[c[0]][c[1]] {
//		if isNomonoTypeArgs(tArgs) || isNomonoCycleAux(ds, graph, c, tArgs, 1) {
//			return true
//		}
//	}
//	return false
//}
//
//func isNomonoTypeArgs(tArgs cTypeArgs) bool {
//	for _, v := range tArgs.psi_recv {
//		if containsNestedTParam(v) {
//			return true
//		}
//	}
//	for _, v := range tArgs.psi_meth {
//		if containsNestedTParam(v) {
//			return true
//		}
//	}
//	return false
//}
//
//func isNomonoCycleAux(ds []Decl, graph cgraph, c cycle, tArgs cTypeArgs, i int) bool {
//	if i >= (len(c) - 1) {
//		return false
//	}
//	next := c[i]
//	md := getMethDecl(ds, next.t_recv, next.m)
//	subs := make(Delta)
//	for i, v := range tArgs.psi_recv {
//		subs[md.Psi_recv.tFormals[i].name] = v
//	}
//	for i, v := range tArgs.psi_meth {
//		subs[md.Psi_meth.tFormals[i].name] = v
//	}
//
//	for _, v := range graph.edges[c[i]][c[i+1]] {
//		tArgs1 := cTypeArgs{v.psi_recv.TSubs(subs), v.psi_meth.TSubs(subs)}
//		if isNomonoTypeArgs(tArgs1) {
//			return true
//		}
//		isNomonoCycleAux(ds, graph, c, tArgs1, i+1)
//	}
//	return false
//}
//
//func containsNestedTParam(u Type) bool {
//	if cast, ok := u.(TNamed); ok {
//		for _, v := range cast.u_args {
//			if isOrContainsTParam(v) {
//				return true
//			}
//		}
//	}
//	return false
//}
//
//type cycle []RecvMethPair
//
//func (x0 cycle) toHash() string {
//	return fmt.Sprintf("%v", x0)
//}
//
//func findCycles(graph cgraph, cycles map[string]cycle) {
//	for k, _ := range graph.edges {
//		stack := []RecvMethPair{k}
//		findCyclesAux(graph, stack, cycles)
//	}
//}
//
//// DFS -- TODO: start from main more efficient? -- CHECKME: maybe more "correct", w.r.t. omega method discarding
//func findCyclesAux(graph cgraph, stack []RecvMethPair, cycles map[string]cycle) {
//	targets := graph.edges[stack[len(stack)-1]]
//	if targets == nil {
//		panic("Shouldn't get in here:")
//	}
//lab:
//	for next, _ := range targets {
//		stack1 := append(stack, next)
//		if stack1[0].equals(next) {
//			cycles[cycle(stack1).toHash()] = stack1
//			continue
//		}
//		for _, prev := range stack[1:] {
//			if prev.equals(next) {
//				continue lab
//			}
//		}
//		findCyclesAux(graph, stack1, cycles)
//	}
//}
//
//// "Flat" graph building -- calls not visited (i.e., `body` not used)
//// Output: mutates cgraph
//func buildGraph(ds []Decl, md MethDecl, graph cgraph) {
//	n := RecvMethPair{md.t_recv, md.name}
//	graph.edges[n] = make(map[RecvMethPair]([]cTypeArgs))
//	delta := md.Psi_meth.ToDelta() // recv params added below
//	gamma := make(Gamma)
//	psi := make(SmallPsi, len(md.Psi_recv.tFormals))
//	for i, v := range md.Psi_recv.tFormals {
//		delta[v.name] = v.u_I
//		psi[i] = v.name
//	}
//	gamma[md.x_recv] = TNamed{md.t_recv, psi}
//	for _, v := range md.pDecls { // TODO: factor out
//		gamma[v.name] = v.u
//	}
//	buildGraphExpr(ds, delta, gamma, n, md.e_body, graph)
//}
//
//// "Flat" graph building -- calls not visited (i.e., `body` not used)
//func buildGraphExpr(ds []Decl, delta Delta, gamma Gamma, curr RecvMethPair, e1 FGGExpr, graph cgraph) {
//	switch e := e1.(type) {
//	case Variable:
//	case StructLit:
//		for _, elem := range e.elems {
//			buildGraphExpr(ds, delta, gamma, curr, elem, graph)
//		}
//	case Select:
//		buildGraphExpr(ds, delta, gamma, curr, e.e_S, graph)
//	case Call:
//		buildGraphExpr(ds, delta, gamma, curr, e.e_recv, graph)
//		for _, arg := range e.args {
//			buildGraphExpr(ds, delta, gamma, curr, arg, graph)
//		}
//		u_recv, _ := e.e_recv.Typing(ds, delta, gamma, true)
//
//		if isStructType(ds, u_recv) { // u_recv is a TNamed struct
//			u_S := u_recv.(TNamed)
//			putTArgs(graph, curr, u_S, e.meth, e.t_args)
//
//		} else { // TNamed interface or TParam
//			u_I := u_recv // Or type param
//			if _, ok := u_I.(TParam); ok {
//				u_I = u_I.TSubs(delta) // CHECKME
//			}
//			//for _, v := range ds {
//			//	if d, ok := v.(STypeLit); ok {
//			//
//			//		// method set unification instead of basic impls? -- or using bounds (hat) sufficient?
//			//		u_S := TNamed{d.t_name, d.Psi.Hat()} // !!!
//			//		if u_S.ImplsDelta(ds, delta, u_I) {
//			//			putTArgs(graph, curr, u_S, e.meth, e.t_args)
//			//		}
//			//
//			//	}
//			//}
//		}
//	case Assert:
//		buildGraphExpr(ds, delta, gamma, curr, e.e_I, graph)
//	default:
//		panic("Unknown Expr kind: " + reflect.TypeOf(e1).String() + "\n\t" +
//			e1.String())
//	}
//}
//
//// u_recv is target u_S
//func putTArgs(graph cgraph, curr RecvMethPair, u_recv TNamed, meth Name, psi_meth SmallPsi) {
//	edges := graph.edges[curr]
//	/*if edges == nil {
//		edges = make(map[node]([]cTypeArgs))
//		graph.edges[curr] = edges
//	}*/
//	target := RecvMethPair{u_recv.t_name, meth}
//	tArgs := edges[target]
//	if tArgs == nil {
//		tArgs = []cTypeArgs{}
//	}
//	tArgs = append(tArgs, cTypeArgs{u_recv.u_args, psi_meth})
//	edges[target] = tArgs
//}
