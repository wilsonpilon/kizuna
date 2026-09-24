# Retomada — MSXLIB Fase 2d (texto em VRAM)

Resumo de encerramento de **2026-09-24**. Para continuar mais tarde no mesmo dia: leia este
arquivo, confira `git log --oneline -8` e comece pelo passo **2d** (seção 3).

> **Combinado com Wilson:** por enquanto **não gerar pacote nem build de release** (sem bump de
> versão, sem `RELEASE_NOTES`, sem `build.ps1` da raiz, sem `installer`). Só código, testes,
> documentação e commits/push. O CHANGELOG está na seção `[Não lançado]`; a versão continua
> **v4.11.0**. A decisão de fechar uma release fica para depois da 2d (ou quando Wilson pedir).

## 1. Onde estamos

Plano geral: `docs/plano-expansao-msxlib.md` (fases 0-7 + camadas de linguagem). MSXgl e Fusion-C
são só **catálogo** do que cobrir — API, nomes e código são nossos, recriados a partir da
documentação de hardware. Cobertura frente ao catálogo: `docs/cobertura.md` (gerada por
`go run ./tools/cobertura`, validada por `-check`; entradas em `docs/cobertura.map`).

| Fase | Conteúdo | Estado |
| ---- | -------- | ------ |
| 0 | infra: descritores `.api`, simulador Z80 (`pkg/z80sim`), `lib/inc`, plano | feita |
| 1 | matemática, memória (heap), texto (`CHAR/CSTR/STR/NUM/CON`) | feita, só simulador |
| 2a | VDP: registradores com sombra, VRAM 128 KB, paleta, `VDP_SetMode` | feita, só simulador |
| 2b | tabelas, sprites modo 1/2, piscar, rolagem, bits de modo | feita, só simulador |
| 2c | motor de comandos V9938/V9958 (`VDP_Cmd*`, `VDP_Hw*`) | feita; **`sample/vdpcmd` OK em hardware/openMSX (2026-09-24)** |
| **2d** | **texto em VRAM (`Print`, fonte da ROM, cursor, cores)** | **próxima** |
| 3-7 | desenho (círculo/`Paint`/`DRAW`), tiles/scroll, entrada, som, sistema/BIOS/DOS/relógio, V9990 | pendentes |
| camadas | MS-BASIC (DIGNAC), Turbo Pascal 4 (`Crt`/`Graph`), macros KAJI80 | depois da biblioteca |

Commits desta rodada (todos em `origin/main`): `47ac606` (KAJI80: imediatos/`DS` com expressão),
`3067cdd` (2a), `b813200` (2b), `a25a772` (2c) e o commit de encerramento (documentação e este
arquivo — `git log -1`).

### O que a 2c confirmou em hardware (e o que não)

`sample/vdpcmd` (SCREEN 5) rodou **OK**: `HwFillRect`, `HwBox`, `HwLine` (**desenha NX+1 pontos**,
NX = lado maior — hipótese que estava em aberto), `HwPlot`, `HwBoxFill`, `HwCopyRect`,
`HwLoadRect` e `HwReadRect` (leitura de volta bateu).

**Ainda só no simulador** (candidatos a um segundo sample de hardware — sugestão `sample/vdpbasic`
para a 2a/2b e um complemento do `vdpcmd`): `HwFillRectFast`, `HwMoveRect`, `HwCopyLines`,
`HwLoadFast`, `HwSearch`, `HwPoint`, `CmdStop`, operações lógicas além da cópia; toda a 2a; toda a
2b. Pontos duvidosos declarados nos comentários: `VDP_SetHScrollCoarse/Fine` e `VDP_SetAdjustRaw`
(valor cru do registrador, sentido do deslocamento em pixels não conferido).

## 2. Como o trabalho é feito (para não redescobrir)

- Testes: `go test ./...` (tudo verde no encerramento). Testes de rotina: `pkg/msxlib/*_test.go`
  com `libtest.NewRunner` (liga só as rotinas pedidas + dependências, chama com registradores de
  entrada, devolve registradores) e um VDP de simulação (`vdpRunner`). Padrão: referência
  independente em Go, `keep()` para conferir preservação de registradores, `canary()` para valores
  marcadores e **verificação por mutação** (quebrar o assembly de propósito e ver o teste falhar).
- Simulador do VDP: `pkg/z80sim/vdp.go` (VRAM 128 KB, endereço de 17 bits, status, paleta, R#17,
  `VBlankEvery`, `JiffyEvery`) e `vdpcmd.go` (motor de comandos, SCREEN 5-8). O modelo **não**
  atualiza DY/SY/NY ao fim do comando e ignora comandos fora dos modos bitmap.
- Um módulo `.asm` por rotina/família em `lib/src/<area>/`; descritores para BASIC/Pascal em
  `lib/api/*.api` (rotinas com duas saídas, como os getters de tabela e `VDP_HwSearch`, ficam só
  em Assembly). Reconstruir a biblioteca: `pwsh -File lib/build.ps1` (`lib/msxlib.hlib` está
  versionado — **não reconstruir nesta pausa a menos que mude código**).
