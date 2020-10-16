/*
 * TODO: fix many magic numbers and other sloppy hacks
 */

package fgg

import (
	"fmt"
	"reflect"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/rhu1/fgg/base"
	"github.com/rhu1/fgg/base/testutils"
	"github.com/rhu1/fgg/parser/fgg"
	"github.com/rhu1/fgg/parser/util"
)

var _ = fmt.Errorf
var _ = reflect.Append

// Convert ANTLR generated CST to an FGNode AST
type FGGAdaptor struct {
	*parser.BaseFGGListener
	stack []FGGNode // Because Listener methods don't return...
}

var _ base.Adaptor = &FGGAdaptor{}

func (a *FGGAdaptor) push(n FGGNode) {
	a.stack = append(a.stack, n)
}

func (a *FGGAdaptor) pop() FGGNode {
	if len(a.stack) < 1 {
		panic(testutils.PARSER_PANIC_PREFIX + "Stack is empty")
	}
	res := a.stack[len(a.stack)-1]
	a.stack = a.stack[:len(a.stack)-1]
	return res
}

// strictParse means panic upon any parsing error -- o/w error recovery is attempted
func (a *FGGAdaptor) Parse(strictParse bool, input string) base.Program {
	is := antlr.NewInputStream(input)
	var lexer antlr.Lexer
	if strictParse { // https://stackoverflow.com/questions/51683104/how-to-catch-minor-errors
		lexer = FGGBailLexer{parser.NewFGGLexer(is)} // FIXME: not working -- e.g., incr{1}, bad token
	} else {
		lexer = parser.NewFGGLexer(is)
	}
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewFGGParser(stream)
	if strictParse {
		p.RemoveErrorListeners()
		p.SetErrorHandler(&util.StrictErrorStrategy{})
	}
	antlr.ParseTreeWalkerDefault.Walk(a, p.Program())
	return a.pop().(FGGProgram)
}

/* #Typeparam ("typ"), #TypeName ("typ"), "typeFormals", "typeFDecls", "typeFDecl" */

func (a *FGGAdaptor) ExitTypeParam(ctx *parser.TypeParamContext) {
	b := TParam(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText())
	a.push(b)
}

func (a *FGGAdaptor) ExitTypeName(ctx *parser.TypeNameContext) {
	t := Name(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText())
	us := []Type{}
	if ctx.GetChildCount() > 3 { // typs "helper" Context, cf. exprs
		nus := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., u1 ',' u2 ',' u3
		us = make([]Type, nus)
		for i := nus - 1; i >= 0; i-- {
			us[i] = a.pop().(Type) // Adding backwards
		}
	}
	a.push(TNamed{t, us})
}

func (a *FGGAdaptor) ExitTypeFormals(ctx *parser.TypeFormalsContext) {
	tfs := []TFormal{}
	if ctx.GetChildCount() > 3 {
		ntfs := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., tf ',' tf ',' tf
		tfs = make([]TFormal, ntfs)
		for i := ntfs - 1; i >= 0; i-- {
			tfs[i] = a.pop().(TFormal) // Adding backwards
		}
	}
	a.push(BigPsi{tfs})
}

func (a *FGGAdaptor) ExitTypeFDecl(ctx *parser.TypeFDeclContext) {
	u := a.pop().(Type)                                              // CHECKME: TName? (\tau_I)
	b := TParam(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText()) // Not pop().(TParam) -- BNF asks for NAME
	a.push(TFormal{b, u})
}

/* "program" */

// Duplicated from FG (generics would be nice!)
func (a *FGGAdaptor) ExitProgram(ctx *parser.ProgramContext) {
	body := a.pop().(FGGExpr)
	ds := []Decl{}
	offset := 0 // TODO: refactor
	printf := false
	c3 := ctx.GetChild(3)                                     // Check if this child is "import"
	foo := ctx.GetChild(ctx.GetChildCount() - 4).GetPayload() // Check if this child is the "=" in "_ = ..."
	if c3_cast, ok := c3.GetPayload().(*antlr.CommonToken); ok &&
		c3_cast.GetText() == "import" {
		if pkg := ctx.GetChild(4).GetPayload().(*antlr.CommonToken).GetText(); pkg != "\"fmt\"" { // TODO: refactor
			panic(testutils.PARSER_PANIC_PREFIX + "The only allowed import is \"fmt\"; found: " + pkg)
		}
		offset = 3
		if cast, ok := foo.(*antlr.CommonToken); !ok || cast.GetText() != "=" {
			printf = true
		}
	} else if cast, ok := foo.(*antlr.CommonToken); !ok || cast.GetText() != "=" {
		panic(testutils.PARSER_PANIC_PREFIX + "Missing \"import fmt;\".")
	}
	bar := ctx.GetChild(offset + 3)                                   // Check if this child is "func", i.e., no decls
	if _, ok := bar.GetPayload().(*antlr.BaseParserRuleContext); ok { // If "func", then *antlr.CommonToken
		nds := ctx.GetChild(offset+3).GetChildCount() / 2 // (decl ';')+ -- i.e, includes trailing ';'
		ds = make([]Decl, nds)
		for i := nds - 1; i >= 0; i-- {
			ds[i] = a.pop().(Decl) // Adding backwards
		}
	}
	a.push(FGGProgram{ds, body, printf})
}

