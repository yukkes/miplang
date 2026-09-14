package lexer

import "testing"

func TestLexCoreSyntaxAndPositions(t *testing.T) {
	src := "set TRIPS;\nvar q[TRIPS] integer >= 0; // comment\n"
	tokens, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		kind      Kind
		lit       string
		line, col int
	}{
		{Keyword, "set", 1, 1}, {Identifier, "TRIPS", 1, 5}, {Semicolon, ";", 1, 10},
		{Keyword, "var", 2, 1}, {Identifier, "q", 2, 5}, {LBracket, "[", 2, 6},
		{Identifier, "TRIPS", 2, 7}, {RBracket, "]", 2, 12}, {Keyword, "integer", 2, 14},
		{GreaterEqual, ">=", 2, 22}, {Number, "0", 2, 25}, {Semicolon, ";", 2, 26},
	}
	if len(tokens) != len(want)+1 {
		t.Fatalf("got %d tokens, want %d (+EOF)", len(tokens), len(want))
	}
	for i, w := range want {
		got := tokens[i]
		if got.Kind != w.kind || got.Literal != w.lit || got.Line != w.line || got.Column != w.col {
			t.Fatalf("token %d = %#v, want kind=%v lit=%q @%d:%d", i, got, w.kind, w.lit, w.line, w.col)
		}
	}
	if tokens[len(tokens)-1].Kind != EOF {
		t.Fatalf("last token = %v, want EOF", tokens[len(tokens)-1].Kind)
	}
}

func TestLexReportsUnexpectedCharacter(t *testing.T) {
	_, err := Lex("set X @;")
	if err == nil {
		t.Fatal("expected lexer error")
	}
}
