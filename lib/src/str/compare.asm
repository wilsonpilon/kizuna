; =============================================================================
; KIZUNA MSXLIB - str/compare
; comparacao de strings tamanho+dados
; =============================================================================

MODULE str_compare
BANK 0

PUBLIC STR_Compare

; STR_Compare: compara a string de HL (a) com a de DE (b), lexicograficamente,
; como bytes SEM sinal; se uma e prefixo da outra, a mais curta e a menor.
; Saída: A = 0 se iguais, 1 se a > b, 0FFh se a < b; flag Z = 1 quando iguais
; Destrói: BC, DE, HL, flags.
STR_Compare:
    LD B, (HL)          ; B = tamanho de a
    INC HL
    LD A, (DE)
    LD C, A             ; C = tamanho de b
    INC DE
STR_Compare_Loop:
    LD A, B
    OR A
    JR Z, STR_Compare_AEnd
    LD A, C
    OR A
    JR Z, STR_Compare_Longer    ; b acabou e a ainda tem caracteres
    LD A, (DE)
    CP (HL)             ; b - a
    JR NZ, STR_Compare_Diff
    INC HL
    INC DE
    DEC B
    DEC C
    JR STR_Compare_Loop
STR_Compare_AEnd:
    LD A, C
    OR A
    JR Z, STR_Compare_Eq
    LD A, 0FFh          ; a acabou primeiro: a < b
    OR A
    RET
STR_Compare_Eq:
    XOR A
    RET
STR_Compare_Longer:
    LD A, 01h
    OR A
    RET
STR_Compare_Diff:
    JR C, STR_Compare_Gt
    LD A, 0FFh
    RET
STR_Compare_Gt:
    LD A, 01h
    RET

ENDMOD
