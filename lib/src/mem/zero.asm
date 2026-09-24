; =============================================================================
; KIZUNA MSXLIB - mem/zero
; zera um bloco de memoria
; =============================================================================

MODULE mem_zero
BANK 0

PUBLIC MEM_Zero
EXTERN MEM_Fill

; MEM_Zero: zera BC bytes a partir de HL
; Entrada: HL = destino, BC = quantidade (0 nao faz nada)
; Preserva: DE. Destrói: A, BC, HL, flags.
MEM_Zero:
    XOR A
    JP MEM_Fill

ENDMOD
