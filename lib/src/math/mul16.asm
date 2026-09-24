; =============================================================================
; KIZUNA MSXLIB - math/mul16
; multiplicacao 16x16
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE math_mul16
BANK 0

PUBLIC Mul16

; -----------------------------------------------------------------------------
; Mul16: Multiplicação inteira não sinalizada de 16 bits
; Entrada: HL = multiplicando, DE = multiplicador
; Saída: HL = produto (16 bits mais baixos; igual para com e sem sinal)
; Preserva: BC, DE. Destrói: A, flags.
; -----------------------------------------------------------------------------
Mul16:
    PUSH BC
    PUSH DE
    LD B, 10h ; 16 iterações
    LD A, H
    LD C, L
    LD HL, 0000h
Mul16_Loop:
    ADD HL, HL
    SLA C
    RLA
    JR NC, Mul16_Skip
    ADD HL, DE
Mul16_Skip:
    DJNZ Mul16_Loop
    POP DE
    POP BC
    RET

ENDMOD
