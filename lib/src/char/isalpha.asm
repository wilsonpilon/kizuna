; =============================================================================
; KIZUNA MSXLIB - char/isalpha
; letra 'A'..'Z' ou 'a'..'z'
; =============================================================================

MODULE char_isalpha
BANK 0

PUBLIC CHAR_IsAlpha

; CHAR_IsAlpha: o caractere em A e letra 'A'..'Z' ou 'a'..'z'?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsAlpha:
    CP 41h
    JR C, CHAR_IsAlpha_No
    CP 5Bh
    JR C, CHAR_IsAlpha_Yes
    CP 61h
    JR C, CHAR_IsAlpha_No
    CP 7Bh
    JR C, CHAR_IsAlpha_Yes
CHAR_IsAlpha_No:
    XOR A
    RET
CHAR_IsAlpha_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
