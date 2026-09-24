; =============================================================================
; KIZUNA MSXLIB - num/skipblanks
; pula espacos e TABs num texto
; =============================================================================

MODULE num_skipblanks
BANK 0

PUBLIC NUM_SkipBlanks

; NUM_SkipBlanks: avanca BC por cima de espacos e TABs. Para no fim do texto se
; DE (o ENDERECO DEPOIS do ultimo caractere; 0 = sem limite, o texto acaba num zero) for alcancado.
; Entrada: BC = ponteiro, DE = fim (0 = sem limite)
; Saída: BC avancado
; Preserva: DE, HL. Destrói: A, flags.
NUM_SkipBlanks:
    LD A, D
    OR E
    JR Z, NUM_SkipBlanks_Look
    LD A, B
    CP D
    JR NZ, NUM_SkipBlanks_Look
    LD A, C
    CP E
    RET Z               ; chegou ao fim
NUM_SkipBlanks_Look:
    LD A, (BC)
    CP 20h
    JR Z, NUM_SkipBlanks_Skip
    CP 09h
    RET NZ
NUM_SkipBlanks_Skip:
    INC BC
    JR NUM_SkipBlanks

ENDMOD
