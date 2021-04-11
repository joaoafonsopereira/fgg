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
INT8      : 'int8' ;
INT32     : 'int32' ;
INT64     : 'int64' ;

/* Tokens */

fragment LETTER : ('a' .. 'z') | ('A' .. 'Z') | '\u03b1' | '\u03b2' ;
fragment DIGIT  : ('0' .. '9') ;
//fragment HACK   : 'ᐸ' | 'ᐳ' ;  // Doesn't seem to work?
fragment MONOM_HACK   : '\u1438' | '\u1433' | '\u1428' ;  // Hack for monom output
NAME            : (LETTER | '_' | MONOM_HACK) (LETTER | '_' | DIGIT | MONOM_HACK)* ;
WHITESPACE      : [ \r\n\t]+ -> skip ;
COMMENT         : '/*' .*? '*/' -> channel(HIDDEN) ;
LINE_COMMENT    : '//' ~[\r\n]* -> channel(HIDDEN) ;
STRING          : '"' (LETTER | DIGIT | ' ' | '.' | ',' | '_' | '%' | '#' | '(' | ')' | '+' | '-')* '"' ;

INT_LIT         : DIGIT+ ;


/* Rules */

// Conventions:
// "tag=" to distinguish repeat productions within a rule: comes out in
// field/getter names.
// "#tag" for cases within a rule: comes out as Context names (i.e., types).
// "plurals", e.g., decls, used for sequences: comes out as "helper" Contexts,
// nodes that group up actual children underneath -- makes "adapting" easier.

primitiveType       : INT8 | INT32 | INT64 ;
//PRIMITIVE_TYPE       : 'int8' | 'int32' | 'int64' ;

//typeName   : name=PRIMITIVE_TYPE                         # TPrimitive
typeName   : name=primitiveType                          # TPrimitive
           | name=NAME                                   # TNamed
           ;
program    : PACKAGE MAIN ';'
             (IMPORT STRING ';')?
             decls? FUNC MAIN '(' ')' '{'
             ('_' '=' expr | FMT '.' PRINTF '(' '"%#v"' ',' expr ')')
             '}' EOF ;
decls      : ((typeDecl | methDecl) ';')+ ;
typeDecl   : TYPE NAME typeLit ;  // TODO: tag id=NAME, better for adapting (vs., index constants)
methDecl   : FUNC '(' paramDecl ')' sig '{' RETURN expr '}' ;
typeLit    : STRUCT '{' fieldDecls? '}'             # StructTypeLit
           | INTERFACE '{' specs? '}'               # InterfaceTypeLit ;
fieldDecls : fieldDecl (';' fieldDecl)* ;
//fieldDecl  : field=NAME typ=NAME ;
fieldDecl  : field=NAME typeName ;
specs      : spec (';' spec)* ;
spec       : sig                                    # SigSpec
           | NAME                                   # InterfaceSpec
           ;
//sig        : meth=NAME '(' params? ')' ret=NAME ;
sig        : meth=NAME '(' params? ')' typeName ;
params     : paramDecl (',' paramDecl)* ;
//paramDecl  : vari=NAME typ=NAME ;
paramDecl  : vari=NAME typeName ;
expr       : NAME                                   # Variable
           | NAME '{' exprs? '}'                    # StructLit
           | expr '.' NAME                          # Select
           | recv=expr '.' NAME '(' args=exprs? ')' # Call
           | expr '.' '(' typeName ')'              # Assert
           | FMT '.' SPRINTF '(' (STRING | '"%#v"') (',' | expr)* ')'  # Sprintf
           | lit=INT_LIT                            # IntLit
           ;
exprs      : expr (',' expr)* ;

