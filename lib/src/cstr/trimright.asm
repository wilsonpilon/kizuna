; =============================================================================
; KIZUNA MSXLIB - cstr/trimright
; remove os espacos do fim de uma string
; =============================================================================

MODULE cstr_trimright
BANK 0

PUBLIC CSTR_TrimRight
EXTERN CHAR_IsSpace

; CSTR_TrimRight: tira os espacos em branco do FIM da string de HL, no lugar
; Preserva: A, BC, DE, HL. Destrói: flags.
CSTR_TrimRight:
    PUSH AF
    PUSH DE
    PUSH HL
    LD D, H
    LD E, L             ; DE = inicio
CSTR_TrimRight_Find:
    LD A, (HL)
    OR A
    JR Z, CSTR_TrimRight_Back
    INC HL
    JR CSTR_TrimRight_Find
CSTR_TrimRight_Back:            ; HL = terminador
    PUSH HL
    OR A
    SBC HL, DE
    POP HL
    JR Z, CSTR_TrimRight_Term   ; chegou ao inicio: ficou vazia
    DEC HL
    LD A, (HL)
    CALL CHAR_IsSpace
    JR NZ, CSTR_TrimRight_Back  ; era espaco: continua voltando
    INC HL                      ; o terminador vai logo depois do ultimo nao-espaco
CSTR_TrimRight_Term:
    XOR A
    LD (HL), A
    POP HL
    POP DE
    POP AF
    RET

ENDMOD
