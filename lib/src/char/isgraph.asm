; =============================================================================
; KIZUNA MSXLIB - char/isgraph
; imprimivel sem o espaco: 21h..7Eh
; =============================================================================

MODULE char_isgraph
BANK 0

PUBLIC CHAR_IsGraph

; CHAR_IsGraph: o caractere em A e imprimivel sem o espaco: 21h..7Eh?
; Entrada: A = caractere (0..255)
; Saída: A = 1 (sim) ou 0 (nao); flag Z = 1 quando a resposta e nao
; Preserva: BC, DE, HL.
CHAR_IsGraph:
    CP 21h
    JR C, CHAR_IsGraph_No
    CP 7Fh
    JR C, CHAR_IsGraph_Yes
CHAR_IsGraph_No:
    XOR A
    RET
CHAR_IsGraph_Yes:
    LD A, 01h
    OR A
    RET

ENDMOD
