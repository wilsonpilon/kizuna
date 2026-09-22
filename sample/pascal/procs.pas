{ ============================================================================= }
{ KIZUNA WIRTH80 - Exemplo 3: procedure/function definidas pelo usuario         }
{ (novidade v4.9.0 -- antes o WIRTH80 so compilava um programa autocontido,     }
{ sem sub-rotinas proprias)                                                     }
{ ============================================================================= }
program Procs;

var
    resultado: Integer;

function Dobro(x: Integer): Integer;
begin
    Dobro := x * 2;
end;

procedure Saudacao(nome: Integer);
begin
    Write('Ola, sujeito numero ');
    WriteLn(nome);
end;

begin
    Saudacao(7);
    resultado := Dobro(21) + 1;
    Write('Dobro de 21 mais 1 = ');
    WriteLn(resultado);
end.