- **Os geradores Python das fases 2a-2c não estão no repositório** (ficaram no scratchpad da
  sessão). Os `.asm` gerados são a fonte da verdade: edite-os direto.
- KAJI80 (assembler): sem `Label+N` em operando com símbolo; `LD B,(DE)` é inválido; erros de
  operando agora são altos (o bug antigo do `parseImm8`, que virava 0 em silêncio, foi corrigido em
  `47ac606`). Convenção nas rotinas: `EI` ao terminar, preservar `A,BC,DE,HL` salvo doc em contrário.
- Ambiente (Windows + Git Bash): heredocs com `\n`/aspas quebram — use Write/Edit e scripts
  Python em arquivo; nomes reservados do Windows (`con`, `aux`, `nul`...) não podem ser nome de
  arquivo/pasta (por isso `lib/src/console`); reconstruir samples muda binários — `git checkout --
  sample` se não for intencional; o aviso de CRLF do git é ruído.

## 3. Passo 2d — texto em VRAM (proposta de desenho)

Objetivo (marco do plano): "hello world" gráfico em SCREEN 1/2/5, com `Print`/`Locate`/cores,
cobrindo os itens `Print_*` do catálogo (conferir `grep "Print_" docs/cobertura.md`).

1. **Fonte.** O MSX guarda a fonte de 8x8 (256 caracteres, 2 KB) na Main-ROM; o endereço está na
   palavra em `0004h`. Em MSX-DOS a página 0 é RAM, então a leitura é inter-slot (`RDSLT` da BIOS
   via `BIOS_Call`/`CALSLT`; o slot da ROM está em `EXPTBL`). Rotinas: `VDP_FontLoad` (ROM → tabela de
   padrões na VRAM, incluindo as 3 partes do SCREEN 2/4) e cópia para um buffer em RAM para o
   desenho em modos bitmap. **O simulador precisa de um mock de `RDSLT`** (já existe o trap de
   `CALSLT`; o teste fornece os 2 KB de fonte). Alternativa a discutir com Wilson: uma fonte
   própria embutida na biblioteca (independe de slot, mesmo visual em qualquer MSX).
2. **Modos de texto/padrão (SCREEN 0, 1, 2, 3, texto 80 col).** Cursor (X,Y) em variáveis do
   módulo, `VDP_Locate`, `VDP_PrintChar`, `VDP_PrintCStr`, números (`Dec16`, `Hex8/16`), `VDP_Cls`
   (limpa a tabela de nomes), rolagem de tela; cores por `VDP_TextColor`/`VDP_Backdrop`/tabela de
   cores. Usa os endereços de tabela da 2b (`VDP_GetNameTable` etc.) — funciona com tabelas
   movidas.
3. **Modos bitmap (SCREEN 5-8).** `VDP_DrawChar(x,y)` / `DrawText` / `DrawDec` / `DrawHex` com
   cor de frente e de fundo (fundo transparente = operação `TIMP`): expandir o glifo em pixels e
   mandar por `VDP_HwLoadRect` (LMMC) — ou por linhas com `HMMC` quando alinhado. Reaproveita a 2c.
4. **Testes.** Rasterizador de referência em Go (fonte → imagem esperada) comparado com a VRAM do
   simulador, nos 4 modos bitmap e nos de tabela de nomes; mutações; preservação de registradores.
5. **Sample de hardware** (`sample/vdptext`), no mesmo formato do `vdpcmd`: texto em SCREEN 1, 5 e
   80 colunas, com a frase de verificação impressa ao sair.
6. Fechar como nas outras: `.api` (o que couber), `docs/cobertura.map` (só equivalências reais),
   `docs/manual-ferramentas.md` §6.1.2, `CHANGELOG.md`, `docs/plano-expansao-msxlib.md`,
   `lib/build.ps1` para reconstruir o `.hlib`, `go vet ./... && go test ./...`, commit e push.

## 4. Depois da 2d

- (Recomendado) um sample de hardware para a 2a/2b, para tirar o "só simulador" delas.
- Fase 3: círculo, `Paint` (preenchimento por SRCH), mini-`DRAW`; tiles e mapas; scroll de
  hardware (lembrete: comandos só em SCREEN 5+).
- Fases 4-7: entrada (teclado/joystick/mouse), som (PSG completo, player, MSX-Music, MSX-Audio,
  SCC), sistema/BIOS/DOS/relógio, V9990.
- Camadas de linguagem por aliases: MS-BASIC no DIGNAC primeiro (`SCREEN`, `COLOR`, `LINE`, `PSET`,
  `SOUND`...), depois Turbo Pascal 4 (`Crt`/`Graph`) no WIRTH80 e macros no KAJI80.
- Só então decidir a release (versão, `RELEASE_NOTES.md`, build e pacote).

## 5. Estado técnico no encerramento

- Branch `main`, sincronizada com `origin/main`; `go vet ./...` e `go test ./...` limpos.
- `lib/msxlib.hlib` reconstruída na 2c (contém todas as rotinas 2a-2c); nada mudou desde então além
  de comentários.
- Sem versão nova, sem pacote, sem build de release.
