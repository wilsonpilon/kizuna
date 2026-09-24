; =============================================================================
; KIZUNA MSXLIB - str/copy
; copia de string tamanho+dados
; =============================================================================

MODULE str_copy
BANK 0

PUBLIC STR_Copy

; STR_Copy: copia a string de HL (tamanho + dados) para DE. O destino precisa ter
; espaco para tamanho + 1 bytes (ate 256).
; Entrada: HL = origem, DE = destino
; Destrói: BC, DE, HL, flags.
STR_Copy:
    LD C, (HL)
    LD B, 00h
    INC BC              ; tamanho + o proprio byte de tamanho
    LDIR
    RET

ENDMOD
