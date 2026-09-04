{ ============================================================================= }
{ KIZUNA WIRTH80 - Exemplo 2: Aritmética e Variáveis em Pascal                   }
{ ============================================================================= }
program Calc;

var
    a, b, c: Integer;

begin
    WriteLn('KIZUNA WIRTH80 - Teste de Aritmetica');
    WriteLn;

    a := 123;
    b := 45;
    c := a * b;

    Write('Calculando: ');
    Write(a);
    Write(' x ');
    Write(b);
    Write(' = ');
    WriteLn(c);

    WriteLn;
    WriteLn('Calculo finalizado com sucesso via Pascal!');
end.
