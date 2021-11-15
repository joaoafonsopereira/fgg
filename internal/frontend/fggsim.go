package frontend

import (
	"github.com/rhu1/fgg/internal/base"
	"github.com/rhu1/fgg/internal/fg"
	"github.com/rhu1/fgg/internal/fgg"
	"github.com/rhu1/fgg/internal/fgr"
	"regexp"
)

// helpers - perform both casts (Type, Program) in a single line
// (useful because Ok returns 2 values)

func castFG(t base.Type, p base.Program) (fg.Type, fg.FGProgram) {
	return t.(fg.Type), p.(fg.FGProgram)
}

func castFGG(t base.Type, p base.Program) (fgg.Type, fgg.FGGProgram) {
	return t.(fgg.Type), p.(fgg.FGGProgram)
}

func castFGR(t base.Type, p base.Program) (fgr.Type, fgr.FGRProgram) {
	return t.(fgr.Type), p.(fgr.FGRProgram)
}

/* monom simulation check */

// TODO: refactor to cmd dir
func TestMonom(printf bool, verbose bool, src string, steps int) {
	intrp_fgg := NewFGGInterp(verbose, src, true)
	p_fgg := intrp_fgg.GetProgram().(fgg.FGGProgram)
	u_fgg, p_fgg := castFGG( p_fgg.Ok(false, base.CHECK) )
	VPrintln(verbose, "\nFGG expr: "+p_fgg.GetMain().String())

	if ok, msg := fgg.IsMonomOK(p_fgg); !ok {
		VPrintln(verbose, "\nAborting simulation: Cannot monomorphise (nomono detected):\n\t"+msg)
		return
	}

	// (Initial) left-vertical arrow
	//p_mono := fgg.Monomorph(p_fgg)
	ds_fgg := p_fgg.GetDecls()
	omega := fgg.GetOmega(ds_fgg, p_fgg.GetMain().(fgg.FGGExpr))
	p_mono := fgg.ApplyOmega(p_fgg, omega) // TODO: can just monom expr (ground main) directly
	VPrintln(verbose, "Monom expr: "+p_mono.GetMain().String())
	t_mono, p_mono := castFG( p_mono.Ok(false, base.CHECK) )
	ds_mono := p_mono.GetDecls()
	u_fg := fgg.ToMonomId(u_fgg)
	if !t_mono.Equals(u_fg) {
		panic("-test-monom failed: types do not match\n\tFGG type=" + u_fgg.String() +
			" -> " + u_fg.String() + "\n\tmono=" + t_mono.String())
	}

	done := steps > EVAL_TO_VAL
	var main_fgg base.Expr
	var main_mono base.Expr
	for i := 0; i < steps || !done; i++ {
		main_fgg = p_fgg.GetMain()
		main_mono = p_mono.GetMain()
		if main_fgg.IsValue() { // N.B. IsValue -- not CanEval (checked below)
			if !main_mono.IsValue() { // TODO: add to -test-oblit
				panic("FGG is value but monom is not:\n\tfgg = " + main_fgg.String() +
					"\n\tmonom=" + main_mono.String())
			}
			break // Both are values
		} else if main_mono.IsValue() {
			panic("Monom is value but FGG is not:\n\tfgg = " + main_fgg.String() +
				"\n\tmonom=" + main_mono.String())
		}
		// Both non-values, check for stuck (e.g., bad asserts -- though panic is technically not stuck)
		if main_fgg.CanEval(ds_fgg) {
			if !main_mono.CanEval(ds_mono) {
				panic("FGG is stuck but monom is not:\n\tfgg = " + main_fgg.String() +
					"\n\tmonom=" + main_mono.String())
			}
		} else {
			if main_mono.CanEval(ds_mono) {
				panic("Monom is stuck but FGG is not:\n\tfgg = " + main_fgg.String() +
					"\n\tmonom=" + main_mono.String())
			}
			if _, ok := main_fgg.(fgg.Assert); ok {
				if _, ok1 := main_mono.(fg.Assert); ok1 {
					break // Both stuck on bad assert
				}
			}
		}

		// Repeat: horizontal arrows and right-vertical arrow
		p_fgg, u_fgg, p_mono = testMonomStep(verbose, omega, p_fgg, u_fgg, p_mono)
	}
	VPrintln(verbose, "\nFinished:\n\tfgg="+p_fgg.GetMain().String()+
		"\n\tmono="+p_mono.GetMain().String())
}

