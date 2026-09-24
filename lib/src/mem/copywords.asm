; =============================================================================
; KIZUNA MSXLIB - mem/copywords
; copia de blocos contados em PALAVRAS de 16 bits
; =============================================================================

MODULE mem_copywords
BANK 0

PUBLIC MEM_CopyWords, MEM_CopyFastWords
EXTERN MEM_Copy, MEM_CopyFast

; MEM_CopyWords: como MEM_Copy, mas BC conta palavras de 16 bits (2 bytes cada).
; Suporta ate 32767 palavras.
; Entrada: HL = origem, DE = destino, BC = quantidade de PALAVRAS
; Destrói: A, BC, DE, HL, flags.
MEM_CopyWords:
    SLA C
    RL B
    JP MEM_Copy

; MEM_CopyFastWords: como MEM_CopyFast (sem checar sobreposicao), em palavras.
MEM_CopyFastWords:
    SLA C
    RL B
    JP MEM_CopyFast

ENDMOD
