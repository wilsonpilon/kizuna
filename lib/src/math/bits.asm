; =============================================================================
; KIZUNA MSXLIB - math/bits
; inversao da ordem dos bits e troca dos bytes de uma palavra
; =============================================================================

MODULE math_bits
BANK 0

PUBLIC MATH_Flip8, MATH_Flip16, MATH_Swap16

; MATH_Flip8: inverte a ordem dos bits de A (bit 7 <-> bit 0, 6 <-> 1, ...)
; Entrada/Saída: A
; Preserva: BC, DE, HL. Destrói: flags.
MATH_Flip8:
    PUSH BC
    LD C, A
    LD B, 08h
MATH_Flip8_Loop:
    RR C                ; bit baixo de C vai para o carry ...
    RLA                 ; ... e entra pela direita em A (A vai ficando invertido)
    DJNZ MATH_Flip8_Loop
    POP BC
    RET

; MATH_Flip16: inverte a ordem dos 16 bits de HL
; Entrada/Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MATH_Flip16:
    PUSH AF
    LD A, L
    CALL MATH_Flip8
    LD L, A
    LD A, H
    CALL MATH_Flip8
    LD H, A
    LD A, L             ; troca H e L: o bit 15 novo vem do bit 0 antigo
    LD L, H
    LD H, A
    POP AF
    RET

; MATH_Swap16: troca os bytes de HL (H <-> L)
; Entrada/Saída: HL
; Preserva: A, BC, DE. Destrói: nada (nem flags).
MATH_Swap16:
    PUSH AF
    LD A, H
    LD H, L
    LD L, A
    POP AF
    RET

ENDMOD
