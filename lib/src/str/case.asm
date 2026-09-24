; =============================================================================
; KIZUNA MSXLIB - str/case
; maiusculas/minusculas numa string tamanho+dados
; =============================================================================

MODULE str_case
BANK 0

PUBLIC STR_ToUpper, STR_ToLower
EXTERN CHAR_ToUpper, CHAR_ToLower

; STR_ToUpper: passa a string de HL para maiusculas, no lugar
; Preserva: A, BC, DE, HL. Destrói: flags.
STR_ToUpper:
    PUSH AF
    PUSH BC
    PUSH HL
    LD B, (HL)
    INC HL
    LD A, B
    OR A
    JR Z, STR_ToUpper_Done
STR_ToUpper_Loop:
    LD A, (HL)
    CALL CHAR_ToUpper
    LD (HL), A
    INC HL
    DJNZ STR_ToUpper_Loop
STR_ToUpper_Done:
    POP HL
    POP BC
    POP AF
    RET

; STR_ToLower: passa a string de HL para minusculas, no lugar
; Preserva: A, BC, DE, HL. Destrói: flags.
STR_ToLower:
    PUSH AF
    PUSH BC
    PUSH HL
    LD B, (HL)
    INC HL
    LD A, B
    OR A
    JR Z, STR_ToLower_Done
STR_ToLower_Loop:
    LD A, (HL)
    CALL CHAR_ToLower
    LD (HL), A
    INC HL
    DJNZ STR_ToLower_Loop
STR_ToLower_Done:
    POP HL
    POP BC
    POP AF
    RET

ENDMOD
