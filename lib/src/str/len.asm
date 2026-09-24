; =============================================================================
; KIZUNA MSXLIB - str/len
; comprimento de uma string tamanho+dados
; =============================================================================

MODULE str_len
BANK 0

PUBLIC STR_Len

; STR_Len: comprimento (o primeiro byte da string)
; Entrada: HL = string
; Saída: A = 0..255
; Preserva: BC, DE, HL. Destrói: nada alem de A.
STR_Len:
    LD A, (HL)
    RET

ENDMOD
