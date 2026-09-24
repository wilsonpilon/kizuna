; =============================================================================
; KIZUNA MSXLIB - mem/heapinit
; heap dinamico: estado e inicializacao
; =============================================================================

MODULE mem_heapinit
BANK 0

PUBLIC MEM_HeapInit

; O heap e uma lista de blocos contiguos. Cada bloco comeca com um cabecalho de
; 16 bits: o TAMANHO do conteudo (sempre par) com o bit 0 = 1 quando o bloco esta
; EM USO. O conteudo vem logo depois do cabecalho. A lista termina numa
; SENTINELA: um cabecalho 0001h (tamanho 0, em uso) nos dois ultimos bytes do heap.
; Blocos livres vizinhos sao fundidos aos poucos (MEM_Alloc e MEM_HeapCompact).
;
; Estado (uma unica instancia de heap por programa; as rotinas nao sao
; reentrantes -- nao chame do tratador de interrupcao):
MEM_HeapStart:
    DW 0000h            ; endereco do primeiro bloco (0 = heap nao iniciado)
MEM_HeapEnd:
    DW 0000h            ; endereco logo depois da sentinela

; MEM_HeapInit: define a regiao do heap
; Entrada: HL = inicio, BC = tamanho em bytes. Precisa de ao menos 6 bytes; o
;          inicio e arredondado para par e o tamanho tambem.
; Saída: A = 0 se deu certo, 1 se a regiao e pequena/invalida (o heap fica nao iniciado)
; Destrói: BC, HL, flags.
MEM_HeapInit:
    LD A, B
    OR C
    JR Z, MEM_HeapInit_Err
    BIT 0, L
    JR Z, MEM_HeapInit_Aligned
    INC HL
    DEC BC
MEM_HeapInit_Aligned:
    RES 0, C
    LD A, B
    OR A
    JR NZ, MEM_HeapInit_SizeOk
    LD A, C
    CP 06h
    JR C, MEM_HeapInit_Err
MEM_HeapInit_SizeOk:
    PUSH HL
    ADD HL, BC          ; fim da regiao
    JR C, MEM_HeapInit_ErrPop
    LD (MEM_HeapEnd), HL
    DEC HL
    DEC HL
    LD (HL), 01h        ; sentinela 0001h
    INC HL
    LD (HL), 00h
    POP HL
    LD (MEM_HeapStart), HL
    DEC BC
    DEC BC
    DEC BC
    DEC BC              ; conteudo do bloco livre = tamanho - cabecalho - sentinela
    LD (HL), C
    INC HL
    LD (HL), B
    XOR A
    RET
MEM_HeapInit_ErrPop:
    POP HL
MEM_HeapInit_Err:
    LD HL, 0000h
    LD (MEM_HeapStart), HL
    LD (MEM_HeapEnd), HL
    LD A, 01h
    RET

ENDMOD