// Pre: u = p_fgg.Ok(), t = p_mono.Ok(), both CanEval
func testMonomStep(verbose bool, omega fgg.Omega, p_fgg fgg.FGGProgram,
	u fgg.Type, p_mono fg.FGProgram) (fgg.FGGProgram, fgg.Type,
	fg.FGProgram) {

	// Upper-horizontal arrow
	p1_fgg, _ := p_fgg.Eval()
	VPrintln(verbose, "\nEval FGG one step: "+p1_fgg.GetMain().String())
	u1, p1_fgg := castFGG( p1_fgg.Ok(true, base.CHECK) )
	if ok, _ := u1.AssignableToDelta(p_fgg.GetDecls(), fgg.Delta{}, u); !ok { // TODO: factor out with Frontend.eval
		panic("-test-monom failed: type not preserved\n\tprev=" + u.String() +
			"\n\tnext=" + u1.String())
	}

	// Lower-horizontal arrow
	p1_mono, _ := p_mono.Eval()
	VPrintln(verbose, "Eval monom one step: "+p1_mono.GetMain().String())
	t1, p1_mono := castFG( p1_mono.Ok(true, base.CHECK) )
	u1_fg := fgg.ToMonomId(u1)
	if !t1.Equals(u1_fg) { // CHECKME: needed? or just do monom-level type preservation?
		panic("-test-monom failed: types do not match\n\tFGG type=" + u1.String() +
			" -> " + u1_fg.String() + "\n\tmono=" + t1.String())
	}

	// Right-vertical arrow
	//res := fgg.Monomorph(p1_fgg.(fgg.FGGProgram))
	res := fgg.ApplyOmega(p1_fgg.(fgg.FGGProgram), omega)
	e_fgg := res.GetMain() // N.B. the monom'd FGG expr (i.e., an FGExpr)
	VPrintln(verbose, "Monom of one step'd FGG: "+e_fgg.String())

	return p1_fgg.(fgg.FGGProgram), u1, p1_mono.(fg.FGProgram)
}

/* oblit "weak" simulation check */

func TestOblit(verbose bool, src string, steps int) {
	intrpFgg := NewFGGInterp(verbose, src, true)
	pFgg := intrpFgg.GetProgram().(fgg.FGGProgram)
	u, pFgg := castFGG( pFgg.Ok(false, base.CHECK) )
	VPrintln(verbose, "\nFGG expr: "+pFgg.GetMain().String())

	// (Initial) left-vertical arrow
	dsFgg := pFgg.GetDecls()
	pOblit := fgr.Obliterate(pFgg)
	VPrintln(verbose, "Oblit expr: "+pOblit.GetMain().String())
	t, pOblit := castFGR( pOblit.Ok(false, base.CHECK) )
	dsOblit := pOblit.GetDecls()
	VPrintln(verbose, "FGG type="+u.String()+", FGR type="+t.String())

	// One-off initial fastforward
	pOblit = ffSilent(verbose, pOblit)
	VPrintln(verbose, "Initial fast forward: "+
		pOblit.GetMain().String())

	done := steps > EVAL_TO_VAL
	var mainFgg base.Expr
	var mainOblit base.Expr
	for i := 0; i < steps || !done; i++ {
		mainFgg = pFgg.GetMain()
		mainOblit = pOblit.GetMain()
		if mainFgg.IsValue() { // N.B. IsValue -- not CanEval (checked below)
			if !mainOblit.IsValue() { // TODO: add to -test-oblit
				panic("FGG is value but monom is not:\n\tfgg = " +
					mainFgg.String() + "\n\toblit=" + mainOblit.String())
			}
			break // Both are values
		} else if mainOblit.IsValue() {
			panic("Oblit is value but FGG is not:\n\tfgg = " +
				mainFgg.String() + "\n\toblit=" + mainOblit.String())
		}
		// Both non-values, check for stuck (e.g., bad asserts -- though panic is technically not stuck)
		if mainFgg.CanEval(dsFgg) {
			if !mainOblit.CanEval(dsOblit) {
				panic("FGG is stuck but monom is not:\n\tfgg = " +
					mainFgg.String() + "\n\toblit=" + mainOblit.String())
			}
		} else {
			if mainOblit.CanEval(dsOblit) {
				panic("Oblit is stuck but FGG is not:\n\tfgg = " +
					mainFgg.String() + "\n\toblit=" + mainOblit.String())
			}
			if _, ok := mainFgg.(fgg.Assert); ok {
				if _, ok1 := mainOblit.(fg.Assert); ok1 {
					break // Both stuck on bad assert
				}
			}
		}

		// Repeat: horizontal arrows and right-vertical arrow
		pFgg, u, pOblit = testOblitStep(verbose, pFgg, u, pOblit)
	}
	VPrintln(verbose, "\nFinished:\n\tfgg="+pFgg.GetMain().String()+
		"\n\toblit="+pOblit.GetMain().String())
}

// Fast forward "silent" FGR steps
func ffSilent(verbose bool, p fgr.FGRProgram) fgr.FGRProgram {
	var foo base.Program = p
	ds := p.GetDecls()
	for isFFSilent(ds, foo.GetMain()) {
		foo, _ = foo.Eval()
		VPrintln(verbose, "Fast forward one step: "+
			foo.GetMain().String())
	}
	return foo.(fgr.FGRProgram)
}

