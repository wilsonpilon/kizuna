; =============================================================================
; KIZUNA MSXLIB - cstr/trimleft
; remove os espacos do inicio de uma string
; =============================================================================

MODULE cstr_trimleft
BANK 0

PUBLIC CSTR_TrimLeft
EXTERN CSTR_Copy, CHAR_IsSpace

; CSTR_TrimLeft: tira os espacos em branco do INICIO da string de HL, no lugar
; Preserva: A, BC, DE, HL. Destrói: flags.
CSTR_TrimLeft:
    PUSH AF
    PUSH DE
    PUSH HL
    LD D, H
    LD E, L             ; DE = destino (o inicio)
CSTR_TrimLeft_Skip:
    LD A, (HL)
    OR A
    JR Z, CSTR_TrimLeft_Copy
    CALL CHAR_IsSpace
    JR Z, CSTR_TrimLeft_Copy    ; nao e espaco
    INC HL
    JR CSTR_TrimLeft_Skip
CSTR_TrimLeft_Copy:
    CALL CSTR_Copy
    POP HL
    POP DE
    POP AF
    RET

ENDMOD
