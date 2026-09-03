program Demo;

{ ============================================================
  KIZUNA demo — croqui hipotetico (nao compila / prova de conceito)
  Compilador: WIRTH80
  Banco: 0 (area comum, sempre mapeada)
  Tamanho estimado: ~12Kb (inclui tela de abertura de ~12Kb como
  recurso separado e fonte propria de ~2Kb, ver Obifile)
  ============================================================ }

{$USES ScreenFX in 'screen.asm'}   { KAJI80, banco 1, ~12Kb }
{$USES Chart    in 'chart.bas'}    { DIGNAC,  banco 2, ~5Kb  }
{$RESOURCE Font in 'font.fnt'}     { fonte propria, 2Kb, banco 2 }
{$RESOURCE Opening in 'opening.scr'} { tela de abertura, 12Kb, banco 3 }

var
  valor: Integer;

begin
  { troca para banco 3, mostra tela de abertura, volta ao banco comum }
  ScreenFX.MostrarAbertura;

  { ajusta modo de video via rotina em Assembly:
    SCREEN 2, COLOR 15,1,1, WIDTH 32, KEY OFF }
  ScreenFX.Setup;

  Write('Informe um valor: ');
  ReadLn(valor);

  { salta para o modulo BASIC Dignified (banco 2) para desenhar o grafico }
  Chart.Desenhar(valor);

  { de volta ao Pascal (banco 0) }
  WriteLn;
  WriteLn('Pressione uma tecla para sair...');
  ReadKey;
end.