/* "typeDecl" */

// Children: 1=NAME, 2=typeFormals, 3=typeLit
func (a *FGGAdaptor) ExitTypeDecl(ctx *parser.TypeDeclContext) {
	t := Name(ctx.GetChild(1).(*antlr.TerminalNodeImpl).GetText())
	td := a.pop().(TypeDecl)
	psi := a.pop().(BigPsi)
	if s, ok := td.(STypeLit); ok { // N.B. s is a *copy* of td
		s.t_name = t
		s.Psi = psi
		a.push(s)
	} else if c, ok := td.(ITypeLit); ok {
		c.t_I = t
		c.Psi = psi
		a.push(c)
	} else {
		panic(testutils.PARSER_PANIC_PREFIX + "Unknown type decl: " + reflect.TypeOf(td).String())
	}
}

/* #StructTypeLit ("typeLit"), "fieldDecls", "fieldDecl" */

// Children: 2=fieldDecls
func (a *FGGAdaptor) ExitStructTypeLit(ctx *parser.StructTypeLitContext) {
	fds := []FieldDecl{}
	if ctx.GetChildCount() > 3 {
		nfds := (ctx.GetChild(2).GetChildCount() + 1) / 2 // fieldDecl (';' fieldDecl)*
		fds = make([]FieldDecl, nfds)
		for i := nfds - 1; i >= 0; i-- {
			fd := a.pop().(FieldDecl)
			fds[i] = fd // Adding backwards
		}
	}
	a.push(STypeLit{"^", BigPsi{}, fds}) // "^" and TFormals{} to be overwritten in ExitTypeDecl
}

func (a *FGGAdaptor) ExitFieldDecl(ctx *parser.FieldDeclContext) {
	f := Name(ctx.GetField().GetText())
	//typ := Type(ctx.GetChild(1).GetText())
	u := a.pop().(Type)
	a.push(FieldDecl{f, u})
}

/* "methDecl", "paramDecl" */

func (a *FGGAdaptor) ExitMethDecl(ctx *parser.MethDeclContext) {
	// Reverse order
	e := a.pop().(FGGExpr)
	g := a.pop().(Sig)
	psi := a.pop().(BigPsi)
	t := Name(ctx.GetTypn().GetText())
	recv := Name(ctx.GetRecv().GetText())
	a.push(MethDecl{recv, t, psi, g.meth, g.Psi, g.pDecls, g.u_ret, e})
}

// Cf. ExitFieldDecl
func (a *FGGAdaptor) ExitParamDecl(ctx *parser.ParamDeclContext) {
	x := ctx.GetVari().GetText()
	u := a.pop().(Type)
	a.push(ParamDecl{x, u})
}

/* #InterfaceTypeLit ("typeLit"), "specs", #SigSpec ("spec"), #InterfaceSpec ("spec"), "sig" */

// Cf. ExitStructTypeLit
func (a *FGGAdaptor) ExitInterfaceTypeLit(ctx *parser.InterfaceTypeLitContext) {
	ss := []Spec{}
	if ctx.GetChildCount() > 3 {
		nss := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., s ';' s ';' s
		ss = make([]Spec, nss)
		for i := nss - 1; i >= 0; i-- {
			s := a.pop().(Spec)
			ss[i] = s // Adding backwards
		}
	}
	a.push(ITypeLit{"^", BigPsi{}, ss}) // "^" and TFormals{} to be overwritten in ExitTypeDecl
}

func (a *FGGAdaptor) ExitSigSpec(ctx *parser.SigSpecContext) {
	// No action -- Sig is at a.stack[len(a.stack)-1]
}

