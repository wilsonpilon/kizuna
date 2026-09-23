{ ============================================================
  KIZUNA sample -- OBI: modulo Pascal "biblioteca" (begin end. vazio,
  logo sem Start proprio -- nenhum conflito com o Start do KAJI80).
  Compilador: WIRTH80
  Chamado a partir de main.asm (KAJI80, banco 0) via MUSUBI, que gera o
  trampolim de troca de banco automaticamente (Main -> banco 1).

  Prova real de que as 3 linguagens de entrada do KIZUNA (Assembly, BASIC
  Dignified e Pascal) linkam juntas num unico .COM multi-banco -- agora
  num banco paginavel de verdade (BANK <n> agora suportado no WIRTH80).
  ============================================================ }
program GreetLib;
BANK 1;
PUBLIC Saudacao;

procedure Saudacao(pontuacao: Integer);
begin
  Write('Pontuacao final: ');
  WriteLn(pontuacao);
end;

begin
end.
