; =============================================================================
; KIZUNA MSXLIB - char/isspace
; espaco em branco: espaco, TAB, LF, VT, FF, CR
; =============================================================================

MODULE char_isspace
BANK 0

PUBLIC CHAR_IsSpace

; CHAR_IsSpace: o caractere em A e espaco em branco: espaco, TAB, LF, VT, FF, CR?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsSpace:
    CP 09h
    JR C, CHAR_IsSpace_No
    CP 0Eh
    JR C, CHAR_IsSpace_Yes
    CP 20h
    JR C, CHAR_IsSpace_No
    CP 21h
    JR C, CHAR_IsSpace_Yes
CHAR_IsSpace_No:
    XOR A
    RET
CHAR_IsSpace_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
