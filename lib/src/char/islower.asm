; =============================================================================
; KIZUNA MSXLIB - char/islower
; letra minuscula 'a'..'z'
; =============================================================================

MODULE char_islower
BANK 0

PUBLIC CHAR_IsLower

; CHAR_IsLower: o caractere em A e letra minuscula 'a'..'z'?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsLower:
    CP 61h
    JR C, CHAR_IsLower_No
    CP 7Bh
    JR C, CHAR_IsLower_Yes
CHAR_IsLower_No:
    XOR A
    RET
CHAR_IsLower_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
