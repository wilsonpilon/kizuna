# KIZUNA (絆) — Ajuda

> O conteúdo técnico deste arquivo (formato `.MOB`, sintaxe do `KAJI80`,
> `MUSUBI`, `HAKO`, `MSXLIB`, `WIRTH80`, `DIGNAC`) foi reorganizado em
> manuais separados por assunto, atualizados e mais completos que a versão
> antiga deste arquivo. Comece por `MANUAL.md` (índice geral) ou vá direto
> ao manual que precisa:

| Manual | Cobre |
| ------- | ------ |
| [`docs/manual-assembly.md`](docs/manual-assembly.md) | Assembly Z80 / `KAJI80`. |
| [`docs/manual-basic-dignified.md`](docs/manual-basic-dignified.md) | MSX-BASIC Dignified / `DIGNAC`. |
| [`docs/manual-pascal.md`](docs/manual-pascal.md) | Pascal / `WIRTH80`. |
| [`docs/manual-ferramentas.md`](docs/manual-ferramentas.md) | Formato `.MOB`/`.MAP`, `MUSUBI`, bank switching, `HAKO`/`.HLIB`, `MSXLIB`, `OBI`, `MOBDUMP`, exemplo multi-linguagem. |

Para dúvidas sobre o próprio Claude Code (a ferramenta usada para
desenvolver este projeto), use `/help` no terminal ou reporte problemas em
<https://github.com/anthropics/claude-code/issues> — isso não tem relação
com o KIZUNA em si.

## Histórico de depuração gráfica (SCREEN 2, v4.5.2)

A causa raiz do traçado bagunçado em SCREEN 2 (dois bugs independentes no
`KAJI80` e no `VDP_PSet_Raw` da `MSXLIB`) está documentada em
`docs/manual-ferramentas.md` §6.2 e no `CHANGELOG.md`.
