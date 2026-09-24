; =============================================================================
; KIZUNA MSXLIB - mem/free
; heap dinamico: liberacao
; =============================================================================

MODULE mem_free
BANK 0

PUBLIC MEM_Free
EXTERN MEM_HeapStart, MEM_HeapEnd

; MEM_Free: devolve um bloco obtido com MEM_Alloc
; Entrada: HL = endereco devolvido por MEM_Alloc (0 e aceito e ignorado)
; Saída: A = 0 se liberou, 1 se o ponteiro e invalido (fora do heap, impar) ou o
;        bloco ja estava livre (liberacao dupla). Um ponteiro que aponte para o
;        MEIO de um bloco nao e detectado.
; Preserva: BC, DE, HL. Destrói: flags.
MEM_Free:
    PUSH HL
    PUSH DE
    LD A, H
    OR L
    JR Z, MEM_Free_Ok       ; free(0) nao faz nada
    BIT 0, L
    JR NZ, MEM_Free_Err
    PUSH HL
    LD HL, (MEM_HeapStart)
    INC HL
    INC HL
    EX DE, HL               ; DE = inicio + 2 (primeiro conteudo possivel)
    POP HL
    OR A
    SBC HL, DE
    JR C, MEM_Free_Err      ; abaixo do primeiro conteudo
    ADD HL, DE              ; HL = ponteiro de novo
    PUSH HL
    EX DE, HL               ; DE = ponteiro
    LD HL, (MEM_HeapEnd)
    OR A
    SBC HL, DE              ; fim - ponteiro
    POP HL
    JR C, MEM_Free_Err
    JR Z, MEM_Free_Err      ; no fim ou alem dele
    DEC HL
    DEC HL                  ; HL = cabecalho
    BIT 0, (HL)
    JR Z, MEM_Free_Err      ; ja livre
    RES 0, (HL)
MEM_Free_Ok:
    POP DE
    POP HL
    XOR A
    RET
MEM_Free_Err:
    POP DE
    POP HL
    LD A, 01h
    OR A
    RET

ENDMOD
