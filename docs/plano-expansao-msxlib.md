# Plano de expansão da MSXLIB (MSXgl e Fusion-C como guia)

Objetivo (Wilson, 2026-09-23): a `MSXLIB` passa a ter **tudo o que uma biblioteca
completa de MSX precisa ter**, usando o MSXgl e o Fusion-C apenas como
**catálogo do que cobrir** — não como código a portar, converter ou linkar.
Tudo é **recriado à nossa maneira**: nossos nomes, nossa API, nossa
implementação Z80 escrita a partir do comportamento do hardware. Disponível a
Assembly, BASIC (DIGNAC) e Pascal (WIRTH80), que depois ficarão mais próximos do
MS-BASIC e do Borland Pascal 4.

Nada de `resource/msxgl` ou `resource/MSXFusionC` entra no build, na lib ou no
`.hlib`. Eles servem para responder "que rotinas existem, o que cada uma faz,
que casos de borda existem". A implementação sai dos manuais de hardware
(TMS9918, V9938/V9958, V9990, AY-3-8910, YM2413, Y8950, SCC), do que a BIOS/MSX-DOS
documentam e do comportamento observado no emulador/hardware.

Pré-requisito já entregue: MSXLIB dividida em módulos de uma rotina
(`lib/src/<area>/*.asm`), constantes em `lib/inc/*.inc` via `INCLUDE`, KAJI80 com
expressões, macros, `IF`/`REPT`, rótulos locais, `CALLBIOS`/`CALLDOS`, `INCBIN`.
O linker só traz o que o programa usa, então a lib pode crescer sem engordar os `.com`.

---

## 1. Áreas a cobrir

Lista pedida por Wilson, na ordem em que faz sentido construir. Na coluna "guia"
está o que consultar para saber **o que existe** (não para copiar):

| # | Área | Prefixo (proposta) | Guia de cobertura | Já temos |
| --- | --- | --- | --- | --- |
| 1 | **Portas** (I/O, slots, constantes de hardware) | `PORT_`, `lib/inc/*.inc` | `system_port.h`, `psg_reg.h`, `vdp_reg.h`, … | `vdp.inc`, `psg.inc`, `bdos.inc`, `bios.inc` |
| 2 | **Memória** | `MEM_` | `memory.h`, Fusion-C `Mem*`/`MMalloc` | — |
| 3 | **Matemática** | `MATH_` | `math.h`, `fixed_point.h` | `Mul16`, `Div16`, `Float_*` |
| 4 | **Strings** (+ ctype, conversões, texto/`Print`) | `STR_`/`PRINT_` | `string.h`, `print.h`, Fusion-C `Str*`/`Is*`/`Itoa`/`Locate` | `StrLen`, `StrCopy`, `StrToUpper`, `PrintHex*`, `PrintDec16*` |
| 5 | **VDP** (MSX1, V9938/V9958, comandos, paleta, sprites) | `VDP_` | `vdp.h`, Fusion-C `vdp_graph*`/`vdp_sprites` | 25 rotinas (SCREEN 2, sprites) |
| 6 | **Draw** | `DRAW_` | `draw.h`, Fusion-C `Line/Circle/Paint/Draw` | `VDP_Line`, `VDP_BoxFill`, `VDP_PSet` |
| 7 | **Tile** | `TILE_` | `tile.h` | — |
| 8 | **Scroll** | `SCROLL_` | `scroll.h`, Fusion-C `SetScroll*` | — |
| 9 | **Teclado** | `KEY_` | `keyboard.h`, Fusion-C `Inkey/GetKeyMatrix` | `BIOS_CHGET` |
| 10 | **Joystick** (+ trigger, mouse, paddle se quiser) | `JOY_` | `joystick.h`, `mouse.h`, `input.h` | — |
| 11 | **PSG** | `PSG_` | `psg.h`, Fusion-C `Sound/SetVolume/PlayEnvelope` | 6 rotinas |
| 12 | **Play** (tocador de música/efeitos) | `PLAY_` | `pt3`, `ayfx`, `vgm` (só para saber o que um player oferece) | `PSG_PlaySequence`, MML do DIGNAC |
| 13 | **MSX-Music** (YM2413) | `OPLL_` | `msx-music.h` | — |
| 14 | **MSX-Audio** (Y8950) | `Y8950_` | `msx-audio.h` | — |
| 15 | **SCC** | `SCC_` | `scc.h` | — |
| 16 | **BIOS** | `BIOS_` | `bios.h`, `bios_*.h` | 8 rotinas |
| 17 | **DOS** | `DOS_` | `dos.h`, `dos_mapper.h`, Fusion-C `fcb_*` | 6 `BDOS_File*` |
| 18 | **System** (slots, versão do MSX, CPU, interrupções) | `SYS_` | `system.h`, Fusion-C `ReadMSXtype`/`GetCPU` | — |
| 19 | **Clock** (RTC, data/hora) | `RTC_` | `clock.h`, Fusion-C `GetDate/GetTime` | — |
| 20 | **V9990** | `V9990_` | `v9990.h`, Fusion-C `g9klib` | — |

