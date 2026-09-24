; =============================================================================
; KIZUNA MSXLIB - char/isalnum
; letra ou digito
; =============================================================================

MODULE char_isalnum
BANK 0

PUBLIC CHAR_IsAlNum

; CHAR_IsAlNum: o caractere em A e letra ou digito?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsAlNum:
    CP 30h
    JR C, CHAR_IsAlNum_No
    CP 3Ah
    JR C, CHAR_IsAlNum_Yes
    CP 41h
    JR C, CHAR_IsAlNum_No
    CP 5Bh
    JR C, CHAR_IsAlNum_Yes
    CP 61h
    JR C, CHAR_IsAlNum_No
    CP 7Bh
    JR C, CHAR_IsAlNum_Yes
CHAR_IsAlNum_No:
    XOR A
    RET
CHAR_IsAlNum_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
