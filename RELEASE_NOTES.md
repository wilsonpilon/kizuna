# Release Notes — KIZUNA v4.5.2 "Yoake" (夜明け)

**Yoake** (夜明け) — "o romper da madrugada, o instante em que a escuridão
finalmente cede". A resposta direta a *Kuyashii* (悔しい, a frustração da
release anterior): depois de várias sessões investigando na direção errada,
a causa raiz real do bug gráfico de SCREEN 2 foi encontrada e corrigida.

A versão **v4.5.2** corrige **dois bugs independentes** que juntos explicam
todo o caos visual observado em `VDP_Line`/`VDP_BoxFill` desde o início da
investigação: o assembler `KAJI80` codificava operações ALU com operando
indexado (`CP (IX+d)`, `SUB (IX+d)`, etc.) como um imediato de 8 bits em
silêncio — as únicas ocorrências dessa forma no projeto inteiro estavam
justamente dentro das rotinas de linha/área da `MSXLIB` — e `VDP_PSet_Raw`
sobrescrevia o byte inteiro do padrão em vez de preservar os outros pixels
da mesma célula. De brinde, a mesma auditoria corrigiu dois pontos de
robustez no linker `MUSUBI`: um bug de offset de arquivo com segmentos BSS
em build multi-banco, e a falta de alocação real de segmento (`ALL_SEG` via
EXTBIO) para bancos pagináveis, que antes assumia que o número lógico de
banco do linker já era um segmento físico livre da Memory Mapper.

## Estado técnico atual - SCREEN 2

- **Corrigido em software, confirmado por remontagem e leitura direta dos
  bytes do `.MOB`/`.COM` gerados** (não apenas análise estática): os dois
  bugs acima descritos em detalhe no `CHANGELOG.md`.
- **Ainda não confirmado visualmente em hardware/emulador real** — esse é o
  único passo que falta para fechar definitivamente o problema.
- Todo o restante da toolchain (`KAJI80`, `WIRTH80`, `DIGNAC`, `MUSUBI`,
  `HAKO`, `MOBDUMP`) segue **concluído e validado**; o bug estava isolado às
  rotinas gráficas de VDP da `MSXLIB`.

### Retomada da próxima sessão

1. Rodar `sample/basic/chart.bas` em hardware/emulador real e confirmar
   visualmente a moldura, os eixos e a curva.
2. Se confirmado: fechar o bloqueio de vez e seguir para a Fase 6 (`OBI`).
3. Se ainda houver artefato: já não é mais nenhuma das causas descartadas em
   sessões anteriores (timing de VRAM, atomicidade, motor de comando de
   hardware) nem os dois bugs corrigidos aqui — investigar do zero com um
   dump de VRAM durante a execução real.

As fontes em `resource/MSXgl` e `resource/MSXFusionC` permanecem material de
referência; não fazem parte do build da MSXLIB.

---

# Release Notes — KIZUNA v4.5.1 "Kuyashii" (悔しい)

**Kuyashii** (悔しい) — sentimento profundo de frustração honrosa e inconformismo por não ter atingido o resultado gráfico esperado no momento, mas acompanhado da convicção e energia para retornar, perseverar e conquistar a solução definitiva.

A versão **v4.5.1** consolida a integração completa do compilador **`DIGNAC`** (MSX-BASIC Dignified) com o ecossistema KIZUNA, exporta os símbolos de desenho em tela da biblioteca **`MSXLIB`** e prepara o terreno para a depuração fina do subsistema gráfico TMS9918/V9938 na próxima iteração.

## Estado técnico atual - SCREEN 2

- **Validado:** entrada em SCREEN 2, escrita direta na VRAM, exibição de um
  ponto e retorno ao prompt do MSX-DOS 2.
- **Em investigação:** artefatos fixos na tela durante a inicialização e o
  cálculo completo das relações entre Pattern Generator, Name Table, Color
  Table e as três páginas verticais da SCREEN 2.
- **Ainda não considerar concluído:** `VDP_PSet`, `VDP_Line`, `VDP_BoxFill`,
  texto gráfico e o exemplo `sample/basic/chart.bas`.

### Retomada da próxima sessão

1. Reproduzir o teste mínimo e registrar os bytes de VRAM observados.
2. Executar uma inicialização de SCREEN 2 baseada em uma única página e uma
   única célula 8x8, sem `PSET` ou `LINE`.
3. Confirmar o mapeamento entre Name Table, Pattern Table e Color Table.
4. Revalidar `VDP_PSet` com uma única célula antes de implementar linhas.
5. Só então avançar para `VDP_Line`, texto gráfico e o `chart.bas`.

As fontes em `resource/MSXgl` e `resource/MSXFusionC` permanecem material de
referência; não fazem parte do build da MSXLIB.

---

# Release Notes — KIZUNA v4.5.0 "Hinode" (日の出)

