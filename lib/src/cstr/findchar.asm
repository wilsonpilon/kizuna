; =============================================================================
; KIZUNA MSXLIB - cstr/findchar
; busca de um caractere numa string terminada em zero
; =============================================================================

MODULE cstr_findchar
BANK 0

PUBLIC CSTR_FindChar

; CSTR_FindChar: primeira ocorrencia do caractere A na string de HL. Procurar o
; proprio zero devolve o endereco do terminador (como o strchr do C).
; Entrada: HL = string, A = caractere
; Saída: HL = endereco da ocorrencia, ou 0 se nao ha; flag Z = 1 se achou
; Preserva: BC. Destrói: A, flags.
CSTR_FindChar:
    PUSH BC
    LD B, A
CSTR_FindChar_Loop:
    LD A, (HL)
    CP B
    JR Z, CSTR_FindChar_Found
    OR A
    JR Z, CSTR_FindChar_None
    INC HL
    JR CSTR_FindChar_Loop
CSTR_FindChar_Found:
    POP BC
    RET                 ; Z ja esta em 1
CSTR_FindChar_None:
    LD HL, 0000h
    LD A, 01h
    OR A                ; Z = 0
    POP BC
    RET

ENDMOD
