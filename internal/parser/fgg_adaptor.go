/*
 * TODO: fix many magic numbers and other sloppy hacks
 */

package parser

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/rhu1/fgg/internal/base"
	"github.com/rhu1/fgg/internal/base/testutils"
	"github.com/rhu1/fgg/internal/fgg"
	"github.com/rhu1/fgg/internal/parser/util"
	"github.com/rhu1/fgg/parser/fgg/parser"
)

var _ = fmt.Errorf
var _ = reflect.Append

// Convert ANTLR generated CST to an FGNode AST
type FGGAdaptor struct {
	*parser.BaseFGGListener
	stack []fgg.FGGNode // Because Listener methods don't return...
}

var _ base.Adaptor = &FGGAdaptor{}

func (a *FGGAdaptor) push(n fgg.FGGNode) {
	a.stack = append(a.stack, n)
}

func (a *FGGAdaptor) pop() fgg.FGGNode {
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
	return a.pop().(fgg.FGGProgram)
}

/* "typ": #TypeParam, #TypeName, #TPrimitive, #TypeLit_ */

func (a *FGGAdaptor) ExitTypeParam(ctx *parser.TypeParamContext) {
	b := fgg.NewTParam(ctx.GetName().GetText())
	a.push(b)
}

func (a *FGGAdaptor) ExitTypeName(ctx *parser.TypeNameContext) {
	//t := fgg.Name(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText())
	// TODO check if changing to .GetName doesn't alter the ChildCount below
	t := fgg.Name(ctx.GetName().GetText())
	us := []fgg.Type{}
	if ctx.GetChildCount() > 3 { // typs "helper" Context, cf. exprs
		nus := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., u1 ',' u2 ',' u3
		us = make([]fgg.Type, nus)
		for i := nus - 1; i >= 0; i-- {
			us[i] = a.pop().(fgg.Type) // Adding backwards
		}
	}

	a.push(fgg.NewTNamed(t, us))
}

func (a *FGGAdaptor) ExitTPrimitive(ctx *parser.TPrimitiveContext) {
	tag := fgg.TagFromName(ctx.GetName().GetText())
	a.push(fgg.NewTPrimitive(tag))
}

func (a *FGGAdaptor) ExitTypeLit_(ctx *parser.TypeLit_Context) {
	// do nothing -- the struct/interface literal is already at top of a.stack
	// cf. ExitStructTypeLit
}

/* #StructTypeLit ("typeLit"), #InterfaceTypeLit ("typeLit") */

// Children: 2=fieldDecls
func (a *FGGAdaptor) ExitStructTypeLit(ctx *parser.StructTypeLitContext) {
	fds := []fgg.FieldDecl{}
	if ctx.FieldDecls() != nil {
		nfds := (ctx.FieldDecls().GetChildCount() + 1) / 2 // fieldDecl (';' fieldDecl)*
		fds = make([]fgg.FieldDecl, nfds)
		for i := nfds - 1; i >= 0; i-- {
			fd := a.pop().(fgg.FieldDecl)
			fds[i] = fd // Adding backwards
		}
	}
	a.push(fgg.NewSTypeLit(fds))
}

// Cf. ExitStructTypeLit
func (a *FGGAdaptor) ExitInterfaceTypeLit(ctx *parser.InterfaceTypeLitContext) {
	specs := []fgg.Spec{}
	tlist := []fgg.Type{}
	// first add specs and then type list (inverse order of syntax)
	if ctx.Specs() != nil {
		nss := (ctx.Specs().GetChildCount() + 1) / 2 // e.g., s ';' s ';' s
		specs = make([]fgg.Spec, nss)
		for i := nss - 1; i >= 0; i-- {
			s := a.pop().(fgg.Spec)
			specs[i] = s // Adding backwards
		}
	}
	if ctx.TypeList() != nil {
		// ctx.TypeList().Children: 0="type", 1=typs
		typs := ctx.TypeList().GetChild(1)
		nus := (typs.GetChildCount() + 1) / 2 // e.g., u1 ',' u2 ',' u3
		tlist = make([]fgg.Type, nus)
		for i := nus - 1; i >= 0; i-- {
			tlist[i] = a.pop().(fgg.Type) // Adding backwards
		}
	}
	a.push(fgg.NewITypeLit(specs, tlist))
}

/* "typeFormals", "typeFDecls", "typeFDecl" */

