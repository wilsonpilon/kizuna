; =============================================================================
; KIZUNA MSXLIB - cstr/copyn
; copia de string terminada em zero, limitada a N caracteres
; =============================================================================

MODULE cstr_copyn
BANK 0

PUBLIC CSTR_CopyN

; CSTR_CopyN: copia no maximo BC caracteres da string de HL para DE e SEMPRE
; termina o destino com zero (que precisa ter BC + 1 bytes).
; Entrada: HL = origem, DE = destino, BC = maximo de caracteres
; Saída: DE = endereco do terminador no destino
; Destrói: A, BC, HL, flags.
CSTR_CopyN:
    LD A, B
    OR C
    JR Z, CSTR_CopyN_Term
    LD A, (HL)
    OR A
    JR Z, CSTR_CopyN_Term
    LDI
    JR CSTR_CopyN
CSTR_CopyN_Term:
    XOR A
    LD (DE), A
    RET

ENDMOD
