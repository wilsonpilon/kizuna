; =============================================================================
; KIZUNA MSXLIB - char/isprint
; imprimivel, incluindo o espaco: 20h..7Eh
; =============================================================================

MODULE char_isprint
BANK 0

PUBLIC CHAR_IsPrint

; CHAR_IsPrint: o caractere em A e imprimivel, incluindo o espaco: 20h..7Eh?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsPrint:
    CP 20h
    JR C, CHAR_IsPrint_No
    CP 7Fh
    JR C, CHAR_IsPrint_Yes
CHAR_IsPrint_No:
    XOR A
    RET
CHAR_IsPrint_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