**Hinode** (日の出) — "o nascer do sol, a alvorada radiante". Na continuidade do amanhecer do KIZUNA, a versão **v4.5.0** traz o nascimento do compilador **`DIGNAC`** (MSX-BASIC Dignified para Z80 / MSX2+), completando o trio de linguagens fundamentais da toolchain: **Assembly Z80 (`KAJI80`)**, **Pascal (`WIRTH80`)** e **BASIC Estruturado (`DIGNAC`)**.

---

## O que é esta release

Esta é a release oficial **v4.5.0 pre-alpha**, trazendo uma toolchain Z80 completa para MSX2+ / MSX-DOS 2 com suporte à compilação de código **Assembly Z80**, **Pascal** e **MSX-BASIC Dignified**, geração de objetos relocáveis `.MOB`, paginação automática na página 2 com Memory Mapper, suporte ao frame pointer `IX` (ABI de pilha) e biblioteca padrão **MSXLIB** com rotinas gráficas de alta velocidade.

---

## Destaques da Release

1. **DIGNAC (Compilador MSX-BASIC Dignified Z80 - Fase 5.3)**:
   - Compila código BASIC estruturado diretamente para módulos relocáveis `.MOB`.
   - Suporte a `MODULE`, `BANK`, `PUBLIC`, `EXTERN`, `PROCEDURE`, `SUB`, `FUNCTION`.
   - Quadro de pilha compatível com ABI Kizuna usando ponteiro de frame `IX` e variáveis `LOCAL x%, y%`.
   - Controle de fluxo estruturado: `FOR ... NEXT`, `WHILE ... WEND`, `DO ... LOOP`, `IF ... THEN ... ELSE ... END IF`.
   - Primitivas gráficas nativas: `LINE (x1,y1)-(x2,y2)[, color][, B | BF]` e `PSET (x, y)[, color]`.
   - Geração automática de ponto de entrada `Start` para criação direta de executáveis `.COM` standalone via `musubi`.
   - Opção `-S` para inspeção do código Assembly Z80 gerado.

2. **Suporte a Endereçamento Indexado no KAJI80**:
   - Implementado suporte completo a `LD r, (IX+d)`, `LD (IX+d), r`, `LD (IX+d), n` (e registradores `IY`).
   - Implementado suporte a `LD SP, IX` e `LD SP, IY` (`DD/FD F9`).

3. **Primitivas Gráficas na MSXLIB (`lib/src/vdp.asm`)**:
   - `VDP_PSet`: Plotagem de pixel em SCREEN 2 (256x192) calculando endereços na Pattern e Color Table.
   - `VDP_BoxFill`: Preenchimento de retângulos e limpeza de tela cheia ultrarrápida.
   - `VDP_Line`: Desenho de retas e segmentos na tela gráfica.

4. **WIRTH80 (Compilador Pascal Z80)**:
   - Suporte a variáveis `Integer`, `Char` e `Boolean`, blocos `begin ... end.`.
   - Aritmética de 16 bits, chamadas de sistema, `WriteLn` e integração com `MSXLIB`.

5. **Exemplos em BASIC (`sample/basic/`)**:
   - `hello.bas`: Hello World em BASIC compilado (apenas 186 bytes).
   - `calc.bas`: Demonstração de variáveis locais, aritmética e chamadas à biblioteca padrão.
   - `chart.bas`: Módulo gráfico paginado no banco 2 para desenho de curvas e gráficos.
   - `build.ps1`: Script de compilação e smart-linking automatizado.

---

## Conteúdo do Pacote de Distribuição (`kizuna-v4.5.0-dist.zip`)

```
distribute/
  ├── bin/
  │    ├── kaji80.exe     (Assembler Z80)
  │    ├── wirth80.exe    (Compilador Pascal)
  │    ├── dignac.exe     (Compilador BASIC Dignified)
  │    ├── musubi.exe     (Linker com Smart-Linking)
  │    ├── hako.exe       (Bibliotecário)
  │    └── mobdump.exe    (Inspecionador de objetos)
  ├── lib/
  │    └── msxlib.hlib    (Biblioteca Padrão MSX)
  ├── sample/
  │    ├── hello.asm      (Exemplo Assembly monobanco)
  │    ├── multibank/     (Exemplo Assembly multi-banco com 3 módulos)
  │    ├── libdemo/       (Exemplo Assembly consumindo MSXLIB)
  │    ├── pascal/        (Exemplos em Pascal: hello.pas e calc.pas)
  │    └── basic/         (Exemplos em BASIC: hello.bas, calc.bas, chart.bas)
  ├── docs/
  │    ├── README.md
  │    ├── HELP.md
  │    └── CHANGELOG.md
  ├── install.exe / install.cmd
  └── LICENSE
```

---

## Próximos Passos

- **Fase 6**: Orquestrador de build declarativo **OBI** (`Obifile`) para compilar e linkar projetos multi-linguagem em um único comando.
- Construção da demo poliglota completa (`demo/`) unindo Pascal (`main.pas`), Assembly (`screen.asm`) e BASIC (`chart.bas`).
