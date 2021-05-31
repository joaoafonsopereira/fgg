package fgg

import (
	"fmt"

	"github.com/rhu1/fgg/internal/base"
	"github.com/rhu1/fgg/internal/fg"
)

type fg2fgg struct {
	fgProg  fg.FGProgram
	fggProg FGGProgram
}

// FromFG converts a FG program prog into a FGG program
// with empty type parameters
func FromFG(prog fg.FGProgram) (FGGProgram, error) {
	c := &fg2fgg{fgProg: prog}
	if err := c.convert(); err != nil {
		return FGGProgram{}, err
	}
	return c.fggProg, nil
}

func (c *fg2fgg) convert() error {
	// convert toplevel declarations
	for _, decl := range c.fgProg.GetDecls() {
		switch decl := decl.(type) {
		case fg.TypeDecl:
			tDecl, err := c.convertTDecl(decl)
			if err != nil {
				return err
			}
			c.fggProg.decls = append(c.fggProg.decls, tDecl)

		case fg.MethDecl:
			mDecl, err := c.convertMDecl(decl)
			if err != nil {
				return err
			}
			c.fggProg.decls = append(c.fggProg.decls, mDecl)

		default:
			return fmt.Errorf("unknown declaration type: %T", decl)
		}
	}

	expr, err := c.convertExpr(c.fgProg.GetMain())
	if err != nil {
		return err
	}
	c.fggProg.e_main = expr

	return nil
}

func (c *fg2fgg) convertTDecl(td fg.TypeDecl) (TypeDecl, error) {
	emptyPsi := BigPsi{tFormals: nil} // 0 formal parameters
	src, err := c.convertType(td.GetSourceType())
	if err != nil {
		return TypeDecl{}, err
	}
	return TypeDecl{td.GetName(), emptyPsi, src}, nil
}

// convertType converts a fg type to a fgg (parameterised) type
func (c *fg2fgg) convertType(t fg.Type) (Type, error) {
	switch t := t.(type) {
	case fg.TNamed:
		return TNamed{t.String(), nil}, nil

	case fg.TPrimitive:
		return TPrimitive{Tag(t.Tag()), t.Undefined()}, nil

	case fg.STypeLit:
		return c.convertSTypeLit(t)

	case fg.ITypeLit:
		return c.convertITypeLit(t)

	default:
		return nil, fmt.Errorf("unknown fg.Type type: %T", t)
	}
}

func (c *fg2fgg) convertSTypeLit(s fg.STypeLit) (STypeLit, error) {
	var fieldDecls []FieldDecl
	for _, f := range s.GetFieldDecls() {
		fd, err := c.convertFieldDecl(f)
		if err != nil {
			return STypeLit{}, err
		}
		fieldDecls = append(fieldDecls, fd)
	}
	return STypeLit{fieldDecls}, nil
}

func (c *fg2fgg) convertITypeLit(i fg.ITypeLit) (ITypeLit, error) {
	var specs []Spec
	for _, s := range i.GetSpecs() {
		switch spec := s.(type) {
		case fg.Sig:
			sig := spec
			var paramDecls []ParamDecl
			for _, p := range sig.GetParamDecls() {
				pd, err := c.convertParamDecl(p)
				if err != nil {
					return ITypeLit{}, nil
				}
				paramDecls = append(paramDecls, pd)
			}
			retTypeName, _ := c.convertType(sig.GetReturn())

			specs = append(specs, Sig{
				meth:   Name(sig.GetMethod()),
				Psi:    BigPsi{tFormals: nil},
				pDecls: paramDecls,
				u_ret:  retTypeName,
			})
		case fg.TNamed:
			emb, _ := c.convertType(spec)
			specs = append(specs, emb.(TNamed))
		}
	}
	return ITypeLit{specs}, nil
}

func (c *fg2fgg) convertFieldDecl(fd fg.FieldDecl) (FieldDecl, error) {
	typeName, err := c.convertType(fd.GetType())
	if err != nil {
		return FieldDecl{}, err
	}
	return FieldDecl{field: fd.GetName(), u: typeName}, nil
}

func (c *fg2fgg) convertParamDecl(pd fg.ParamDecl) (ParamDecl, error) {
	typeName, err := c.convertType(pd.GetType())
	if err != nil {
		return ParamDecl{}, err
	}
	return ParamDecl{name: pd.GetName(), u: typeName}, nil
}

func (c *fg2fgg) convertMDecl(md fg.MethDecl) (MethDecl, error) {
	recvType, _ := c.convertType(md.GetReceiver().GetType())

	var paramDecls []ParamDecl
	for _, p := range md.GetParamDecls() {
		pd, err := c.convertParamDecl(p)
		if err != nil {
			return MethDecl{}, err
		}
		paramDecls = append(paramDecls, pd)
	}

	retType, _ := c.convertType(md.GetReturn())
	methImpl, err := c.convertExpr(md.GetBody())
	if err != nil {
		return MethDecl{}, err
	}

	return MethDecl{
		x_recv:   md.GetReceiver().GetName(),
		t_recv:   recvType.(TNamed).GetName(),
		Psi_recv: BigPsi{nil}, // empty parameter
		name:     Name(md.GetName()),
		Psi_meth: BigPsi{nil}, // empty parameter
		pDecls:   paramDecls,
		u_ret:    retType,
		e_body:   methImpl,
	}, nil
}

func (c *fg2fgg) convertExpr(expr base.Expr) (FGGExpr, error) {
	switch expr := expr.(type) {
	case fg.Variable:
		return Variable{name: expr.String()}, nil

	case fg.StructLit:
		sLitExpr, err := c.convertStructLit(expr)
		if err != nil {
			return nil, err
		}
		return sLitExpr, nil

	case fg.Call:
		callExpr, err := c.convertCall(expr)
		if err != nil {
			return nil, err
		}
		return callExpr, nil

	case fg.Select:
		selExpr, err := c.convertExpr(expr.GetExpr())
		if err != nil {
			return nil, err
		}
		return Select{e_S: selExpr, field: Name(expr.GetField())}, nil

	case fg.Assert:
		assertExpr, err := c.convertExpr(expr.GetExpr())
		if err != nil {
			return nil, err
		}
		assType, _ := c.convertType(expr.GetType())
		return Assert{e_I: assertExpr, u_cast: assType}, nil
	}

	return nil, fmt.Errorf("unknown expression type: %T", expr)
}

func (c *fg2fgg) convertStructLit(sLit fg.StructLit) (StructLit, error) {
	structType, _ := c.convertType(sLit.GetType())

	var es []FGGExpr
	for _, expr := range sLit.GetElems() {
		fieldExpr, err := c.convertExpr(expr)
		if err != nil {
			return StructLit{}, err
		}
		es = append(es, fieldExpr)
	}

	return StructLit{u_S: structType.(TNamed), elems: es}, nil
}

func (c *fg2fgg) convertCall(call fg.Call) (Call, error) {
	e, err := c.convertExpr(call.GetReceiver())
	if err != nil {
		return Call{}, err
	}

	var args []FGGExpr
	for _, arg := range call.GetArgs() {
		argExpr, err := c.convertExpr(arg)
		if err != nil {
			return Call{}, err
		}
		args = append(args, argExpr)
	}

	return Call{e_recv: e, meth: Name(call.GetMethod()), args: args}, nil
}
