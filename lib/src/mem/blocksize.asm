; =============================================================================
; KIZUNA MSXLIB - mem/blocksize
; heap dinamico: tamanho de um bloco alocado
; =============================================================================

MODULE mem_blocksize
BANK 0

PUBLIC MEM_BlockSize

; MEM_BlockSize: tamanho util do bloco obtido com MEM_Alloc. Pode ser MAIOR que
; o pedido: o pedido e arredondado para par e, quando a sobra do bloco e pequena
; demais para virar outro bloco, ela fica junto.
; Entrada: HL = ponteiro devolvido por MEM_Alloc (nao e validado)
; Saída: HL = tamanho em bytes
; Preserva: A, BC, DE. Destrói: flags.
MEM_BlockSize:
    PUSH DE
    DEC HL
    LD D, (HL)
    DEC HL
    LD E, (HL)
    RES 0, E            ; tira o bit de "em uso"
    EX DE, HL
    POP DE
    RET

ENDMOD
