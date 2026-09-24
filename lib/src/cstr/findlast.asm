; =============================================================================
; KIZUNA MSXLIB - cstr/findlast
; busca da ultima ocorrencia de um caractere
; =============================================================================

MODULE cstr_findlast
BANK 0

PUBLIC CSTR_FindLastChar

; CSTR_FindLastChar: ULTIMA ocorrencia do caractere A na string de HL
; Entrada: HL = string, A = caractere
; Saída: HL = endereco da ocorrencia, ou 0 se nao ha; flag Z = 1 se achou
; Preserva: BC, DE. Destrói: A, flags.
CSTR_FindLastChar:
    PUSH BC
    PUSH DE
    LD B, A
    LD DE, 0000h        ; DE = ultima ocorrencia vista
CSTR_FindLastChar_Loop:
    LD A, (HL)
    CP B
    JR NZ, CSTR_FindLastChar_NoMatch
    LD D, H
    LD E, L
CSTR_FindLastChar_NoMatch:
    OR A
    JR Z, CSTR_FindLastChar_End
    INC HL
    JR CSTR_FindLastChar_Loop
CSTR_FindLastChar_End:
    LD H, D
    LD L, E
    LD A, D
    OR E
    JR Z, CSTR_FindLastChar_None
    XOR A               ; Z = 1 (achou)
    POP DE
    POP BC
    RET
CSTR_FindLastChar_None:
    LD A, 01h
    OR A                ; Z = 0
    POP DE
    POP BC
    RET

ENDMOD
