# KIZUNA (絆)

> "Um laço entre linguagens, amarrado num único `.COM`."

**KIZUNA** é uma proposta de toolchain para MSX2+ (Z80 / MSX-DOS 2 / 256Kb
ou mais com memory mapper) que permite escrever partes de um mesmo programa em
**Assembly Z80**, **Pascal** (estilo Turbo Pascal 4, com units) e
**MSX-BASIC Dignified**, compilar cada parte separadamente e linkar tudo
num único executável — inclusive distribuindo módulos por bancos de
memória diferentes, com troca de banco resolvida automaticamente pelo
linker.

Versão Atual: `v4.5.2` — Release **Yoake (夜明け)**.

## Por quê

MSX-DOS 2 herda compatibilidade de BDOS do CP/M-80, plataforma onde
linguagens como Macro-80, Fortran-80 e Cobol-80 já linkavam juntas via um
formato de objeto comum. O Turbo Pascal 3 (até a 3.3) rodava nesse mesmo
mundo, mas **não permitia linkar com outras linguagens**; foi o Turbo
Pascal 4, com suas units e a diretiva `{$L}` para `.OBJ` externos, que
passou a **permitir** essa integração — só que ficou restrito a
MS-DOS/x86. O KIZUNA recupera essa ideia — um Pascal fiel ao TP4, gerando
código nativo para MSX2+, aproveitando os 256Kb (ou mais) via bank
switching com Memory Mapper, e permitindo combinar Assembly, Pascal e
BASIC estruturado no mesmo binário `.COM`.

## As ferramentas

| Nome      | Papel                                 | Status                            |
| --------- | ------------------------------------- | --------------------------------- |
| `KAJI80`  | Assembler Z80 modular                 | **Concluído & Validado** (v4.3)   |
| `WIRTH80` | Compilador Pascal (TP4-like)          | **Concluído & Validado** (v4.4)   |
| `DIGNAC`  | Compilador MSX-BASIC Dignified        | **Concluído & Validado** (v4.5)   |
| `MUSUBI`  | Linker com Smart-Linking e Mapper     | **Concluído & Validado** (v4.3)   |
| `HAKO`    | Bibliotecário / Empacotador (`.hlib`) | **Concluído & Validado** (v4.3)   |
| `MOBDUMP` | Inspecionador de objetos `.MOB`       | **Concluído & Validado** (v4.2)   |
| `MSXLIB`  | Biblioteca padrão (BDOS/BIOS/VDP/PSG) | **SCREEN 2 corrigida, aguardando confirmação em hardware** (v4.5.2) |
| `OBI`     | Orquestrador de build (`Obifile`)     | _Em planejamento_ (Fase 6)        |

Cada compilador/assembler gera um objeto relocável no formato próprio `.MOB`;
`MUSUBI` linka os módulos (com eliminação de código morto via Smart-Linking e
trampolins automáticos de bank switching) e produz o `.COM` final para MSX-DOS 2.

## Exemplos e Testes

- `sample/hello.asm`: Exemplo clássico em Assembly Z80.
- `sample/multibank/`: Exemplo multi-banco chaveando bancos na Página 2 (0x8000..0xBFFF).
- `sample/libdemo/`: Exemplo consumindo rotinas da biblioteca `msxlib.hlib` com Smart-Linking.
- `sample/pascal/`:
  - `hello.pas`: Primeiro Hello World em Pascal nativo para MSX (binário de apenas 326 bytes).
  - `calc.pas`: Programa demonstrando variáveis inteiras, cálculo aritmético de 16 bits e chamadas à biblioteca.
- `sample/basic/`:
  - `hello.bas`: Hello World em MSX-BASIC Dignified compilado para `.COM` (apenas 186 bytes).
  - `calc.bas`: Aritmética de 16 bits, variáveis locais e formatação de texto com smart-linking.
  - `chart.bas`: Módulo gráfico paginado no banco 2 para desenhar gráficos com `LINE`, `BF` e `PSET`.

### Estado atual da SCREEN 2 (causa raiz encontrada — aguardando confirmação em hardware)

**A causa raiz real do traçado gráfico bagunçado em SCREEN 2 foi encontrada e
corrigida na v4.5.2, após várias sessões investigando na direção errada
(timing de VRAM, atomicidade de interrupção, motor de comando de hardware do
V9938 — nenhuma delas era o problema).** Eram **dois bugs independentes**:

