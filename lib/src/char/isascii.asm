; =============================================================================
; KIZUNA MSXLIB - char/isascii
; ASCII de 7 bits, 00h..7Fh
; =============================================================================

MODULE char_isascii
BANK 0

PUBLIC CHAR_IsAscii

; CHAR_IsAscii: o caractere em A e ASCII de 7 bits, 00h..7Fh?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsAscii:
    CP 80h
    JR C, CHAR_IsAscii_Yes
CHAR_IsAscii_No:
    XOR A
    RET
CHAR_IsAscii_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
