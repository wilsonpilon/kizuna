; =============================================================================
; KIZUNA MSXLIB - math/fixed88
; ponto fixo com sinal 8.8 (8 bits de parte inteira, 8 de fracao): 1.0 = 0100h
; =============================================================================

MODULE math_fixed88
BANK 0

PUBLIC MATH_FixMul88, MATH_FixDiv88
EXTERN MATH_MulS16x16, MATH_Neg16, MATH_DivU32By16

; MATH_FixMul88: HL = HL * DE em 8.8 com sinal (o produto de 32 bits, sem os 8
; bits de baixo). Estoura em silencio se o resultado nao couber em 8.8.
; Entrada: HL, DE
; Saída: HL
; Preserva: BC, DE. Destrói: A, flags.
MATH_FixMul88:
    PUSH DE
    PUSH BC
    CALL MATH_MulS16x16 ; DE:HL = produto de 32 bits
    LD L, H
    LD H, E             ; bits 23..8 do produto
    POP BC
    POP DE
    RET

; MATH_FixDiv88: HL = HL / DE em 8.8 com sinal (o quociente de (HL << 8) / DE),
; truncado em direcao a zero. Divisor zero: HL = 7FFFh (dividendo >= 0) ou 8000h.
; Entrada: HL = dividendo, DE = divisor
; Saída: HL
; Preserva: BC, DE. Destrói: A, flags.
MATH_FixDiv88:
    PUSH DE
    PUSH BC
    LD A, H
    XOR D
    PUSH AF             ; bit 7 do A empilhado = sinal do resultado
    LD A, H
    PUSH AF             ; bit 7 = sinal do dividendo (so para divisor zero)
    BIT 7, H
    CALL NZ, MATH_Neg16 ; HL = |dividendo|
    EX DE, HL
    BIT 7, H
    CALL NZ, MATH_Neg16
    EX DE, HL           ; DE = |divisor|
    LD A, D
    OR E
    JR Z, MATH_FixDiv88_Zero
    LD B, D
    LD C, E             ; BC = divisor
    LD D, 00h
    LD E, H             ; DE:HL = |dividendo| << 8 = 00 : Hh : Hl : 00
    LD H, L
    LD L, 00h
    CALL MATH_DivU32By16
    POP AF              ; descarta o sinal do dividendo
    POP AF
    BIT 7, A
    CALL NZ, MATH_Neg16 ; quociente (16 bits baixos) com o sinal certo
    POP BC
    POP DE
    RET
MATH_FixDiv88_Zero:
    POP AF              ; sinal do dividendo
    LD HL, 7FFFh
    BIT 7, A
    JR Z, MATH_FixDiv88_ZeroOut
    LD HL, 8000h
MATH_FixDiv88_ZeroOut:
    POP AF              ; descarta o sinal do resultado
    POP BC
    POP DE
    RET

ENDMOD
