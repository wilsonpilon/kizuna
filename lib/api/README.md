# lib/api -- descritores de API da MSXLIB

Um `.api` por área diz ao DIGNAC e ao WIRTH80 **como chamar** cada rotina da
MSXLIB que usa convenção de registradores: quais parâmetros, em que registrador
cada um entra e em que registrador o valor volta. Com isso BASIC e Pascal
chamam a rotina como uma procedure/função qualquer, sem comando dedicado no
compilador. Formato e regras: `pkg/api/api.go` e `docs/manual-ferramentas.md`
("Descritores de API").

Regras para quem escreve rotinas descritas aqui:
- só uma saída em registrador é descrita (A, BC, DE ou HL); rotinas com
  várias saídas (ex.: `BDOS_FileOpen` devolve erro em A *e* handle em B) ficam
  de fora até haver sintaxe para isso;
- a rotina pode destruir qualquer registrador, **menos SP**: o compilador
  salva e restaura IX (o DIGNAC/WIRTH80 o usam como frame pointer);
- rotinas de pilha (`VDP_Line`, `VDP_BoxFill`) não cabem aqui -- já têm
  comandos próprios (`LINE`).
