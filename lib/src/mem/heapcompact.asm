; =============================================================================
; KIZUNA MSXLIB - mem/heapcompact
; heap dinamico: funde todos os blocos livres vizinhos
; =============================================================================

MODULE mem_heapcompact
BANK 0

PUBLIC MEM_HeapCompact
EXTERN MEM_HeapStart

; MEM_HeapCompact: percorre o heap todo fundindo blocos livres vizinhos (o
; MEM_Alloc so faz isso no trecho que percorre). Util antes de perguntar
; MEM_HeapLargest.
; Preserva: A, BC, DE, HL. Destrói: flags.
MEM_HeapCompact:
    PUSH AF
    PUSH HL
    PUSH DE
    LD HL, (MEM_HeapStart)
    LD A, H
    OR L
    JR Z, MEM_HeapCompact_Done
MEM_HeapCompact_Scan:
    LD E, (HL)
    INC HL
    LD D, (HL)
    DEC HL
    BIT 0, E
    JR NZ, MEM_HeapCompact_Used
MEM_HeapCompact_Coal:
    PUSH HL
    ADD HL, DE
    INC HL
    INC HL
    BIT 0, (HL)
    JR NZ, MEM_HeapCompact_CoalEnd
    LD A, (HL)
    INC HL
    LD H, (HL)
    LD L, A
    ADD HL, DE
    INC HL
    INC HL
    EX DE, HL
    POP HL
    LD (HL), E
    INC HL
    LD (HL), D
    DEC HL
    JR MEM_HeapCompact_Coal
MEM_HeapCompact_CoalEnd:
    POP HL
    JR MEM_HeapCompact_Next
MEM_HeapCompact_Used:
    RES 0, E
    LD A, D
    OR E
    JR Z, MEM_HeapCompact_Done
MEM_HeapCompact_Next:
    ADD HL, DE
    INC HL
    INC HL
    JR MEM_HeapCompact_Scan
MEM_HeapCompact_Done:
    POP DE
    POP HL
    POP AF
    RET

ENDMOD