1. **`KAJI80` codificava operações ALU com operando indexado
   (`CP (IX+d)`, `SUB (IX+d)`, etc.) como se fossem um imediato de 8 bits —
   em silêncio, sem erro de montagem.** `CP (IX+8)` virava `CP 0` (opcode
   `FE 00`) em vez do `DD BE 08` correto. As únicas 9 ocorrências dessa forma
   no projeto inteiro estavam todas dentro de `VDP_Line`/`VDP_BoxFill`
   (`lib/src/vdp.asm`) — o que explica por que o bug só aparecia ali. Com a
   comparação de fim de linha e o cálculo de Bresenham recebendo sempre `0`
   em vez do X/Y real, `VDP_Line` desenhava uma escada de parâmetros errados
   em vez de uma linha reta, e o teste de horizontal/vertical pura nunca
   disparava — o que também explica o dump de VRAM de uma sessão anterior
   que parecia (erradamente) apontar para corrupção do registrador `DE`
   entre chamadas.
2. **`VDP_PSet_Raw` sobrescrevia o byte inteiro do padrão em vez de fazer
   leitura-modificação-escrita**, apagando os outros 7 pixels da mesma linha
   da célula 8x8 a cada ponto plotado — por isso sobrava só 1 pixel a cada 8
   em qualquer `LINE`/`BOXFILL`. Esse bug ficava mascarado pelo primeiro
   (só teria efeito visível depois de corrigi-lo).

Confirmado por remontagem e leitura direta dos bytes do `.MOB` (não apenas
análise estática); ambos corrigidos juntos. **Ainda não confirmado em
hardware/emulador real** — próximo passo é rodar `sample/basic/chart.bas`
novamente e verificar visualmente.

De brinde, a mesma auditoria encontrou (e corrigiu) mais dois pontos de
robustez no linker `MUSUBI`, nenhum deles relacionado ao bug gráfico:
- Um segmento `BSS` num banco comum não contribuía bytes reais ao `.COM`
  (por design do formato `.MOB`), mas o endereço reservava o espaço mesmo
  assim — em build multi-banco, isso deslocava para trás qualquer coisa
  posicionada depois dele (dispatcher/trampolins). Corrigido materializando
  BSS como zeros reais no binário final.
- O bootstrap multi-banco usava o número lógico de banco do linker (1, 2, 3…)
  diretamente como número de segmento físico da Memory Mapper, sem checar se
  esse segmento já estava em uso pelo MSX-DOS 2 nas Páginas 0/1/3. Agora, com
  EXTBIO disponível, cada banco pagineável é alocado via `ALL_SEG` real antes
  do uso (com fallback para o comportamento antigo se o EXTBIO não estiver
  presente).

Bugs históricos já corrigidos em sessões anteriores, mantidos aqui por
completude:
- **Dois bugs de dessincronia Pass 1/Pass 2 no assembler `KAJI80`**
  (`LD A,(rótulo)` mal dimensionado, e `rótulo: DB valor` contando o próprio
  mnemônico como dado) — cada um corrompia silenciosamente o endereço de todo
  rótulo declarado depois no mesmo módulo. O `Assemble()` verifica essa
  consistência internamente e falha a montagem com erro claro em vez de gerar
  um binário corrompido silenciosamente.
- **Um bug real no linker `MUSUBI`**: um salto relativo (`JR Z`) calculado
  errado por 1 byte no carregador multi-banco.
- **`BIOS_CHGET` só respondia à tecla ESC**, ignorando qualquer outra tecla.
- **`VDP_PSet` nunca escrevia a cor** (só o bit do pixel).

Ver `CHANGELOG.md` para o detalhamento completo.

Enquanto isso, **Assembly (KAJI80), Pascal (WIRTH80) e o restante de MSX-BASIC
Dignified (DIGNAC) continuam avançando normalmente** — o problema estava
isolado às rotinas gráficas de VDP da MSXLIB, não à toolchain em si.

O MSXgl e o Fusion-C em `resource/` são usados somente como referências de
hardware e algoritmos. A implementação final continuará na ABI própria da
MSXLIB, compartilhada por Assembly, Pascal e BASIC.

## Instalação Rápida

Baixe o pacote `kizuna-vX.Y.Z-dist.zip` na página de Releases do GitHub e execute:

```cmd
install.cmd
```

O instalador interativo em modo texto (TUI) copiará os binários, a biblioteca padrão e configurará o PATH do sistema automaticamente.

## Licença

Consulte o arquivo [LICENSE](LICENSE) para mais detalhes.
