; =============================================================================
; KIZUNA MSXLIB - mem/swap
; troca o conteudo de dois blocos de memoria
; =============================================================================

MODULE mem_swap
BANK 0

PUBLIC MEM_Swap

; MEM_Swap: troca BC bytes entre o bloco de HL e o de DE. As regioes NAO podem
; se sobrepor.
; Entrada: HL = bloco a, DE = bloco b, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, DE, HL, flags.
MEM_Swap:
    LD A, B
    OR C
    RET Z
MEM_Swap_Loop:
    LD A, (DE)
    PUSH AF
    LD A, (HL)
    LD (DE), A
    POP AF
    LD (HL), A
    INC HL
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, MEM_Swap_Loop
    RET

ENDMOD