func (a *FGGAdaptor) ExitTypeFormals(ctx *parser.TypeFormalsContext) {
	tfs := []fgg.TFormal{}
	if ctx.GetChildCount() > 3 {
		ntfs := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., tf ',' tf ',' tf
		tfs = make([]fgg.TFormal, ntfs)
		for i := ntfs - 1; i >= 0; i-- {
			tfs[i] = a.pop().(fgg.TFormal) // Adding backwards
		}
	}
	a.push(fgg.NewBigPsi(tfs))
}

func (a *FGGAdaptor) ExitTypeFDecl(ctx *parser.TypeFDeclContext) {
	u := a.pop().(fgg.Type)                                              // CHECKME: TName? (\tau_I)
	b := fgg.NewTParam(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText()) // Not pop().(TParam) -- BNF asks for NAME
	a.push(fgg.NewTFormal(b, u))
}

/* "program" */

// Duplicated from FG (generics would be nice!)
func (a *FGGAdaptor) ExitProgram(ctx *parser.ProgramContext) {
	body := a.pop().(fgg.FGGExpr)
	ds := []fgg.Decl{}
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
		ds = make([]fgg.Decl, nds)
		for i := nds - 1; i >= 0; i-- {
			ds[i] = a.pop().(fgg.Decl) // Adding backwards
		}
	}
	a.push(fgg.NewProgram(ds, body, printf))
}

/* "typeDecl" */

// Children: 1=NAME, 2=typeFormals, 3=typeLit
func (a *FGGAdaptor) ExitTypeDecl(ctx *parser.TypeDeclContext) {
	name := fgg.Name(ctx.GetId().GetText())
	src := a.pop().(fgg.Type)
	psi := a.pop().(fgg.BigPsi)
	a.push(fgg.NewTypeDecl(name, psi, src))
}

/* "fieldDecls", "fieldDecl" */

func (a *FGGAdaptor) ExitFieldDecl(ctx *parser.FieldDeclContext) {
	f := fgg.Name(ctx.GetField().GetText())
	//typ := Type(ctx.GetChild(1).GetText())
	u := a.pop().(fgg.Type)
	a.push(fgg.NewFieldDecl(f, u))
}

/* "methDecl", "paramDecl" */

func (a *FGGAdaptor) ExitMethDecl(ctx *parser.MethDeclContext) {
	// Reverse order
	e := a.pop().(fgg.FGGExpr)
	g := a.pop().(fgg.Sig)
	psi := a.pop().(fgg.BigPsi)
	t := fgg.Name(ctx.GetTypn().GetText())
	recv := fgg.Name(ctx.GetRecv().GetText())
	a.push(fgg.NewMDecl(recv, t, psi, g.GetMethod(), g.GetPsi(), g.GetParamDecls(), g.GetReturn(), e))
}

// Cf. ExitFieldDecl
func (a *FGGAdaptor) ExitParamDecl(ctx *parser.ParamDeclContext) {
	x := ctx.GetVari().GetText()
	u := a.pop().(fgg.Type)
	a.push(fgg.NewParamDecl(x, u))
}

/* "sig" */

func (a *FGGAdaptor) ExitSig(ctx *parser.SigContext) {
	m := ctx.GetMeth().GetText()
	// Reverse order
	t := a.pop().(fgg.Type)
	pds := []fgg.ParamDecl{}
	if ctx.GetChildCount() > 5 {
		npds := (ctx.GetChild(3).GetChildCount() + 1) / 2 // e.g., pd ',' pd ',' pd
		pds = make([]fgg.ParamDecl, npds)
		for i := npds - 1; i >= 0; i-- {
			pds[i] = a.pop().(fgg.ParamDecl) // Adding backwards
		}
	}
	psi := a.pop().(fgg.BigPsi)
	a.push(fgg.NewSig(m, psi, pds, t))
}

/* "expr": #Variable, #StructLit, #Select, #Call, #Assert, #Sprintf */

// Same as FG
func (a *FGGAdaptor) ExitVariable(ctx *parser.VariableContext) {
	id := fgg.Name(ctx.GetChild(0).(*antlr.TerminalNodeImpl).GetText())
	a.push(fgg.NewVariable(id))
}

