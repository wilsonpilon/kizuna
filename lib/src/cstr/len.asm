; =============================================================================
; KIZUNA MSXLIB - cstr/len
; comprimento de uma string terminada em zero
; =============================================================================

MODULE cstr_len
BANK 0

PUBLIC CSTR_Len

; CSTR_Len: numero de caracteres antes do terminador zero
; Entrada: HL = string
; Saída: HL = comprimento (0..65535)
; Preserva: BC, DE. Destrói: A, flags.
CSTR_Len:
    PUSH DE
    LD D, H
    LD E, L
    XOR A
CSTR_Len_Loop:
    CP (HL)
    JR Z, CSTR_Len_Done
    INC HL
    JR CSTR_Len_Loop
CSTR_Len_Done:
    OR A
    SBC HL, DE
    POP DE
    RET

ENDMOD
