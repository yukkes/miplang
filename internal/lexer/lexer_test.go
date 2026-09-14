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

func TestLexAMPLCoreSyntax(t *testing.T) {
	src := "# model comment\nsubject to Balance {i in I}: x[i] = y[i] / 2;\nparam p {I} > 0, < 10;\n"
	tokens, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		kind Kind
		lit  string
	}{
		{Keyword, "subject"}, {Keyword, "to"}, {Identifier, "Balance"}, {LBrace, "{"},
		{Identifier, "i"}, {Keyword, "in"}, {Identifier, "I"}, {RBrace, "}"}, {Colon, ":"},
		{Identifier, "x"}, {LBracket, "["}, {Identifier, "i"}, {RBracket, "]"}, {Equal, "="},
		{Identifier, "y"}, {LBracket, "["}, {Identifier, "i"}, {RBracket, "]"}, {Slash, "/"},
		{Number, "2"}, {Semicolon, ";"},
		{Keyword, "param"}, {Identifier, "p"}, {LBrace, "{"}, {Identifier, "I"}, {RBrace, "}"},
		{Greater, ">"}, {Number, "0"}, {Comma, ","}, {Less, "<"}, {Number, "10"}, {Semicolon, ";"},
	}
	if len(tokens) != len(want)+1 {
		t.Fatalf("got %d tokens, want %d (+EOF): %#v", len(tokens), len(want), tokens)
	}
	for i, w := range want {
		if tokens[i].Kind != w.kind || tokens[i].Literal != w.lit {
			t.Fatalf("token %d = %#v, want kind=%v lit=%q", i, tokens[i], w.kind, w.lit)
		}
	}
}

func TestAMPL_LEX_004_BlockComments(t *testing.T) {
	src := "set I; /* comment\nacross lines */ param p {I};\n"
	tokens, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"set", "I", ";", "param", "p", "{", "I", "}", ";"}
	if len(tokens) != len(want)+1 {
		t.Fatalf("tokens=%#v", tokens)
	}
	for i, lit := range want {
		if tokens[i].Literal != lit {
			t.Fatalf("token %d=%q, want %q", i, tokens[i].Literal, lit)
		}
	}
}

func TestAMPL_LEX_004_UnterminatedBlockCommentIsError(t *testing.T) {
	_, err := Lex("set I; /* never closed")
	if err == nil {
		t.Fatal("expected unterminated block comment error")
	}
}

func TestAMPL_LEX_007_ScientificNumbers(t *testing.T) {
	tokens, err := Lex("param a >= 1e3; param b <= 2.5E-2;")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tok := range tokens {
		if tok.Kind == Number {
			got = append(got, tok.Literal)
		}
	}
	if len(got) != 2 || got[0] != "1e3" || got[1] != "2.5E-2" {
		t.Fatalf("numbers=%v", got)
	}
}

func TestAMPL_LEX_001_NonASCIIIdentifierIsRejected(t *testing.T) {
	_, err := Lex("set 配送;")
	if err == nil {
		t.Fatal("expected non-ASCII identifier error")
	}
}
