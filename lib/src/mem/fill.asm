; =============================================================================
; KIZUNA MSXLIB - mem/fill
; preenchimento de memoria com um byte
; =============================================================================

MODULE mem_fill
BANK 0

PUBLIC MEM_Fill

; MEM_Fill: preenche BC bytes a partir de HL com o valor A
; Entrada: HL = destino, BC = quantidade (0 nao faz nada), A = valor
; Preserva: DE. Destrói: A, BC, HL, flags.
MEM_Fill:
    PUSH AF
    LD A, B
    OR C
    JR NZ, MEM_Fill_Go
    POP AF
    RET
MEM_Fill_Go:
    POP AF
    PUSH DE
    LD (HL), A
    DEC BC
    LD A, B
    OR C
    JR Z, MEM_Fill_Done ; era 1 byte so
    LD D, H
    LD E, L
    INC DE
    LDIR                ; propaga o primeiro byte por todo o bloco
MEM_Fill_Done:
    POP DE
    RET

ENDMOD
