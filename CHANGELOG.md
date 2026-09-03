# Changelog

Todas as mudanças notáveis deste projeto são documentadas aqui.
Formato inspirado em [Keep a Changelog](https://keepachangelog.com/).

## [0.0.0] - 2026-09-03

### Adicionado
- Primeira versão conceitual do projeto: nome, identidade das ferramentas
  (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`, `HAKO`, `OBI`) e especificação
  técnica inicial (`SPEC.md`).
- Definição do escopo da linguagem Pascal alvo: fiel ao Turbo Pascal 4,
  com suporte a units.
- Definição do modelo de memória: MSX2, 128Kb com memory mapper, 4
  páginas de 16Kb, janela comutável na página 2.
- Desenho do formato de objeto relocável próprio `.MOB` (segmentos,
  símbolos, relocations), decidido como formato novo em vez de reusar o
  REL/M80 já existente no projeto `msxbasica`.
- Desenho do mecanismo de trampolim automático para chamadas entre
  módulos em bancos de memória diferentes.
- Definição da ABI própria do projeto: parâmetros via pilha, frame
  pointer em `IX`, limpeza de pilha pelo chamador, retorno em `A`/`HL`,
  strings no formato short string (1 byte de tamanho + dados).
- Croqui de demonstração (`demo/`) combinando Assembly, Pascal e BASIC
  Dignified num único programa hipotético, com `Obifile` de exemplo.
- Registro de direções futuras fora do escopo desta versão: dialeto
  Forth, interpretador Lisp, BASIC compilado mais amplo, COMAL.

### Notas
- Nenhum componente da toolchain está implementado nesta versão — esta
  release é puramente conceitual/documental.
