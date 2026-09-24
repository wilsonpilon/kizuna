; =============================================================================
; KIZUNA MSXLIB - char/ishexdigit
; digito hexadecimal '0'..'9', 'A'..'F', 'a'..'f'
; =============================================================================

MODULE char_ishexdigit
BANK 0

PUBLIC CHAR_IsHexDigit

; CHAR_IsHexDigit: o caractere em A e digito hexadecimal '0'..'9', 'A'..'F', 'a'..'f'?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsHexDigit:
    CP 30h
    JR C, CHAR_IsHexDigit_No
    CP 3Ah
    JR C, CHAR_IsHexDigit_Yes
    CP 41h
    JR C, CHAR_IsHexDigit_No
    CP 47h
    JR C, CHAR_IsHexDigit_Yes
    CP 61h
    JR C, CHAR_IsHexDigit_No
    CP 67h
    JR C, CHAR_IsHexDigit_Yes
CHAR_IsHexDigit_No:
    XOR A
    RET
CHAR_IsHexDigit_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
