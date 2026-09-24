; =============================================================================
; KIZUNA MSXLIB - char/isdigit
; digito decimal '0'..'9'
; =============================================================================

MODULE char_isdigit
BANK 0

PUBLIC CHAR_IsDigit

; CHAR_IsDigit: o caractere em A e digito decimal '0'..'9'?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsDigit:
    CP 30h
    JR C, CHAR_IsDigit_No
    CP 3Ah
    JR C, CHAR_IsDigit_Yes
CHAR_IsDigit_No:
    XOR A
    RET
CHAR_IsDigit_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