func isFFSilent(ds []base.Decl, e base.Expr) bool {
	switch e1 := e.(type) {
	case fgr.StructLit:
		for _, v := range e1.GetElems() {
			if isFFSilent(ds, v) {
				return true
			} else if v.CanEval(ds) {
				return false
			}
		}
		return false
	case fgr.Select:
		if e1.CanEval(ds) {
			e2, _ := e1.Eval(ds)
			// N.B. the following overlaps with DropSynthAsserts -- but keep: do FF greedily first, then do DropSynthAsserts at end (o/w may need to interleave)
			if _, ok := e2.(fgr.TRep); ok { // !!! cf. nomono.fgg
				return true
			}
		}
		return isFFSilent(ds, e1.GetExpr())
	case fgr.Call:
		e2 := e1.GetReceiver()
		if isFFSilent(ds, e2) {
			return true
		} else if e2.CanEval(ds) {
			return false
		}
		for _, v := range e1.GetArgs() {
			if isFFSilent(ds, v) {
				return true
			} else if v.CanEval(ds) {
				return false
			}
		}
		return false
	case fgr.Assert:
		return isFFSilent(ds, e1.GetExpr())
	case fgr.SynthAssert:
		e2 := e1.GetExpr()
		// N.B. the following overlaps with DropSynthAsserts -- but keep: do FF greedily first, then do DropSynthAsserts at end (o/w may need to interleave)
		if e2.IsValue() {
			return true
		}
		return isFFSilent(ds, e2)
	case fgr.IfThenElse:
		return true
	case fgr.Let:
		eX := e1.GetDef()
		if eX.IsValue() {
			return true
		}
		return isFFSilent(ds, eX)
	case fgr.TRep:
		for _, v := range e1.GetArgs() {
			if v.CanEval(ds) {
				return true
			}
		}
		return false
	default: // Variable, TRep
		return false
	}
}

// HERE: refactor to fggsim (or fgrsim?) -- fix -test-oblit in Makefile to use fgrsim -- make mini examples -- check "non-empty typerep" cases for typerep-select dropping

// TODO: factor out with testMonomStep
// Pre: u = p_fgg.Ok(), t = p_mono.Ok(), both CanEval
func testOblitStep(verbose bool, pFgg fgg.FGGProgram,
	u fgg.Type, pOblit fgr.FGRProgram) (fgg.FGGProgram, fgg.Type,
	fgr.FGRProgram) {

	// Upper-horizontal arrow
	pFgg1, _ := pFgg.Eval()
	VPrintln(verbose, "\nEval FGG one step: "+
		pFgg1.GetMain().String())
	u1, pFgg1 := castFGG( pFgg1.Ok(true, base.CHECK) )
	if ok, _ := u1.AssignableToDelta(pFgg.GetDecls(), fgg.Delta{}, u); !ok { // TODO: factor out with eval
		panic("-test-monom failed: type not preserved\n\tprev=" + u.String() +
			"\n\tnext=" + u1.String())
	}

	// Lower-horizontal arrow(s)
	pTmp, _ := pOblit.Eval()
	VPrintln(verbose, "Eval oblit one step: "+pTmp.GetMain().String())
	pOblit1 := ffSilent(verbose, pTmp.(fgr.FGRProgram)) // "silent" steps

	// Right-vertical arrow
	bar := fgr.Obliterate(pFgg1.(fgg.FGGProgram))
	pFgg1Oblit := ffSilent(verbose, bar)
	eFgg1Oblit := pFgg1Oblit.GetMain().(fgr.FGRExpr) // N.B. the monom'd FGG expr (i.e., an FGExpr)
	VPrintln(verbose, "Obliterated one step'd FGG: "+
		eFgg1Oblit.String())

	eOblit1 := pOblit1.GetMain().(fgr.FGRExpr)
	fggDrop := eFgg1Oblit.DropSynthAsserts(pFgg1Oblit.GetDecls())
	oblitDrop := eOblit1.DropSynthAsserts(pOblit1.GetDecls())
	re := regexp.MustCompile("_x[0-9]*")
	fggHack := re.ReplaceAllString(fggDrop.String(), "_xxx") // !!! TODO FIXME: oblit indexes vars by a counter
	oblitHack := re.ReplaceAllString(oblitDrop.String(), "_xxx")
	if fggHack != oblitHack {
		panic("-test-oblit failed: exprs do not correspond\n\tFGG->oblit   =" +
			//fggDrop.String() + "\n\tStepped oblit=" + oblitDrop.String())
			fggHack + "\n\tStepped oblit=" + oblitHack)
	}

	return pFgg1.(fgg.FGGProgram), u1, pOblit1
}