Prefixos são uma **proposta na convenção que a lib já usa** (`VDP_`, `PSG_`,
`BDOS_`, `BIOS_`); os nomes existentes não mudam (DIGNAC/WIRTH80 emitem `EXTERN`
por eles). Os nomes de cada rotina nova são nossos, escolhidos por área ao começar
a fase, não herdados.

**Fora da lista por enquanto** (existem no MSXgl, mas não foram pedidos; ficam como
"talvez depois"): `sprite_fx`, framework de jogo (`fsm`, `pawn`, `menu`), compressão
(`bitbuster`, `pletter`, `zx0`, RLE), `crypt`, `localize`, `msxi`, rede
(`obsonet`, `UNAPI`, `gr8net`), dispositivos raros (`ninjatap`, `lightgun`, `joymega`).
`crt0`/ROM/MegaROM ficam de fora de vez: nosso alvo é `.COM` de MSX-DOS 2.

## 2. Como usar o MSXgl e o Fusion-C como guia

1. **Checklist de cobertura** (`docs/cobertura.md`, gerado por script a partir dos
   `.h`): uma linha por função do MSXgl/Fusion-C, agrupada por área, com o estado
   `nossa rotina: X` / `coberta por Y` / `não faremos (motivo)`. Serve para provar
   que nada importante ficou de fora e para acompanhar o progresso. É só uma lista
   de nomes e resumos — não contém código deles.
2. **Por rotina**: olha-se o que a função deles *faz* (entradas, saídas, casos de
   borda que o autor achou importantes), e escreve-se a nossa a partir do manual do
   hardware. Onde a nossa API for mais simples, mais rápida ou mais próxima do
   BASIC/Pascal, ela é a nossa — não precisa espelhar a assinatura deles.
3. **Onde eles divergirem do hardware/documentação**, vale o hardware. O MSXgl também
   tem suposições próprias (por exemplo VRAM de 16K vs 128K); os testes é que decidem.

## 3. Decisões de arquitetura (Fase 0)

### D1. Nomes e compatibilidade
Convenção `AREA_Verbo...` (a da lib atual). Os 120 símbolos existentes ficam
como estão. Colisão só é possível com nomes futuros nossos; o teste da lib já
rejeita `PUBLIC` duplicado.

### D2. Convenção de chamada e descritores de API (implementado)
Hoje as rotinas da MSXLIB usam **registradores** e só são alcançáveis pelo
BASIC/Pascal via comandos dedicados (`PSET`, `LINE`, `WriteLn`…), limitação
registrada em `docs/manual-pascal.md` §5. Com centenas de rotinas novas, não dá para
escrever um comando de compilador para cada uma.

Proposta: um **descritor de API por área** (`lib/api/<area>.api`, texto simples)
declarando, por rotina pública: nome, parâmetros com o registrador de cada um,
retorno, registradores destruídos e o nome BASIC/Pascal opcional. DIGNAC e WIRTH80
leem o `.api` e geram a carga de registradores na chamada. Resultado: Assembly usa
a rotina direto; BASIC e Pascal ganham cada rotina nova sem tocar no compilador por
rotina; o `.api` alimenta também a documentação. Alternativa descartada: reescrever
tudo em ABI de pilha (mais lenta e quebra o que existe). É a **Fase 0b** e deve ser
provada com 3–4 rotinas antes de escalar.

### D3. Estado compartilhado
Vários recursos exigem RAM de estado (cópia dos registradores do VDP — que são
write-only —, modo de tela atual, cursor do texto, semente aleatória, heap). Um módulo de
dados por área exporta as células e as rotinas usam `EXTERN`. Em `.COM` a RAM é
gravável, então células `DB`/`DW` no próprio módulo funcionam (o que a lib já faz).

### D4. Origem do código e créditos
Como as rotinas são escritas por nós a partir de documentação de hardware, a
implementação é original e sob a licença do projeto (GPL-3). Mesmo assim vale manter
um `docs/CREDITS.md` reconhecendo MSXgl e Fusion-C como referência de cobertura —
é honesto e barato. Para não borrar essa linha: não copiar trechos de `.c`/`.asm`
deles para dentro da lib, nem estruturas internas que não sejam o nome público
óbvio de uma função de hardware. (Ambos são CC BY-SA 4.0; copiar código deles
traria obrigações de licença que, seguindo este plano, não existem.)

