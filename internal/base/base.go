package base

/* Name */

type Name = string // Type alias (cf. type def)

/* AST Nodes */

type AstNode interface {
	String() string
}

// Top-level decls -- not Field/ParamDecl
type Decl interface {
	AstNode
	GetName() Name
	Ok(ds []Decl)
}

type Program interface {
	AstNode
	GetDecls() []Decl
	GetMain() Expr
	Ok(allowStupid bool, mode TypingMode) (Type, Program) // Set false for source check
	Eval() (Program, string)  // Eval one step; string is the name of the (innermost) applied rule
}

type Expr interface {
	AstNode
	IsValue() bool
	CanEval(ds []Decl) bool      // More like, canReduce -- basically, bad assert returns false, cf. IsValue()
	ToGoString(ds []Decl) string // Basically, type T printed as main.T  // TODO (cf. %#v, Go-syntax value representation)
}

/* Types */

type Type interface {
	// Returns a "coerced" AST leaf if tested on a Literal expr
	//AssignableTo(ds []Decl, t Type) (bool, FGExpr)
	//AssignableTo(ds []Decl, t Type) (bool, Coercion)

	Equals(t Type) bool
	String() string
}

/* ANTLR (parsing) */

type Adaptor interface {
	Parse(strictParse bool, input string) Program
}

/* Typing modes */ // TODO decide where (pkg/file) does it make more sense to put this

type TypingMode int

const (
	CHECK TypingMode = iota
	INFER
)