// Author: Zak Nesler
// Date: 2022-05-07
//
// This file is a part of TexBot. Its job is to test the parser, ensuring that
// it correctly parses LaTeX expressions and returns the correct data, as well
// as testing the various errors that can occur.
//
// At no point are the commands used to render the image tested, just the
// parsing and configuration that is done by TexBot.

package main

import (
	"testing"
)

func TestParseSimpleExpr(t *testing.T) {
	parsed := ParseString(`$$hello$$`)

	if !parsed.HasMatch {
		t.Errorf("Expected match, none found.")
	}

	if parsed.Expr != "hello" {
		t.Errorf("Expected hello, got '%s'", parsed.Expr)
	}

	if parsed.Config.Scale != 4 {
		t.Errorf("Expected scale 4, got %d", parsed.Config.Scale)
	}
}

func TestParsingIgnoresWhitespaceOnBothSides(t *testing.T) {
	parsed := ParseString(`$$  hello   $$`)

	if !parsed.HasMatch {
		t.Error("Expected match, none found.")
	}

	if parsed.Expr != "hello" {
		t.Errorf("Expected hello, got '%s'", parsed.Expr)
	}

	if parsed.Config.Scale != 4 {
		t.Errorf("Expected scale 4, got %d", parsed.Config.Scale)
	}
}

func TestParseExprWithComplexString(t *testing.T) {
	parsed := ParseString(`$$ \frac{1}{\sqrt{2}} = \left[ \int^{10}_2 x^2 dx \right] \begin{pmatrix} 1 0 \\ 0 1 \end{pmatrix} $$`)

	if !parsed.HasMatch {
		t.Error("Expected match, none found.")
	}

	if parsed.Expr != `\frac{1}{\sqrt{2}} = \left[ \int^{10}_2 x^2 dx \right] \begin{pmatrix} 1 0 \\ 0 1 \end{pmatrix}` {
		t.Errorf("Expected complex string, got '%s'", parsed.Expr)
	}

	if parsed.Config.Scale != 4 {
		t.Errorf("Expected scale 4, got %d", parsed.Config.Scale)
	}
}

func TestParseExprWithConfig(t *testing.T) {
	parsed := ParseString(`$$hello$$[3]`)

	if !parsed.HasMatch {
		t.Error("Expected match, none found.")
	}

	if parsed.Expr != "hello" {
		t.Errorf("Expected hello, got '%s'", parsed.Expr)
	}

	if parsed.Config.Scale != 3 {
		t.Errorf("Expected scale 3, got %d", parsed.Config.Scale)
	}
}
