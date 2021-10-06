package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rhu1/fgg/internal/base"
)

const PARSER_PANIC_PREFIX = "[Parser] "

/* Test harness functions */

func parseAndType(a base.Adaptor, src string, mode base.TypingMode) base.Program {
	ast := a.Parse(true, src)
	allowStupid := false
	_, ast = ast.Ok(allowStupid, mode)
	return ast
}

func expectNoPanic(t *testing.T, src string) {
	if r := recover(); r != nil {
		t.Errorf("Unexpected panic: " + fmt.Sprintf("%v", r) + "\n" +
			src)
	}
}

func expectPanic(t *testing.T, src string, msg string) {
	if r := recover(); r == nil {
		t.Errorf("Expected panic, but none occurred: " + msg + "\n" +
			src)
	} else {
		rec := fmt.Sprintf("%v", r)
		if strings.HasPrefix(rec, PARSER_PANIC_PREFIX) {
			t.Errorf("Unexpected panic: " + rec + "\n" + src)
		}

		//t.Errorf("Unexpected panic: " + rec)
		// TODO: check panic more specifically
	}
}

func ParseAndOkGood(t *testing.T, a base.Adaptor, src string) base.Program {
	defer expectNoPanic(t, src)
	return parseAndType(a, src, base.CHECK)
}

func ParseAndInferGood(t *testing.T, a base.Adaptor, src string) base.Program {
	defer expectNoPanic(t, src)
	return parseAndType(a, src, base.INFER)
}

// N.B. do not use to check for bad *syntax* -- see the PARSER_PANIC_PREFIX panic check
func ParseAndOkBad(t *testing.T, msg string, a base.Adaptor, src string) base.Program {
	defer expectPanic(t, src, msg)
	return parseAndType(a, src, base.CHECK)
}

func ParseAndInferBad(t *testing.T, msg string, a base.Adaptor, src string) base.Program {
	defer expectPanic(t, src, msg)
	return parseAndType(a, src, base.INFER)
}

// Pre: parseAndOkGood
func EvalAndOkGood(t *testing.T, p base.Program, steps int) base.Program {
	defer expectNoPanic(t, p.String())
	allowStupid := true

	println()
	println(p.String())
	println("-----------------------------------------------------------------------")
	println()

	for i := 0; i < steps; i++ {
		println(p.GetMain().String())
		println()

		p, _ = p.Eval() // CHECKME: check rule names as part of test?

		print(">    ")
		println(p.GetMain().String())
		println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		println()

		p.Ok(allowStupid, base.CHECK)
	}
	return p
}

func EvalAndInferGood(t *testing.T, p base.Program, steps int) base.Program {
	defer expectNoPanic(t, p.String())
	allowStupid := true

	println()
	println(p.String())
	println("-----------------------------------------------------------------------")
	println()

	for i := 0; i < steps; i++ {
		println(p.GetMain().String())
		println()

		p, _ = p.Eval() // CHECKME: check rule names as part of test?

		print(">    ")
		println(p.GetMain().String())
		println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		println()

		p.Ok(allowStupid, base.INFER)
	}
	return p
}

// Pre: parseAndOkGood
func EvalAndOkBad(t *testing.T, p base.Program, msg string, steps int) base.Program {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but none occurred: " + msg + "\n" +
				p.String())
		} else {
			// PARSER_PANIC_PREFIX panic should be already checked by parseAndOkGood
			// TODO: check panic more specifically
		}
	}()
	allowStupid := true
	for i := 0; i < steps; i++ {
		p, _ = p.Eval()
		p.Ok(allowStupid, base.CHECK)
	}
	return p
}

// Pre: parseAndInferGood
func EvalAndInferBad(t *testing.T, p base.Program, msg string, steps int) base.Program {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but none occurred: " + msg + "\n" +
				p.String())
		} else {
			// PARSER_PANIC_PREFIX panic should be already checked by parseAndOkGood
			// TODO: check panic more specifically
		}
	}()
	allowStupid := true
	for i := 0; i < steps; i++ {
		p, _ = p.Eval()
		p.Ok(allowStupid, base.INFER)
	}
	return p
}
