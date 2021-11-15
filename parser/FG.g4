//$ antlr4 -Dlanguage=Go -o parser/fg parser/FG.g4


// FG.g4
grammar FG;


/* Keywords */

FUNC      : 'func' ;
INTERFACE : 'interface' ;
MAIN      : 'main' ;
PACKAGE   : 'package' ;
RETURN    : 'return' ;
STRUCT    : 'struct' ;
TYPE      : 'type' ;

IMPORT    : 'import' ;
FMT       : 'fmt' ;
PRINTF    : 'Printf' ;
SPRINTF   : 'Sprintf' ;

// base/primitive types
TRUE      : 'true' ;
FALSE     : 'false' ;

BOOL      : 'bool' ;
INT32     : 'int32' ;
INT64     : 'int64' ;
FLOAT32   : 'float32' ;
FLOAT64   : 'float64' ;
STRING    : 'string' ;

// arithmetic ops
PLUS      : '+' ;
MINUS     : '-' ;
// logical ops
AND       : '&&' ;
OR        : '||' ;
// relational ops
GT        : '>' ;
LT        : '<' ;
// ...

/* Tokens */

fragment LETTER : ('a' .. 'z') | ('A' .. 'Z') | '\u03b1' | '\u03b2' ;
fragment DIGIT  : ('0' .. '9') ;
//fragment HACK   : 'ᐸ' | 'ᐳ' ;  // Doesn't seem to work?
fragment MONOM_HACK   : '\u1438' | '\u1433' | '\u1428' ;  // Hack for monom output
NAME            : (LETTER | '_' | MONOM_HACK) (LETTER | '_' | DIGIT | MONOM_HACK)* ;
WHITESPACE      : [ \r\n\t]+ -> skip ;
COMMENT         : '/*' .*? '*/' -> channel(HIDDEN) ;
LINE_COMMENT    : '//' ~[\r\n]* -> channel(HIDDEN) ;
STRING_LIT      : '"' (LETTER | DIGIT | ' ' | '.' | ',' | '_' | '%' | '#' | '(' | ')' | '+' | '-')* '"' ;

fragment DIGITS : DIGIT+ ;
fragment EXPON  : [eE] [+-]? DIGITS ;
INT_LIT         : DIGITS ;
FLOAT_LIT       : DIGITS ('.' DIGIT* EXPON? | EXPON)
                | '.' DIGITS EXPON?
                ;


/* Rules */

// Conventions:
// "tag=" to distinguish repeat productions within a rule: comes out in
// field/getter names.
// "#tag" for cases within a rule: comes out as Context names (i.e., types).
// "plurals", e.g., decls, used for sequences: comes out as "helper" Contexts,
// nodes that group up actual children underneath -- makes "adapting" easier.

typ        : name=NAME                              # TNamed
           | name=primName                          # TPrimitive
           | typeLit                                # TypeLit_
           ;
primName   : BOOL
           | INT32 | INT64
           | FLOAT32 | FLOAT64
           | STRING
           ;
typeLit    : STRUCT '{' fieldDecls? '}'             # StructTypeLit
           | INTERFACE '{' specs? '}'               # InterfaceTypeLit
           ;
program    : PACKAGE MAIN ';'
             (IMPORT STRING_LIT ';')?
             decls? FUNC MAIN '(' ')' '{'
             ('_' '=' expr | FMT '.' PRINTF '(' '"%#v"' ',' expr ')')
             '}' EOF ;
decls      : ((typeDecl | methDecl) ';')+ ;
typeDecl   : TYPE id=NAME typ ;
methDecl   : FUNC '(' paramDecl ')' sig '{' RETURN expr '}' ;
fieldDecls : fieldDecl (';' fieldDecl)* ;
fieldDecl  : field=NAME typ ;
specs      : spec (';' spec)* ;
spec       : (sig | typ) ;
sig        : meth=NAME '(' params? ')' typ ;
params     : paramDecl (',' paramDecl)* ;
paramDecl  : vari=NAME typ ;
expr       : NAME                                   # Variable
           | typ '{' exprs? '}'                     # StructLit
           | expr '.' NAME                          # Select
           | recv=expr '.' NAME '(' args=exprs? ')' # Call
           | expr '.' '(' typ ')'                   # Assert
           | FMT '.' SPRINTF '(' (STRING_LIT | '"%#v"') (',' | expr)* ')'  # Sprintf
           | expr op=(PLUS | MINUS) expr            # BinaryOp
           | expr op=(GT | LT) expr                 # BinaryOp
           | expr op=AND expr                       # BinaryOp
           | expr op=OR expr                        # BinaryOp
           | '(' expr ')'                           # Paren
           | primLit                                # PrimaryLit
           ;
exprs      : expr (',' expr)* ;

primLit    : lit=(TRUE|FALSE)                       # BoolLit
           | lit=INT_LIT                            # IntLit
           | lit=FLOAT_LIT                          # FloatLit
           | lit=STRING_LIT                         # StringLit
           ;

