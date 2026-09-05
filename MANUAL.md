# KIZUNA — Manual do Usuário (v4.4.0 — Release Akatsuki)

> Este manual descreve o uso da toolchain **KIZUNA** para MSX2+ / MSX-DOS 2.
> As ferramentas `kaji80`, `wirth80`, `musubi`, `hako`, `mobdump` e a biblioteca `msxlib.hlib`
> estão totalmente implementadas e funcionais.

## 1. Visão geral do fluxo

```
fonte.asm  ──KAJI80──►  fonte.mob  ─┐
fonte.pas  ──WIRTH80─►  fonte.mob  ─┼──MUSUBI (+ msxlib.hlib)──►  programa.com
fonte.bas  ──DIGNAC──►  fonte.mob  ─┘
```

## 2. As Ferramentas

| Ferramenta | Comando | Status | O que faz |
|---|---|---|---|
| **`KAJI80`** | `kaji80 arq.asm -o arq.mob` | **Funcional** | Assembler Z80 modular com suporte a bancos e relocações. |
| **`WIRTH80`**| `wirth80 arq.pas -o arq.mob` | **Funcional** | Compilador Pascal nativo (TP4-like) gerando objetos `.mob`. |
| **`MUSUBI`** | `musubi *.mob *.hlib -o prog.com` | **Funcional** | Linker multi-banco com Smart-Linking e trampolins de mapper. |
| **`HAKO`**   | `hako -c lib.hlib *.mob` | **Funcional** | Bibliotecário / empacotador de arquivos estáticos `.HLIB`. |
| **`MOBDUMP`**| `mobdump arq.mob` | **Funcional** | Despejo legível de cabeçalhos, segmentos, símbolos e relocs. |
| **`MSXLIB`** | `msxlib.hlib` | **Funcional** | Biblioteca padrão MSX (BDOS, BIOS, VDP, PSG, String, Math). |
| **`DIGNAC`** | `dignac arq.bas -o arq.mob` | **Funcional** | Compilador MSX-BASIC Dignified para Z80 gerando objetos `.mob`. |
| **`OBI`**    | `obi build` | *Planejado* | Orquestrador declarativo de build via `Obifile`. |

## 3. Guia Rápido de Uso

### 3.1. Compilando Programa em Pascal:
```bash
# 1. Compila Pascal para objeto relocável .mob
wirth80 sample/pascal/hello.pas -o sample/pascal/hello.mob

# 2. Linka com a biblioteca padrão (elimina código não utilizado automaticamente)
musubi -v -m sample/pascal/hello.map -o sample/pascal/hello.com \
  sample/pascal/hello.mob lib/msxlib.hlib
```

### 3.2. Compilando Programa em Assembly Z80:
```bash
# 1. Monta o Assembly
kaji80 sample/hello.asm -o sample/hello.mob

# 2. Linka gerando o executável .COM
musubi -v -o sample/hello.com sample/hello.mob
```

### 3.3. Compilando Programa em MSX-BASIC Dignified:
```bash
# 1. Compila BASIC Dignified para objeto relocável .mob
dignac sample/basic/hello.bas -o sample/basic/hello.mob

# 2. Linka com a biblioteca padrão gerando o executável .COM
musubi -v -o sample/basic/hello.com sample/basic/hello.mob lib/msxlib.hlib
```

## 4. Interoperabilidade entre linguagens

- Um módulo Pascal declara uma dependência externa com `{$USES Nome in
  'arquivo'}`, apontando para um módulo Assembly ou BASIC Dignified.
- Símbolos exportados (`PUBLIC` no Assembly, `interface` no Pascal,
  `PUBLIC PROCEDURE` no BASIC Dignified) ficam visíveis para os demais
  módulos linkados.
- Chamadas entre módulos em bancos diferentes são resolvidas
  automaticamente por `MUSUBI` via trampolim (transparente para quem
  escreve o código-fonte).
- Strings passadas entre linguagens seguem sempre o formato short string
  (1 byte de tamanho + dados) — não há conversão necessária nas
  fronteiras.

## 5. Exemplo mínimo

Veja o diretório `demo/` para um croqui completo: um programa Pascal que
chama uma rotina de ajuste de tela em Assembly, lê uma entrada do
usuário, chama um módulo BASIC Dignified para desenhar um gráfico, e
retorna ao Pascal — com os três módulos em bancos de memória diferentes.

## 6. Ver também

- `SPEC.md` — especificação técnica completa (formato `.MOB`, ABI,
  modelo de bancos de memória).
