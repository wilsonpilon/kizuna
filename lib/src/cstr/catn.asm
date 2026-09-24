; =============================================================================
; KIZUNA MSXLIB - cstr/catn
; concatenacao limitada de strings terminadas em zero
; =============================================================================

MODULE cstr_catn
BANK 0

PUBLIC CSTR_CatN
EXTERN CSTR_CopyN

; CSTR_CatN: anexa no maximo BC caracteres da string de HL ao fim da de DE e
; termina com zero.
; Entrada: HL = origem, DE = destino, BC = maximo de caracteres a anexar
; Saída: DE = endereco do novo terminador
; Destrói: A, BC, HL, flags.
CSTR_CatN:
    LD A, (DE)
    OR A
    JR Z, CSTR_CatN_Go
    INC DE
    JR CSTR_CatN
CSTR_CatN_Go:
    JP CSTR_CopyN

ENDMOD
