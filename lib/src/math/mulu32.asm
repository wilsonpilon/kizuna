; =============================================================================
; KIZUNA MSXLIB - math/mulu32
; multiplicacao 16x16 -> 32 bits, sem sinal
; =============================================================================

MODULE math_mulu32
BANK 0

PUBLIC MATH_MulU16x16

; MATH_MulU16x16: produto completo de 32 bits
; Entrada: HL, DE (sem sinal)
; Saída: DE:HL = HL * DE (DE = 16 bits altos, HL = 16 bits baixos)
; Preserva: BC. Destrói: A, flags.
MATH_MulU16x16:
    PUSH BC
    LD B, H
    LD C, L             ; BC = multiplicando
    LD HL, 0000h        ; HL:DE = acumulador (alto:baixo); DE traz o multiplicador
    LD A, 10h
MATH_MulU16x16_Loop:
    BIT 0, E
    JR Z, MATH_MulU16x16_NoAdd
    ADD HL, BC          ; carry = bit 16 da soma parcial
    JR MATH_MulU16x16_Shift
MATH_MulU16x16_NoAdd:
    OR A                ; carry = 0
MATH_MulU16x16_Shift:
    RR H
    RR L
    RR D
    RR E                ; desloca o acumulador de 32 bits para a direita, com o carry
    DEC A
    JR NZ, MATH_MulU16x16_Loop
    EX DE, HL           ; agora DE = alto, HL = baixo
    POP BC
    RET

ENDMOD
