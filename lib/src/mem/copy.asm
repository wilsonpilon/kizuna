; =============================================================================
; KIZUNA MSXLIB - mem/copy
; copia de bloco que trata sobreposicao (como memmove)
; =============================================================================

MODULE mem_copy
BANK 0

PUBLIC MEM_Copy

; MEM_Copy: copia BC bytes de HL para DE. Funciona mesmo com as regioes
; sobrepostas (escolhe copiar do inicio ou do fim, como memmove).
; Entrada: HL = origem, DE = destino, BC = quantidade (0 nao faz nada)
; Destrói: A, BC, DE, HL, flags.
MEM_Copy:
    LD A, B
    OR C
    RET Z
    PUSH HL
    OR A
    SBC HL, DE          ; HL = origem - destino (carry se origem < destino)
    POP HL
    JR NC, MEM_Copy_Fwd ; origem >= destino: copiar do inicio e sempre seguro
    PUSH HL
    ADD HL, BC          ; origem + quantidade
    JR C, MEM_Copy_Back ; passou de 64K: com certeza invade o destino
    OR A
    SBC HL, DE          ; (origem + quantidade) - destino
    JR C, MEM_Copy_FwdPop
    JR Z, MEM_Copy_FwdPop   ; termina antes (ou junto) do destino: sem sobreposicao
MEM_Copy_Back:
    POP HL
    ADD HL, BC
    DEC HL              ; HL = ultimo byte da origem
    EX DE, HL
    ADD HL, BC
    DEC HL              ; ultimo byte do destino
    EX DE, HL
    LDDR
    RET
MEM_Copy_FwdPop:
    POP HL
MEM_Copy_Fwd:
    LDIR
    RET

ENDMOD
