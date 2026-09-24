; =============================================================================
; KIZUNA MSXLIB - cstr/case
; conversao de maiusculas/minusculas numa string
; =============================================================================

MODULE cstr_case
BANK 0

PUBLIC CSTR_ToUpper, CSTR_ToLower
EXTERN CHAR_ToUpper, CHAR_ToLower

; CSTR_ToUpper: passa a string de HL para maiusculas, no lugar
; Preserva: BC, DE, HL. Destrói: A, flags.
CSTR_ToUpper:
    PUSH HL
CSTR_ToUpper_Loop:
    LD A, (HL)
    OR A
    JR Z, CSTR_ToUpper_Done
    CALL CHAR_ToUpper
    LD (HL), A
    INC HL
    JR CSTR_ToUpper_Loop
CSTR_ToUpper_Done:
    POP HL
    RET

; CSTR_ToLower: passa a string de HL para minusculas, no lugar
; Preserva: BC, DE, HL. Destrói: A, flags.
CSTR_ToLower:
    PUSH HL
CSTR_ToLower_Loop:
    LD A, (HL)
    OR A
    JR Z, CSTR_ToLower_Done
    CALL CHAR_ToLower
    LD (HL), A
    INC HL
    JR CSTR_ToLower_Loop
CSTR_ToLower_Done:
    POP HL
    RET

ENDMOD
