package lanzou

import (
	"strings"
	"testing"
)

func TestCalcAcwScV2(t *testing.T) {
	const arg1 = "0CCD57BCF5ACEE5B2DF878457E6DCF03D1AF6954"
	const want = "6ac5cc695f752682786f5cdfd73da6bd024d15ba"
	tests := []struct {
		name string
		html string
	}{
		{
			name: "fixture",
			html: "<script>var arg1='" + arg1 + "';</script>",
		},
		{
			name: "double quotes",
			html: `arg1="` + arg1 + `";`,
		},
		{
			name: "lowercase hex",
			html: "arg1='" + strings.ToLower(arg1) + "';",
		},
		{
			name: "mixed case hex",
			html: "arg1='" + strings.ToLower(arg1[:20]) + arg1[20:] + "';",
		},
		{
			name: "assignment case",
			html: "ARG1='" + arg1 + "';",
		},
		{
			name: "assignment whitespace",
			html: "<script>\nvar\targ1 \r\n=\t\"" + arg1 + "\" ;\n</script>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcAcwScV2(tt.html)
			if err != nil {
				t.Fatalf("CalcAcwScV2() error = %v", err)
			}
			if got != want {
				t.Fatalf("CalcAcwScV2() = %q, want %q", got, want)
			}
		})
	}
}

func TestCalcAcwScV2ZeroPadding(t *testing.T) {
	tests := []struct {
		name string
		arg1 string
		want string
	}{
		{
			name: "zero bytes",
			arg1: strings.Repeat("0", 40),
			want: "3000176000856006061501533003690027800375",
		},
		{
			name: "leading zero digit",
			arg1: strings.Repeat("3", 40),
			want: "0333245333b653353526326003305a3314b33046",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcAcwScV2("arg1='" + tt.arg1 + "';")
			if err != nil {
				t.Fatalf("CalcAcwScV2() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("CalcAcwScV2() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalcAcwScV2InvalidInput(t *testing.T) {
	const arg1 = "0CCD57BCF5ACEE5B2DF878457E6DCF03D1AF6954"
	tests := []struct {
		name string
		html string
	}{
		{name: "empty page", html: ""},
		{name: "missing assignment", html: "<html>acw_sc__v2</html>"},
		{name: "empty value", html: "arg1='';"},
		{name: "empty double quoted value", html: `arg1="";`},
		{name: "short value", html: "arg1='00';"},
		{name: "odd length", html: "arg1='" + arg1[:39] + "';"},
		{name: "too long", html: "arg1='" + arg1 + "00';"},
		{name: "nonhex value", html: "arg1='" + arg1[:39] + "G';"},
		{name: "nonhex double quoted value", html: `arg1="` + arg1[:39] + `g";`},
		{name: "embedded whitespace", html: "arg1='" + arg1[:20] + " " + arg1[21:] + "';"},
		{name: "unquoted value", html: "arg1=" + arg1 + ";"},
		{name: "missing equals", html: "arg1 '" + arg1 + "';"},
		{name: "comparison", html: "arg1=='" + arg1 + "';"},
		{name: "missing closing quote", html: "arg1='" + arg1 + ";"},
		{name: "mismatched single quote", html: "arg1='" + arg1 + "\";"},
		{name: "mismatched double quote", html: "arg1=\"" + arg1 + "';"},
		{name: "different variable", html: "otherarg1='" + arg1 + "';"},
		{name: "dollar variable prefix", html: "$arg1='" + arg1 + "';"},
		{name: "variable suffix", html: "arg10='" + arg1 + "';"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcAcwScV2(tt.html)
			if err == nil {
				t.Fatal("CalcAcwScV2() error = nil, want invalid challenge error")
			}
			if got != "" {
				t.Fatalf("CalcAcwScV2() = %q, want empty result", got)
			}
		})
	}
}
