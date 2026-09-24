; =============================================================================
; KIZUNA MSXLIB - char/ispunct
; pontuacao: imprimivel que nao e letra, digito nem espaco
; =============================================================================

MODULE char_ispunct
BANK 0

PUBLIC CHAR_IsPunct

; CHAR_IsPunct: o caractere em A e pontuacao: imprimivel que nao e letra, digito nem espaco?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsPunct:
    CP 21h
    JR C, CHAR_IsPunct_No
    CP 30h
    JR C, CHAR_IsPunct_Yes
    CP 3Ah
    JR C, CHAR_IsPunct_No
    CP 41h
    JR C, CHAR_IsPunct_Yes
    CP 5Bh
    JR C, CHAR_IsPunct_No
    CP 61h
    JR C, CHAR_IsPunct_Yes
    CP 7Bh
    JR C, CHAR_IsPunct_No
    CP 7Fh
    JR C, CHAR_IsPunct_Yes
CHAR_IsPunct_No:
    XOR A
    RET
CHAR_IsPunct_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
