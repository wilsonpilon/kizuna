; =============================================================================
; KIZUNA MSXLIB - cstr/reverse
; inverte a ordem dos caracteres de uma string
; =============================================================================

MODULE cstr_reverse
BANK 0

PUBLIC CSTR_Reverse

; CSTR_Reverse: inverte a string de HL no lugar
; Preserva: A, BC, DE, HL. Destrói: flags.
CSTR_Reverse:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD D, H
    LD E, L
    XOR A
CSTR_Reverse_Find:
    CP (HL)             ; HL avanca ate o terminador
    JR Z, CSTR_Reverse_End
    INC HL
    JR CSTR_Reverse_Find
CSTR_Reverse_End:
    EX DE, HL           ; HL = inicio, DE = terminador
    DEC DE              ; DE = ultimo caractere (inicio - 1 se vazia)
CSTR_Reverse_Swap:
    PUSH HL
    OR A
    SBC HL, DE
    POP HL
    JR NC, CSTR_Reverse_Done    ; inicio >= fim: acabou
    LD A, (HL)
    LD C, A
    LD A, (DE)
    LD (HL), A
    LD A, C
    LD (DE), A
    INC HL
    DEC DE
    JR CSTR_Reverse_Swap
CSTR_Reverse_Done:
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
