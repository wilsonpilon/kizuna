package kaji80

import (
	"math"
	"testing"
)

func evalOK(t *testing.T, a *Assembler, expr string) float64 {
	t.Helper()
	val, ok, err := a.EvalExpr(expr)
	if err != nil {
		t.Fatalf("EvalExpr(%q) retornou erro inesperado: %v", expr, err)
	}
	if !ok {
		t.Fatalf("EvalExpr(%q) retornou ok=false, esperado ok=true", expr)
	}
	return val
}

func TestEvalExprArithmetic(t *testing.T) {
	a := NewAssembler()
	cases := map[string]float64{
		"1+2":     3,
		"10-3":    7,
		"4*5":     20,
		"20/4":    5,
		"7 MOD 3": 1,
		"-5+2":    -3,
		"+5":      5,
	}
	for expr, want := range cases {
		if got := evalOK(t, a, expr); got != want {
			t.Errorf("EvalExpr(%q) = %v, esperado %v", expr, got, want)
		}
	}
}

func TestEvalExprBitwiseAndShift(t *testing.T) {
	a := NewAssembler()
	cases := map[string]float64{
		"6&3":   2,
		"6|1":   7,
		"6^3":   5,
		"~0":    -1,
		"1<<4":  16,
		"256>>4": 16,
	}
	for expr, want := range cases {
		if got := evalOK(t, a, expr); got != want {
			t.Errorf("EvalExpr(%q) = %v, esperado %v", expr, got, want)
		}
	}
}

func TestEvalExprLogicalAndComparison(t *testing.T) {
	a := NewAssembler()
	cases := map[string]float64{
		"1==1":      1,
		"1!=1":      0,
		"3<5":       1,
		"5<=5":      1,
		"5>3":       1,
		"3>=5":      0,
		"1&&1":      1,
		"1&&0":      0,
		"0||0":      0,
		"1||0":      1,
		"NOT 0":     1,
		"NOT 5":     0,
	}
	for expr, want := range cases {
		if got := evalOK(t, a, expr); got != want {
			t.Errorf("EvalExpr(%q) = %v, esperado %v", expr, got, want)
		}
	}
}

func TestEvalExprPrecedence(t *testing.T) {
	a := NewAssembler()
	// Exemplo motivador exato do Wilson.
	got := evalOK(t, a, "((2*8)/(1+3))<<2")
	if got != 16 {
		t.Errorf("((2*8)/(1+3))<<2 = %v, esperado 16", got)
	}
	// 2+3*4 deve respeitar precedência (* antes de +), não avaliar da
	// esquerda pra direita ingenuamente.
	if got := evalOK(t, a, "2+3*4"); got != 14 {
		t.Errorf("2+3*4 = %v, esperado 14", got)
	}
}

func TestEvalExprMathFunctions(t *testing.T) {
	a := NewAssembler()
	// Exemplo motivador exato do Wilson.
	got := evalOK(t, a, "sin(pi*45.0/180.0)")
	want := math.Sqrt(2) / 2
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("sin(pi*45.0/180.0) = %v, esperado %v", got, want)
	}

	if got := evalOK(t, a, "sqrt(16)"); got != 4 {
		t.Errorf("sqrt(16) = %v, esperado 4", got)
	}
	if got := evalOK(t, a, "abs(-7)"); got != 7 {
		t.Errorf("abs(-7) = %v, esperado 7", got)
	}
	if got := evalOK(t, a, "pow(2,10)"); got != 1024 {
		t.Errorf("pow(2,10) = %v, esperado 1024", got)
	}
	if got := evalOK(t, a, "int(7.9)"); got != 7 {
		t.Errorf("int(7.9) = %v, esperado 7", got)
	}
}

func TestEvalExprFixedPoint(t *testing.T) {
	a := NewAssembler()
	// FIX(1.5) em 8.8: 1.5 * 256 = 384 = 0x0180.
	if got := evalOK(t, a, "fix(1.5)"); got != 384 {
		t.Errorf("fix(1.5) = %v, esperado 384", got)
	}
	// FIXMUL de dois valores já em ponto fixo: FIX(2.0)=512, FIX(3.0)=768.
	// (512*768)/256 = 1536 = FIX(6.0), confirma 2.0*3.0=6.0 em ponto fixo.
	got := evalOK(t, a, "fixmul(fix(2.0), fix(3.0))")
	wantFix6 := evalOK(t, a, "fix(6.0)")
	if got != wantFix6 {
		t.Errorf("fixmul(fix(2.0),fix(3.0)) = %v, esperado %v (fix(6.0))", got, wantFix6)
	}
}

func TestEvalExprRandomDeterministic(t *testing.T) {
	SeedExprRandom(42)
	a := NewAssembler()
	got := evalOK(t, a, "random(100)")
	if got < 0 || got >= 100 {
		t.Fatalf("random(100) = %v, fora do intervalo [0,100)", got)
	}
	// Mesma semente, mesma sequência -- resultado determinístico pra teste.
	SeedExprRandom(42)
	got2 := evalOK(t, a, "random(100)")
	if got != got2 {
		t.Errorf("random(100) com a mesma semente deveria repetir: %v != %v", got, got2)
	}
}

func TestEvalExprBareIdentifierFallsBackToSymbol(t *testing.T) {
	a := NewAssembler()
	_, ok, err := a.EvalExpr("MinhaLabel")
	if err != nil {
		t.Fatalf("identificador solto não reconhecido não deveria dar erro, deu: %v", err)
	}
	if ok {
		t.Fatalf("identificador solto não reconhecido deveria ter ok=false (cai pro tratamento de símbolo)")
	}
}

func TestEvalExprUnknownNameInsideExpressionErrors(t *testing.T) {
	a := NewAssembler()
	_, ok, err := a.EvalExpr("1 + MinhaLabel")
	if ok {
		t.Fatalf("aritmética com símbolo desconhecido não deveria ter ok=true")
	}
	if err == nil {
		t.Fatal("esperado erro claro pra aritmética misturando número com símbolo desconhecido, mas err foi nil")
	}
}

func TestEvalExprEquAndVariableLookup(t *testing.T) {
	a := NewAssembler()
	a.constants["MEUCONST"] = 10
	if got := evalOK(t, a, "MEUCONST*2"); got != 20 {
		t.Errorf("MEUCONST*2 = %v, esperado 20 (constante EQU referenciada em expressão)", got)
	}
	a.variables["X"] = 5
	if got := evalOK(t, a, "X+1"); got != 6 {
		t.Errorf("X+1 = %v, esperado 6 (variável reatribuível referenciada em expressão)", got)
	}
}

func TestEvalExprDivisionByZeroErrors(t *testing.T) {
	a := NewAssembler()
	if _, _, err := a.EvalExpr("1/0"); err == nil {
		t.Fatal("esperado erro de divisão por zero")
	}
}

func TestEvalExprNumericFormats(t *testing.T) {
	a := NewAssembler()
	cases := map[string]float64{
		"0x10":   16,
		"10h":    16,
		"$10":    16,
		"10o":    8,
		"1010b":  10,
		"42":     42,
	}
	for expr, want := range cases {
		if got := evalOK(t, a, expr); got != want {
			t.Errorf("EvalExpr(%q) = %v, esperado %v", expr, got, want)
		}
	}
}
