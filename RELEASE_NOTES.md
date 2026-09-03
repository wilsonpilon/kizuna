# Release Notes — KIZUNA v0.0.0 "Hatsumusubi"

**Hatsumusubi** (初結び) — "o primeiro nó/laço amarrado". Nome escolhido em
harmonia com `MUSUBI`, o linker do projeto: esta é a primeira vez que a
ideia inteira do KIZUNA fica "amarrada" num documento coerente, mesmo sem
nenhuma linha de código funcional ainda.

## O que é esta release

Uma release **puramente conceitual**. Não há binários, não há compilador
funcionando — o que existe é a especificação completa do projeto e um
croqui de como um programa real usaria a toolchain depois de pronta.

## Destaques

- **Nome e identidade do projeto**: KIZUNA (絆, "laço/vínculo"), com as
  seis ferramentas nomeadas em torno do mesmo tema japonês retrô já usado
  no `msxbasica` (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`, `HAKO`, `OBI`).
- **Especificação técnica** (`SPEC.md`) cobrindo:
  - Escopo da linguagem Pascal (fiel ao Turbo Pascal 4, com units)
  - Modelo de memória MSX2+ de 256Kb ou mais com bank switching
  - Formato de objeto relocável próprio `.MOB`
  - Mecanismo de trampolim automático entre bancos
  - ABI própria (pilha, `IX` como frame pointer, retorno em `A`/`HL`)
- **Manual conceitual** (`MANUAL.md`) descrevendo o fluxo de uso
  pretendido e a sintaxe do `Obifile`.
- **Demo ilustrativa** (`demo/`): um programa hipotético em três
  linguagens (Pascal + Assembly + BASIC Dignified) linkado num único
  `.COM`, com receita de build de exemplo.

## O que NÃO está nesta release

- Nenhum compilador, assembler ou linker funcional.
- Nenhum teste em emulador ou hardware real.
- Formato `.MOB` ainda não tem parser/serializador de referência — só a
  descrição em `SPEC.md`.

## Próxima release planejada

Foco em `KAJI80` mínimo (assembler Z80 com subconjunto de mnemônicos) e
`MUSUBI` sem suporte a múltiplos bancos, para validar o pipeline
`.asm → .mob → .com` ponta a ponta antes de atacar `WIRTH80` e `DIGNAC`.
