; =============================================================================
; KIZUNA MSXLIB - char/iscontrol
; caractere de controle 00h..1Fh e 7Fh
; =============================================================================

MODULE char_iscontrol
BANK 0

PUBLIC CHAR_IsControl

; CHAR_IsControl: o caractere em A e caractere de controle 00h..1Fh e 7Fh?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsControl:
    CP 20h
    JR C, CHAR_IsControl_Yes
    CP 7Fh
    JR C, CHAR_IsControl_No
    CP 80h
    JR C, CHAR_IsControl_Yes
CHAR_IsControl_No:
    XOR A
    RET
CHAR_IsControl_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