### D5. Verificação de cada rotina
O projeto já achou bugs que só o emulador pegou (`LD B,(nn)`→`LD B,0`, `CP` sem
sinal), então "parece certo" não basta:
- O harness Z80 (hoje no scratchpad) vai para dentro do repositório, como
  dependência de **teste**, com traps de BDOS/CALSLT e registro de portas.
- Cada módulo tem teste de tabela em Go: entradas conhecidas → saídas/efeitos
  esperados. Para VDP/PSG/OPLL/SCC, o oráculo é o **traço de escritas de porta**
  derivado do manual do chip (não do código do MSXgl).
- A cada fase, um `sample/` de demonstração para você validar em **hardware real**,
  no formato dos `sample/macroasm/`.

Rotina "pronta": módulo monta, entrada no `.api`, teste Go passa, documentada,
marcada no `cobertura.md`.

## 4. Camadas de linguagem (depois da biblioteca)

**Ordem decidida por Wilson (2026-09-23):** primeiro toda a API/LIB para Assembly e
chamadas diretas (Fases 1–7); só então as linguagens, por cima, com *aliases* — a MSXLIB
continua com o nome nosso (`VDP_SetColor`) e cada dialeto é uma fachada. Meta final: cada
dialeto oferece **tudo o que o guia (MSXgl/Fusion-C) oferece**, na sua sintaxe:

| Dialeto | Linguagem | Como cada função do guia aparece |
| --- | --- | --- |
| **MS-BASIC** | DIGNAC | comandos e funções no estilo MSX-BASIC/Microsoft: `COLOR 15,4`, `SOUND 7,62`, `LOCATE`, `VPOKE`, `PUT SPRITE`, `CIRCLE`, `PAINT`, `STICK()`, `INKEY$`… |
| **Turbo Pascal 4** | WIRTH80 | *units* no estilo Borland: `Crt` (o análogo de conio), `Graph` (o análogo da BGI), `Dos`… com `uses` |
| **asMSX-style** | KAJI80 | macros/diretivas: `CALLBIOS`, `CALLDOS` e macros por área (`SETCOLOR 15,4`…) sobre as mesmas rotinas |

**BASIC vem primeiro** entre as linguagens. Peças necessárias (detalhadas na conversa de
2026-09-23, ainda a implementar depois das Fases 1–7):
1. forma `command basic NOME = ROTINA` no `.api` e parser do DIGNAC aceitando `NOME expr, expr`
   (sem parênteses) para comandos registrados — o parser precisa conhecer os descritores;
2. parâmetros **opcionais com valor padrão** e argumento omitido (`COLOR ,4`);
3. rotinas "fachada" na lib quando o MS-BASIC faz mais que uma chamada (`COLOR` também grava
   `FORCLR/BAKCLR/BDRCLR`; `SCREEN` reinicializa a tela);
4. ligar o `COLOR` que já existe no lexer, criar `SOUND` etc.

5. **`STRING` como valor de primeira classe** no DIGNAC (hoje só atribuição e `PRINT`; o compilador recusa
   `a$` como argumento de chamada): passar `STRING` e literais de string como argumento (`ptr`), devolver `STRING`
   de função e escrever `LEFT$`/`MID$`/`INSTR`/`VAL`/`UCASE$` como expressões — as rotinas `STR_*` já existem e
   operam no formato das `STRING` do DIGNAC (tamanho + dados); falta o compilador ligar uma coisa à outra.

Depois, no WIRTH80: `uses` + `unit` no `.api` (escopo dos nomes), estado das units
(`TextColor`, `SetColor`, `GotoXY`) e mais tipos (`array`, `record`, `real`).

Critério de cobertura por dialeto: cada item de `docs/cobertura.md` ganha uma coluna por
dialeto ("BASIC", "Pascal", "asMSX") quando a fase de linguagens começar.

## 5. Fases

Cada fase é um marco: commit, CHANGELOG, testes, `sample/` para hardware.