// Children: 0=typ (*antlr.TerminalNodeImpl), 1='{', 2=exprs (*parser.ExprsContext), 3='}'
func (a *FGGAdaptor) ExitStructLit(ctx *parser.StructLitContext) {
	es := []fgg.FGGExpr{}
	if ctx.GetChildCount() > 3 {
		nes := (ctx.GetChild(2).GetChildCount() + 1) / 2 // e.g., 'x' ',' 'y' ',' 'z'
		es = make([]fgg.FGGExpr, nes)
		for i := nes - 1; i >= 0; i-- {
			es[i] = a.pop().(fgg.FGGExpr) // Adding backwards
		}
	}
	t := a.pop().(fgg.Type)
	a.push(fgg.NewStructLit(t, es))
}

// Same as Fg
func (a *FGGAdaptor) ExitSelect(ctx *parser.SelectContext) {
	e := a.pop().(fgg.FGGExpr)
	f := fgg.Name(ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText())
	a.push(fgg.NewSelect(e, f))
}

func (a *FGGAdaptor) ExitCall(ctx *parser.CallContext) {
	argCs := ctx.GetArgs()
	args := []fgg.FGGExpr{}
	if argCs != nil {
		nargs := (argCs.GetChildCount() + 1) / 2 // e.g., e ',' e ',' e
		args = make([]fgg.FGGExpr, nargs)
		for i := nargs - 1; i >= 0; i-- {
			args[i] = a.pop().(fgg.FGGExpr) // Adding backwards
		}
	}
	targCs := ctx.GetTargs()
	targs := []fgg.Type{}
	if targCs != nil {
		ntargs := (targCs.GetChildCount() + 1) / 2 // e.g., t ',' t ',' t
		targs = make([]fgg.Type, ntargs)
		for i := ntargs - 1; i >= 0; i-- {
			targs[i] = a.pop().(fgg.Type) // Adding backwards
		}
	}
	m := fgg.Name(ctx.GetChild(2).(*antlr.TerminalNodeImpl).GetText())
	e := a.pop().(fgg.FGGExpr)
	a.push(fgg.NewCall(e, m, targs, args))
}

func (a *FGGAdaptor) ExitAssert(ctx *parser.AssertContext) {
	u := a.pop().(fgg.Type)
	e := a.pop().(fgg.FGGExpr)
	a.push(fgg.NewAssert(e, u))
}

// TODO: check for import "fmt"
func (a *FGGAdaptor) ExitSprintf(ctx *parser.SprintfContext) {
	var format string = ctx.GetChild(4).(*antlr.TerminalNodeImpl).GetText()
	nargs := (ctx.GetChildCount() - 6) / 2 // Because of the comma
	args := make([]fgg.FGGExpr, nargs)
	for i := nargs - 1; i >= 0; i-- {
		tmp := a.pop()
		args[i] = tmp.(fgg.FGGExpr)
	}
	a.push(fgg.NewSprintf(format, args))
}

/* For "strict" parsing, *lexer* errors */

type FGGBailLexer struct {
	*parser.FGGLexer
}

// FIXME: not working -- e.g., incr{1}, bad token
// Want to "override" *BaseLexer.Recover -- XXX that's not how Go works (because BaseLexer is a struct, not interface)
func (b *FGGBailLexer) Recover(re antlr.RecognitionException) {
	message := "lex error after token " + re.GetOffendingToken().GetText() +
		" at position " + strconv.Itoa(re.GetOffendingToken().GetStart())
	panic(message)
}

/* Primitive binary operations: #BinaryOp */

func (a *FGGAdaptor) ExitBinaryOp(ctx *parser.BinaryOpContext) {
	r, l := a.pop().(fgg.FGGExpr), a.pop().(fgg.FGGExpr)
	op := fgg.Operator(ctx.GetOp().GetText())
	a.push(fgg.NewBinaryOp(l, r, op))
}

/* "primLit": #BoolLit, #IntLit, #FloatLit */

func (a *FGGAdaptor) ExitBoolLit(ctx *parser.BoolLitContext) {
	lit := ctx.GetLit().GetText()
	a.push(fgg.NewBoolLit(lit))
}

func (a *FGGAdaptor) ExitIntLit(ctx *parser.IntLitContext) {
	lit := ctx.GetLit().GetText()
	a.push(fgg.NewIntLit(lit))
}

func (a *FGGAdaptor) ExitFloatLit(ctx *parser.FloatLitContext) {
	lit := ctx.GetLit().GetText()
	a.push(fgg.NewFloatLit(lit))
}

func (a *FGGAdaptor) ExitStringLit(ctx *parser.StringLitContext) {
	lit := ctx.GetLit().GetText()
	a.push(fgg.NewStringLit(lit))
}