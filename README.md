# KIZUNA (絆)

> "Um laço entre linguagens, amarrado num único `.COM`."

**KIZUNA** é uma proposta de toolchain para MSX2+ (Z80 / MSX-DOS 2 / 256Kb
ou mais com memory mapper) que permite escrever partes de um mesmo programa em
**Assembly Z80**, **Pascal** (estilo Turbo Pascal 4, com units) e
**MSX-BASIC Dignified**, compilar cada parte separadamente e linkar tudo
num único executável — inclusive distribuindo módulos por bancos de
memória diferentes, com troca de banco resolvida automaticamente pelo
linker.

Versão Atual: `v4.5.0` — Release **Hinode (日の出)**.

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
| `MSXLIB`  | Biblioteca padrão (BDOS/BIOS/VDP/PSG) | **Em depuração gráfica** (v4.5.1) |
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

### Estado atual da SCREEN 2

O teste Assembly [sample/screen2_test.asm](sample/screen2_test.asm) já comprovou
entrada em SCREEN 2, escrita direta na VRAM e retorno ao MSX-DOS. A renderização
de `PSET`, `LINE` e texto gráfico ainda está em depuração: permanecem artefatos
visuais durante a inicialização e o endereçamento das tabelas SCREEN 2.

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
