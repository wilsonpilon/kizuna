; =============================================================================
; KIZUNA MSXLIB - cstr/replace
; troca um caractere por outro numa string
; =============================================================================

MODULE cstr_replace
BANK 0

PUBLIC CSTR_ReplaceChar

; CSTR_ReplaceChar: troca toda ocorrencia do caractere D pelo caractere E na
; string de HL, no lugar
; Entrada: HL = string, D = caractere a trocar (nao zero), E = caractere novo
; Preserva: A, BC, DE, HL. Destrói: flags.
CSTR_ReplaceChar:
    PUSH AF
    PUSH HL
CSTR_ReplaceChar_Loop:
    LD A, (HL)
    OR A
    JR Z, CSTR_ReplaceChar_Done
    CP D
    JR NZ, CSTR_ReplaceChar_Next
    LD (HL), E
CSTR_ReplaceChar_Next:
    INC HL
    JR CSTR_ReplaceChar_Loop
CSTR_ReplaceChar_Done:
    POP HL
    POP AF
    RET

ENDMOD
