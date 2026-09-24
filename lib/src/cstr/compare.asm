; =============================================================================
; KIZUNA MSXLIB - cstr/compare
; comparacao de strings terminadas em zero
; =============================================================================

MODULE cstr_compare
BANK 0

PUBLIC CSTR_Compare

; CSTR_Compare: compara a string de HL (a) com a de DE (b), caractere a
; caractere, como bytes SEM sinal (a mais curta e "menor" que a que a continua).
; Saída: A = 0 se iguais, 1 se a > b, 0FFh se a < b; flag Z = 1 quando iguais
; Preserva: BC. Destrói: DE, HL, flags.
CSTR_Compare:
    LD A, (DE)
    CP (HL)             ; b - a
    JR NZ, CSTR_Compare_Diff
    OR A
    JR Z, CSTR_Compare_Eq   ; os dois terminaram juntos
    INC HL
    INC DE
    JR CSTR_Compare
CSTR_Compare_Eq:
    XOR A
    RET
CSTR_Compare_Diff:
    JR C, CSTR_Compare_Gt
    LD A, 0FFh
    RET
CSTR_Compare_Gt:
    LD A, 01h
    RET

ENDMOD
