
grammar FGG;

/* Keywords */

FUNC: 'func';
INTERFACE: 'interface';
MAIN: 'main';
PACKAGE: 'package';
RETURN: 'return';
STRUCT: 'struct';
TYPE: 'type';

IMPORT: 'import';
FMT: 'fmt';
PRINTF: 'Printf';
SPRINTF: 'Sprintf';

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

fragment LETTER : ('a' .. 'z') | ('A' .. 'Z')
	            | 'α' // For FGR deserialization
	            | 'β'
	            ;
fragment DIGIT  : ('0' .. '9') ;
NAME            : (LETTER | '_') (LETTER | '_' | DIGIT)* ;
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

// Conventions: "tag=" to distinguish repeat productions within a rule: comes out in field/getter
// names. "#tag" for cases within a rule: comes out as Context names (i.e., types). "plurals", e.g.,
// decls, used for sequences: comes out as "helper" Contexts, nodes that group up actual children
// underneath -- makes "adapting" easier.

typ        : name=NAME                              # TypeParam
           | name=NAME '(' typs? ')'                # TypeName
           | name=primName                          # TPrimitive
           | typeLit                                # TypeLit_
           ;
typs       : typ (',' typ)* ;
primName   : BOOL
           | INT32   | INT64
           | FLOAT32 | FLOAT64
           | STRING
           ;
typeLit    : STRUCT '{' fieldDecls? '}'	            # StructTypeLit
//	       | INTERFACE '{' (typeList ';')? specs? '}'	    # InterfaceTypeLit
	       | INTERFACE '{' typeList? specs? '}'	    # InterfaceTypeLit
	       ;
typeFormals: '(' TYPE typeFDecls? ')' ; // Refactored "(...)" into here
typeFDecls : typeFDecl (',' typeFDecl)* ;
typeFDecl  : NAME typ ;
program    : PACKAGE MAIN ';'
             (IMPORT STRING_LIT ';')?
             decls? FUNC MAIN '(' ')' '{'
                ( '_' '=' expr | FMT '.' PRINTF '(' '"%#v"' ',' expr ')' )
             '}' EOF ;
decls      : ((typeDecl | methDecl) ';')+ ;
typeDecl   : TYPE id=NAME typeFormals typ ;
methDecl   : FUNC '(' recv = NAME typn = NAME typeFormals ')' sig '{' RETURN expr '}' ;
fieldDecls : fieldDecl (';' fieldDecl)*;
fieldDecl  : field = NAME typ;
typeList   : TYPE typs ;
specs      : spec (';' spec)*;
spec       : (sig | typ) ;
sig        : meth = NAME typeFormals '(' params? ')' typ;
params     : paramDecl (',' paramDecl)*;
paramDecl  : vari = NAME typ;
expr       :
	NAME																# Variable
	| typ '{' exprs? '}' /* typ is #TypeName, \tau_S */			        # StructLit
	| expr '.' NAME														# Select
	| recv = expr '.' NAME '(' targs = typs? ')' '(' args = exprs? ')'	# Call
	| expr '.' '(' typ ')'												# Assert
	| FMT '.' SPRINTF '(' (STRING_LIT | '"%#v"') (',' | expr)* ')'		# Sprintf
	| expr op=(PLUS | MINUS) expr                                       # BinaryOp
	| expr op=(GT | LT) expr                                            # BinaryOp
	| expr op=AND expr                                                  # BinaryOp
	| expr op=OR expr                                                   # BinaryOp
	| '(' expr ')'                                                      # Paren
	| primLit                                                           # PrimaryLit
	;
exprs      : expr (',' expr)*;
primLit    : lit=(TRUE|FALSE)                       # BoolLit
           | lit=INT_LIT                            # IntLit
           | lit=FLOAT_LIT                          # FloatLit
           | lit=STRING_LIT                         # StringLit
           ;
