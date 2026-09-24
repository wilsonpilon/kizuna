; =============================================================================
; KIZUNA MSXLIB - str/chr
; CHR$: uma string de um caractere
; =============================================================================

MODULE str_chr
BANK 0

PUBLIC STR_Chr

; STR_Chr: monta em DE uma string de 1 caractere (o codigo A)
; Entrada: A = caractere, DE = destino (2 bytes)
; Preserva: A, BC, DE, HL. Destrói: flags.
STR_Chr:
    PUSH AF
    LD A, 01h
    LD (DE), A
    INC DE
    POP AF
    LD (DE), A
    DEC DE
    RET

ENDMOD
