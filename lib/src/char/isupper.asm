; =============================================================================
; KIZUNA MSXLIB - char/isupper
; letra maiuscula 'A'..'Z'
; =============================================================================

MODULE char_isupper
BANK 0

PUBLIC CHAR_IsUpper

; CHAR_IsUpper: o caractere em A e letra maiuscula 'A'..'Z'?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsUpper:
    CP 41h
    JR C, CHAR_IsUpper_No
    CP 5Bh
    JR C, CHAR_IsUpper_Yes
CHAR_IsUpper_No:
    XOR A
    RET
CHAR_IsUpper_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
