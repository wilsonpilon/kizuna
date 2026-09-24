; =============================================================================
; KIZUNA MSXLIB - cstr/cat
; concatenacao de strings terminadas em zero
; =============================================================================

MODULE cstr_cat
BANK 0

PUBLIC CSTR_Cat
EXTERN CSTR_Copy

; CSTR_Cat: anexa a string de HL ao fim da string de DE. O destino precisa ter
; espaco para os dois comprimentos + 1.
; Entrada: HL = origem (o que anexar), DE = destino
; Saída: DE = endereco do novo terminador
; Preserva: BC. Destrói: A, HL, flags.
CSTR_Cat:
    LD A, (DE)
    OR A
    JR Z, CSTR_Cat_Go
    INC DE
    JR CSTR_Cat
CSTR_Cat_Go:
    JP CSTR_Copy

ENDMOD
