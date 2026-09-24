; =============================================================================
; KIZUNA MSXLIB - cstr/comparen
; comparacao limitada de strings terminadas em zero
; =============================================================================

MODULE cstr_comparen
BANK 0

PUBLIC CSTR_CompareN

; CSTR_CompareN: como CSTR_Compare, olhando no maximo BC caracteres
; Entrada: HL = a, DE = b, BC = maximo de caracteres (0 = iguais)
; Saída: A = 0 iguais, 1 se a > b, 0FFh se a < b; flag Z = 1 quando iguais
; Destrói: BC, DE, HL, flags.
CSTR_CompareN:
    LD A, B
    OR C
    JR Z, CSTR_CompareN_Eq
    LD A, (DE)
    CP (HL)
    JR NZ, CSTR_CompareN_Diff
    OR A
    JR Z, CSTR_CompareN_Eq
    INC HL
    INC DE
    DEC BC
    JR CSTR_CompareN
CSTR_CompareN_Eq:
    XOR A
    RET
CSTR_CompareN_Diff:
    JR C, CSTR_CompareN_Gt
    LD A, 0FFh
    RET
CSTR_CompareN_Gt:
    LD A, 01h
    RET

ENDMOD
