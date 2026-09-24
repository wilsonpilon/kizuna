; =============================================================================
; KIZUNA MSXLIB - str/left
; LEFT$: os N primeiros caracteres
; =============================================================================

MODULE str_left
BANK 0

PUBLIC STR_Left

; STR_Left: os A primeiros caracteres da string de HL (ou ela toda, se for mais curta)
; Entrada: HL = origem, A = quantidade, DE = destino (ate 256 bytes; nao pode ser a propria origem)
; Destrói: A, BC, DE, HL, flags.
STR_Left:
    LD C, (HL)
    CP C
    JR C, STR_Left_Use  ; pedido < tamanho: vale o pedido
    LD A, C
STR_Left_Use:
    LD (DE), A
    INC HL
    INC DE
    LD C, A
    LD B, 00h
    OR A
    RET Z
    LDIR
    RET

ENDMOD
