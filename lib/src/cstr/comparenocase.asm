; =============================================================================
; KIZUNA MSXLIB - cstr/comparenocase
; comparacao de strings ignorando maiusculas/minusculas
; =============================================================================

MODULE cstr_comparenocase
BANK 0

PUBLIC CSTR_CompareNoCase
EXTERN CHAR_ToUpper

; CSTR_CompareNoCase: como CSTR_Compare, mas 'a' e 'A' (e as demais letras) valem o mesmo
; Saída: A = 0 iguais, 1 se a > b, 0FFh se a < b; flag Z = 1 quando iguais
; Preserva: BC. Destrói: DE, HL, flags.
CSTR_CompareNoCase:
    PUSH BC
CSTR_CompareNoCase_Loop:
    LD A, (HL)
    CALL CHAR_ToUpper
    LD B, A
    LD A, (DE)
    CALL CHAR_ToUpper
    CP B                ; b - a
    JR NZ, CSTR_CompareNoCase_Diff
    OR A
    JR Z, CSTR_CompareNoCase_Eq
    INC HL
    INC DE
    JR CSTR_CompareNoCase_Loop
CSTR_CompareNoCase_Eq:
    XOR A
    POP BC
    RET
CSTR_CompareNoCase_Diff:
    JR C, CSTR_CompareNoCase_Gt
    LD A, 0FFh
    POP BC
    RET
CSTR_CompareNoCase_Gt:
    LD A, 01h
    POP BC
    RET

ENDMOD
