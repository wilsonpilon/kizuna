; =============================================================================
; KIZUNA MSXLIB - mem/heapstats
; heap dinamico: tamanho, espaco livre e maior bloco livre
; =============================================================================

MODULE mem_heapstats
BANK 0

PUBLIC MEM_HeapSize, MEM_HeapFree, MEM_HeapLargest
EXTERN MEM_HeapStart, MEM_HeapEnd, MEM_HeapCompact

; MEM_HeapSize: tamanho total da regiao do heap, em bytes (0 = nao iniciado)
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MEM_HeapSize:
    PUSH DE
    LD HL, (MEM_HeapStart)
    EX DE, HL
    LD HL, (MEM_HeapEnd)
    OR A
    SBC HL, DE
    POP DE
    RET

; MEM_HeapFree: soma do conteudo dos blocos livres (sem contar os cabecalhos)
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MEM_HeapFree:
    PUSH AF
    PUSH BC
    PUSH DE
    LD BC, 0000h            ; BC = total
    LD HL, (MEM_HeapStart)
    LD A, H
    OR L
    JR Z, MEM_HeapFree_Done
MEM_HeapFree_Scan:
    LD E, (HL)
    INC HL
    LD D, (HL)
    DEC HL
    BIT 0, E
    JR NZ, MEM_HeapFree_Used
    PUSH HL
    LD H, B
    LD L, C
    ADD HL, DE
    LD B, H
    LD C, L
    POP HL
    JR MEM_HeapFree_Next
MEM_HeapFree_Used:
    RES 0, E
    LD A, D
    OR E
    JR Z, MEM_HeapFree_Done
MEM_HeapFree_Next:
    ADD HL, DE
    INC HL
    INC HL
    JR MEM_HeapFree_Scan
MEM_HeapFree_Done:
    LD H, B
    LD L, C
    POP DE
    POP BC
    POP AF
    RET

; MEM_HeapLargest: o maior bloco que um MEM_Alloc conseguiria entregar agora
; (funde os livres vizinhos antes de medir)
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MEM_HeapLargest:
    CALL MEM_HeapCompact
    PUSH AF
    PUSH BC
    PUSH DE
    LD BC, 0000h            ; BC = maior visto
    LD HL, (MEM_HeapStart)
    LD A, H
    OR L
    JR Z, MEM_HeapLargest_Done
MEM_HeapLargest_Scan:
    LD E, (HL)
    INC HL
    LD D, (HL)
    DEC HL
    BIT 0, E
    JR NZ, MEM_HeapLargest_Used
    PUSH HL
    LD H, B
    LD L, C
    OR A
    SBC HL, DE              ; maior - este
    POP HL
    JR NC, MEM_HeapLargest_Next
    LD B, D
    LD C, E
    JR MEM_HeapLargest_Next
MEM_HeapLargest_Used:
    RES 0, E
    LD A, D
    OR E
    JR Z, MEM_HeapLargest_Done
MEM_HeapLargest_Next:
    ADD HL, DE
    INC HL
    INC HL
    JR MEM_HeapLargest_Scan
MEM_HeapLargest_Done:
    LD H, B
    LD L, C
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
