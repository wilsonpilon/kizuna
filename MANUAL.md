# KIZUNA — Manual do Usuário

> Este documento é um índice curto. A documentação completa está dividida
> por assunto em `docs/`:

| Manual | Cobre |
| ------- | ------ |
| [`docs/manual-assembly.md`](docs/manual-assembly.md) | Sintaxe do Assembly Z80 (`KAJI80`): diretivas, instruções suportadas, literais, linha de comando, recursos de macro-assembler (avaliador de expressões, rótulos locais, `IF`/`REPT`/`MACRO`, rótulos pré-definidos de BIOS/BDOS, `CALLBIOS`/`CALLDOS`, `INCBIN`). |
| [`docs/manual-basic-dignified.md`](docs/manual-basic-dignified.md) | Sintaxe do MSX-BASIC Dignified (`DIGNAC`): tipos, controle de fluxo, sprites, música, arquivos, linha de comando. |
| [`docs/manual-pascal.md`](docs/manual-pascal.md) | Sintaxe do Pascal (`WIRTH80`) — escopo real hoje vs. visão do projeto, linha de comando. |
| [`docs/manual-ferramentas.md`](docs/manual-ferramentas.md) | Formato `.MOB`/`.MAP`, o linker `MUSUBI`, bank switching, `HAKO`/`.HLIB`, `MSXLIB`, o orquestrador `OBI`, `MOBDUMP`, e um exemplo combinando as três linguagens num mesmo projeto. |
| [`SPEC.md`](SPEC.md) | Decisões de design originais e visão completa do projeto (inclui partes ainda não implementadas). |

## Visão geral do fluxo

```
fonte.asm  ──KAJI80──►  fonte.mob  ─┐
fonte.pas  ──WIRTH80─►  fonte.mob  ─┼──MUSUBI (+ msxlib.hlib)──►  programa.com
fonte.bas  ──DIGNAC──►  fonte.mob  ─┘
```

## As ferramentas

| Ferramenta | Comando | O que faz |
| ----------- | -------- | ----------- |
| **`KAJI80`** | `kaji80 arq.asm -o arq.mob` | Assembler Z80 modular com suporte a bancos e relocações. |
| **`WIRTH80`** | `wirth80 arq.pas -o arq.mob` | Compilador Pascal (subset atual: sem units/procedimentos — ver `docs/manual-pascal.md`). |
| **`DIGNAC`** | `dignac arq.bas -o arq.mob` | Compilador MSX-BASIC Dignified: tipos reais (STRING/INTEGER/SINGLE/DOUBLE), sprites, música, arquivo. |
| **`MUSUBI`** | `musubi *.mob *.hlib -o prog.com` | Linker multi-banco com Smart-Linking e trampolins de bank switching. |
| **`HAKO`** | `hako -c lib.hlib *.mob` | Bibliotecário/empacotador de objetos `.HLIB`. |
| **`MOBDUMP`** | `mobdump arq.mob` | Despejo legível de cabeçalhos, segmentos, símbolos e relocações. |
| **`OBI`** | `obi build [Obifile]` | Orquestrador declarativo de build — o "make" do projeto. |
| **`MSXLIB`** | `lib/msxlib.hlib` | Biblioteca padrão (BDOS, BIOS, VDP, PSG, String, Math). |

## Início rápido

```bash
# Assembly
kaji80 sample/hello.asm -o sample/hello.mob
musubi -v -o sample/hello.com sample/hello.mob

# BASIC Dignified
dignac sample/basic/hello.bas -o sample/basic/hello.mob
musubi -v -o sample/basic/hello.com sample/basic/hello.mob lib/msxlib.hlib

# Pascal
wirth80 sample/pascal/hello.pas -o sample/pascal/hello.mob
musubi -v -o sample/pascal/hello.com sample/pascal/hello.mob lib/msxlib.hlib

# Ou, para um projeto com vários módulos/bancos, via OBI:
obi build sample/obi/Obifile -v
```

Veja `docs/manual-ferramentas.md` §10 para o que "misturar as três
linguagens" realmente significa hoje (e o que ainda falta para um único
`.COM` cruzando as três).
