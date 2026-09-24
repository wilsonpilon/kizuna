; =============================================================================
; KIZUNA MSXLIB - mem/find
; busca de um byte num bloco de memoria
; =============================================================================

MODULE mem_find
BANK 0

PUBLIC MEM_Find

; MEM_Find: procura o valor A nos BC primeiros bytes a partir de HL
; Entrada: HL = inicio, BC = quantidade (0 = nao procura), A = valor
; Saída: HL = endereco da primeira ocorrencia, ou 0 se nao achou;
;        flag Z = 1 se achou
; Destrói: A, BC, flags.
MEM_Find:
    PUSH AF
    LD A, B
    OR C
    JR Z, MEM_Find_None
    POP AF
    CPIR
    JR NZ, MEM_Find_Miss
    DEC HL              ; o CPIR para um byte depois do achado
    RET                 ; Z ja esta em 1
MEM_Find_None:
    POP AF
MEM_Find_Miss:
    LD HL, 0000h
    LD A, 01h
    OR A                ; Z = 0
    RET

ENDMOD