func (a *FGGAdaptor) ExitInterfaceSpec(ctx *parser.InterfaceSpecContext) {
	popped := a.pop()
	cast, ok := popped.(TNamed)
	if !ok {
		panic(testutils.PARSER_PANIC_PREFIX + "Expected TNamed, not: " + reflect.TypeOf(popped).String() +
			"\n\t" + popped.String())
	}
	a.push(cast) // Check TName (should specifically be a \tau_I) -- CHECKME: enforce in BNF?
}

func (a *FGGAdaptor) ExitSig(ctx *parser.SigContext) {
	m := ctx.GetMeth().GetText()
	// Reverse order
	t := a.pop().(Type)
	pds := []ParamDecl{}
	if ctx.GetChildCount() > 5 {
		npds := (ctx.GetChild(3).GetChildCount() + 1) / 2 // e.g., pd ',' pd ',' pd
		pds = make([]ParamDecl, npds)
		for i := npds - 1; i >= 0; i-- {
			pds[i] = a.pop().(ParamDecl) // Adding backwards
		}
	}
	psi := a.pop().(BigPsi)
	a.push(Sig{m, psi, pds, t})
}

/* "expr": #Variable, #StructLit, #Select, #Call, #Assert, #Sprintf */

// Same as FG
func (a *FGGAdaptor) ExitVariable(ctx *parser.VariableContext) {
	id := Name(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText())
	a.push(Variable{id})
}

// Children: 0=typ (*antlr.TerminalNodeImpl), 1='{', 2=exprs (*parser.ExprsContext), 3='}'
func (a *FGGAdaptor) ExitStructLit(ctx *parser.StructLitContext) {
	es := []FGGExpr{}
	if ctx.GetChildCount() > 3 {
		nes := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., 'x' ',' 'y' ',' 'z'
		es = make([]FGGExpr, nes)
		for i := nes - 1; i >= 0; i-- {
			es[i] = a.pop().(FGGExpr) // Adding backwards
		}
	}
	// If targs omitted, following will fail attempting to cast the non-param name parsed as a TParam
	tmp := a.pop()
	cast, ok := tmp.(TNamed)
	if !ok { // N.B. \tau_S, means "of the form t_S(~\tau)" (so a TName) -- i.e., not \alpha
		panic(testutils.PARSER_PANIC_PREFIX + "Expected named type, not: " +
			reflect.TypeOf(tmp).String() + "\n\t" + tmp.String())
	}
	a.push(StructLit{cast, es})
}

// Same as Fg
func (a *FGGAdaptor) ExitSelect(ctx *parser.SelectContext) {
	e := a.pop().(FGGExpr)
	f := Name(ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText())
	a.push(Select{e, f})
}

func (a *FGGAdaptor) ExitCall(ctx *parser.CallContext) {
	argCs := ctx.GetArgs()
	args := []FGGExpr{}
	if argCs != nil {
		nargs := (argCs.GetChildCount() + 1) / 2 // e.g., e ',' e ',' e
		args = make([]FGGExpr, nargs)
		for i := nargs - 1; i >= 0; i-- {
			args[i] = a.pop().(FGGExpr) // Adding backwards
		}
	}
	targCs := ctx.GetTargs()
	targs := []Type{}
	if targCs != nil {
		ntargs := (targCs.GetChildCount() + 1) / 2 // e.g., t ',' t ',' t
		targs = make([]Type, ntargs)
		for i := ntargs - 1; i >= 0; i-- {
			targs[i] = a.pop().(Type) // Adding backwards
		}
	}
	m := Name(ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText())
	e := a.pop().(FGGExpr)
	a.push(Call{e, m, targs, args})
}

func (a *FGGAdaptor) ExitAssert(ctx *parser.AssertContext) {
	u := a.pop().(Type)
	e := a.pop().(FGGExpr)
	a.push(Assert{e, u})
}

// TODO: check for import "fmt"
func (a *FGGAdaptor) ExitSprintf(ctx *parser.SprintfContext) {
	var format string = ctx.GetChild(4).(*antlr.TerminalNodeImpl).GetText()
	nargs := (ctx.GetChildCount() - 6) / 2 // Because of the comma
	args := make([]FGGExpr, nargs)
	for i := nargs - 1; i >= 0; i-- {
		tmp := a.pop()
		args[i] = tmp.(FGGExpr)
	}
	a.push(Sprintf{format, args})
}
