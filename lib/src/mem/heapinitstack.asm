; =============================================================================
; KIZUNA MSXLIB - mem/heapinitstack
; heap dinamico: inicializacao ate o topo da pilha
; =============================================================================

MODULE mem_heapinitstack
BANK 0

PUBLIC MEM_HeapInitToStack
EXTERN MEM_HeapInit

; MEM_HeapInitToStack: heap de HL ate (SP do chamador - margem). Bom para "use
; tudo que sobra da TPA, deixando espaco para a pilha crescer".
; Entrada: HL = inicio, BC = margem reservada para a pilha, em bytes
; Saída: A = 0 se deu certo, 1 se nao sobra espaco suficiente
; Destrói: BC, DE, HL, flags.
MEM_HeapInitToStack:
    PUSH HL
    LD HL, 0004h
    ADD HL, SP          ; SP do chamador depois do RET
    OR A
    SBC HL, BC          ; menos a margem
    JR C, MEM_HeapInitToStack_Err
    POP DE              ; DE = inicio
    OR A
    SBC HL, DE          ; HL = tamanho disponivel
    JR C, MEM_HeapInitToStack_Err2
    LD B, H
    LD C, L
    EX DE, HL           ; HL = inicio, BC = tamanho
    JP MEM_HeapInit
MEM_HeapInitToStack_Err:
    POP DE
MEM_HeapInitToStack_Err2:
    LD A, 01h
    RET

ENDMOD