- **Fase 0 — Fundação** — ✅ **concluída (2026-09-23)**:
  (a) `lib/inc/` ganhou `ports.inc`, `vdpreg.inc`, `psgreg.inc`, `opll.inc`, `y8950.inc`,
  `scc.inc`, `v9990.inc` (só portas), `ascii.inc`, `color.inc` — escritos por nós e
  conferidos contra os cabeçalhos-guia; um teste de montagem inclui todos juntos e confere
  ~50 valores de hardware conhecidos;
  (b) descritores `.api` (`pkg/api`, `lib/api/*.api`) lidos por DIGNAC, WIRTH80 e OBI
  (`-api`, `api:`), provados com **rotinas reais** (VDP, PSG, BDOS, BIOS, matemática) e
  rodando no simulador — ver `docs/manual-ferramentas.md` §6.3 e `sample/api/`;
  (c) simulador Z80 em `pkg/z80sim` + auxiliares de teste em `pkg/libtest` (monta a
  MSXLIB fresca de `lib/src`, liga, roda, registra console/portas/CALSLT);
  (d) `docs/cobertura.md` (gerado por `go run ./tools/cobertura` a partir de
  `docs/cobertura.map`) e `docs/CREDITS.md`.
  Decisões tomadas no caminho: o compilador salva/restaura **IX** em volta de toda
  chamada por `.api` (rotinas que usam BIOS destroem IX); o carregamento de parâmetros de
  8 bits escolhe, a cada `POP`, um par que ainda não tenha registrador carregado (ver
  `Routine.plan`), o que dispensou a regra "um par de rascunho livre" do desenho inicial
  e permite assinaturas reais como `BDOS_FileRead(B, DE, HL)`.
- **Fase 1 — Núcleo sem hardware**: memória, matemática, strings/ctype/conversões.
  Ideal para exercitar o processo (sem dependência de VDP). Inclui aleatório,
  `Mul32/Div32`, cópia/preenchimento (com variantes rápidas `LDIR`/`LDDR`), heap simples.
- **Fase 2 — VDP e texto**: modos MSX1/2/2+, registradores com cópia sombra, VRAM
  (incluindo 128K), paleta, tabelas, sprites, motor de comandos V9938/V9958, `Print`/
  `Locate`, formatação. Marco visível: hello world gráfico em SCREEN 1/2/5.
  Andamento: **2a e 2b feitas** (registradores/sombra, VRAM 128K, paleta, `VDP_SetMode`, tabelas,
  sprites modo 1/2, piscar, rolagem; testadas no simulador com modelo do V9938). Faltam 2c (motor de
  comandos), 2d (`Print` em VRAM com a fonte da ROM).
- **Fase 3 — Draw, Tile, Scroll**: linha/caixa/círculo/preenchimento/`Paint`/mini-`DRAW`;
  bancos de tiles e mapas; scroll de hardware. Lembrete: o motor de comandos do V9938
  **não funciona em SCREEN 2**, só em SCREEN 5–8 — os testes de hardware são em SCREEN 5+.
- **Fase 4 — Entrada**: teclado (varredura de matriz, buffer, teclas mortas), joystick,
  gatilhos, mouse.
- **Fase 5 — Som**: PSG completo (tom/ruído/mixer/envelope), depois um **player**
  próprio (sequências/MML e efeitos), MSX-Music (YM2413), MSX-Audio (Y8950), SCC.
  Detecção dos cartuchos depende de slots (Fase 6). Sem hardware, só o traço de portas
  é verificável; o som em si só o hardware confirma.
- **Fase 6 — Sistema**: BIOS (wrappers via `CALLBIOS`, incluindo sub-ROM), DOS (arquivos,
  diretórios, ambiente, mapper de memória, FCB do DOS 1), System (slots, versão do MSX,
  CPU/Turbo R, interrupções), Clock (RTC e data/hora). O risco é o DOS 2 e o mapper,
  que só o hardware/openMSX confirmam.
- **Fase 7 — V9990**: modos, VRAM, paleta, blitter, sprites, scroll. Só vale se houver
  GFX9000 para testar; por isso fica por último.

## 6. Riscos

1. **Convenção de chamada (D2)**: se errada, refaz-se muito. Mitigação: provar na Fase 0b.
2. **Rotinas de hardware sem oráculo** (SCC, OPLL, Y8950, V9990): o emulador só verifica
   o traço de portas. Mitigação: `sample/` por fase para teste no hardware real.
3. **Escrever tudo do manual**: mais trabalho que portar, mas evita herdar suposições
   alheias e mantém a lib coerente com o resto do KIZUNA. Os testes de tabela são
   obrigatórios por isso.
4. **Escopo de linguagem**: MS-BASIC/Borland Pascal completos são projetos por si só;
   a lib não espera por eles (trilha paralela).

## 7. Primeiro passo concreto (proposta)

1. Gerar `docs/cobertura.md` (checklist por área, a partir dos `.h`).
2. `lib/inc/`: constantes de portas/registradores.
3. `.api` + suporte mínimo em DIGNAC e WIRTH80, provado com 3–4 rotinas.
4. Harness Z80 no repositório + primeiro teste de tabela.
5. Fase 1 começando por **matemática** (aleatório, abs, div/mod, shifts com sinal),
   escrita por nós do zero.
