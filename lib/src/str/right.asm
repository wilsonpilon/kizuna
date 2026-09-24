; =============================================================================
; KIZUNA MSXLIB - str/right
; RIGHT$: os N ultimos caracteres
; =============================================================================

MODULE str_right
BANK 0

PUBLIC STR_Right

; STR_Right: os A ultimos caracteres da string de HL (ou ela toda, se for mais curta)
; Entrada: HL = origem, A = quantidade, DE = destino (ate 256 bytes; nao pode ser a propria origem)
; Destrói: A, BC, DE, HL, flags.
STR_Right:
    LD C, (HL)          ; C = tamanho
    CP C
    JR C, STR_Right_Use
    LD A, C
STR_Right_Use:
    LD B, A             ; B = quantidade
    LD A, C
    SUB B               ; A = quantos pular do inicio
    INC HL              ; HL = dados
    ADD A, L
    LD L, A
    JR NC, STR_Right_Go
    INC H
STR_Right_Go:
    LD A, B
    LD (DE), A
    INC DE
    LD C, B
    LD B, 00h
    LD A, C
    OR A
    RET Z
    LDIR
    RET

ENDMOD
