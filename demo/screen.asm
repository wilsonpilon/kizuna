; ============================================================
; KIZUNA demo -- croqui hipotetico (nao monta / prova de conceito)
; Compilador: KAJI80
; Banco: 1 (paginavel)
; Tamanho estimado: ~12Kb (inclui outras rotinas alem de Setup)
; ============================================================

        MODULE  SCREENFX
        BANK    1

        PUBLIC  Setup
        PUBLIC  MostrarAbertura
        EXTERN  BIOS_CHGMOD, BIOS_CHGCLR, BIOS_WIDTH, BIOS_KEYOFF
        EXTERN  RES_Opening          ; resource de 12Kb, banco 3

; --------------------------------------------------------------
; Setup: SCREEN 2, COLOR 15,1,1, WIDTH 32, KEY OFF
; --------------------------------------------------------------
Setup:
        ld      a, 2
        call    BIOS_CHGMOD         ; SCREEN 2
        ld      a, 15
        ld      b, 1
        ld      c, 1
        call    BIOS_CHGCLR         ; COLOR 15,1,1
        ld      a, 32
        call    BIOS_WIDTH          ; WIDTH 32
        call    BIOS_KEYOFF         ; KEY OFF
        ret

; --------------------------------------------------------------
; MostrarAbertura: troca para o banco do resource, copia para VRAM,
; volta ao banco anterior. Chamado direto do Pascal (mesmo banco
; nao -- Pascal esta no banco 0, isto aqui e banco 1: MUSUBI ja
; gera o trampolim sozinho, o codigo abaixo nao precisa se
; preocupar com isso).
; --------------------------------------------------------------
MostrarAbertura:
        push    af
        in      a, (MAPPER_PAGE2)
        push    af
        ld      a, 3                ; banco do resource Opening
        out     (MAPPER_PAGE2), a
        ld      hl, RES_Opening
        ; ... copia 12Kb para VRAM (detalhe omitido no croqui) ...
        pop     af
        out     (MAPPER_PAGE2), a
        pop     af
        ret

        ; ... demais rotinas do modulo, completando os ~12Kb do banco ...

        ENDMOD
