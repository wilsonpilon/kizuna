; =============================================================================
; KIZUNA MSXLIB - mem/alloc
; heap dinamico: alocacao
; =============================================================================

MODULE mem_alloc
BANK 0

PUBLIC MEM_Alloc
EXTERN MEM_HeapStart

; MEM_Alloc: reserva um bloco de memoria (primeiro que couber, "first fit")
; Entrada: HL = tamanho pedido em bytes (arredondado para par; 0 falha)
; Saída: HL = endereco do bloco, ou 0 se nao ha espaco / heap nao iniciado
; Preserva: BC, DE. Destrói: A, flags.
; Enquanto percorre a lista, funde blocos livres vizinhos. Sobra do bloco
; escolhido vira um novo bloco livre quando tem ao menos 4 bytes (cabecalho + 2).
MEM_Alloc:
    PUSH BC
    PUSH DE
    LD A, H
    OR L
    JP Z, MEM_Alloc_Fail
    INC HL
    LD A, H
    OR L
    JP Z, MEM_Alloc_Fail    ; pedido de 65535 bytes: nao existe bloco assim
    RES 0, L                ; HL = pedido arredondado para par
    LD B, H
    LD C, L                 ; BC = pedido
    LD HL, (MEM_HeapStart)
    LD A, H
    OR L
    JP Z, MEM_Alloc_Fail    ; heap nao iniciado
MEM_Alloc_Scan:
    LD E, (HL)
    INC HL
    LD D, (HL)
    DEC HL                  ; DE = cabecalho do bloco em HL
    BIT 0, E
    JP NZ, MEM_Alloc_Used
MEM_Alloc_Coal:             ; bloco livre: funde com os livres que vem logo depois
    PUSH HL
    ADD HL, DE
    INC HL
    INC HL                  ; HL = cabecalho do proximo bloco
    BIT 0, (HL)
    JR NZ, MEM_Alloc_CoalEnd
    LD A, (HL)
    INC HL
    LD H, (HL)
    LD L, A                 ; HL = tamanho do proximo
    ADD HL, DE
    INC HL
    INC HL                  ; tamanho fundido
    EX DE, HL
    POP HL                  ; HL = este bloco
    LD (HL), E
    INC HL
    LD (HL), D
    DEC HL
    JR MEM_Alloc_Coal
MEM_Alloc_CoalEnd:
    POP HL                  ; HL = este bloco, DE = tamanho (livre)
    PUSH HL
    LD H, D
    LD L, E
    OR A
    SBC HL, BC              ; HL = tamanho - pedido
    JR C, MEM_Alloc_NoFit
    LD A, H
    OR A
    JR NZ, MEM_Alloc_Split
    LD A, L
    CP 04h
    JR C, MEM_Alloc_Whole   ; resto pequeno demais para um bloco: entrega tudo
MEM_Alloc_Split:
    DEC HL
    DEC HL                  ; HL = tamanho do novo bloco livre (resto - cabecalho)
    POP DE                  ; DE = este bloco
    PUSH DE                 ; (guarda para o retorno)
    EX DE, HL               ; DE = tamanho livre novo, HL = este bloco
    INC HL
    INC HL                  ; HL = conteudo
    ADD HL, BC              ; HL = onde comeca o novo bloco livre
    LD (HL), E
    INC HL
    LD (HL), D              ; cabecalho do bloco livre da sobra
    POP HL                  ; HL = este bloco
    LD D, B
    LD E, C
    INC DE                  ; pedido | 1 (em uso)
    LD (HL), E
    INC HL
    LD (HL), D
    INC HL                  ; HL = conteudo
    JR MEM_Alloc_Done
MEM_Alloc_Whole:
    POP HL                  ; HL = este bloco; DE = tamanho
    INC DE                  ; tamanho | 1
    LD (HL), E
    INC HL
    LD (HL), D
    INC HL
    JR MEM_Alloc_Done
MEM_Alloc_NoFit:
    POP HL                  ; HL = este bloco; DE = tamanho
MEM_Alloc_Next:
    ADD HL, DE
    INC HL
    INC HL                  ; proximo bloco
    JP MEM_Alloc_Scan
MEM_Alloc_Used:
    RES 0, E                ; DE = tamanho do bloco em uso
    LD A, D
    OR E
    JR Z, MEM_Alloc_Fail    ; tamanho 0 em uso = sentinela: fim da lista
    JR MEM_Alloc_Next
MEM_Alloc_Fail:
    LD HL, 0000h
MEM_Alloc_Done:
    POP DE
    POP BC
    RET

ENDMOD
