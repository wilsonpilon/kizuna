; =============================================================================
; KIZUNA MSXLIB - mem/fill16
; preenchimento de memoria com uma palavra de 16 bits
; =============================================================================

MODULE mem_fill16
BANK 0

PUBLIC MEM_Fill16

; MEM_Fill16: escreve BC palavras de 16 bits (byte baixo primeiro) a partir de HL
; Entrada: HL = destino, BC = numero de PALAVRAS (0 nao faz nada), DE = valor
; Destrói: A, BC, HL, flags.
MEM_Fill16:
    LD A, B
    OR C
    RET Z
MEM_Fill16_Loop:
    LD (HL), E
    INC HL
    LD (HL), D
    INC HL
    DEC BC
    LD A, B
    OR C
    JR NZ, MEM_Fill16_Loop
    RET

ENDMOD
