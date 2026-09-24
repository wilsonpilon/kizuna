; =============================================================================
; KIZUNA MSXLIB - math/rand16
; gerador de numeros pseudoaleatorios de 16 bits (xorshift, periodo 65535)
; =============================================================================

MODULE math_rand16
BANK 0

PUBLIC MATH_RandSeed, MATH_Rand16

; Estado do gerador (nunca pode ser zero: o xorshift ficaria preso em 0).
MATH_RandState:
    DW 0ACE1h

; MATH_RandSeed: define o estado do gerador (a mesma semente repete a mesma
; sequencia). A semente 0 e trocada por 0ACE1h.
; Entrada: HL = semente
; Preserva: A, BC, DE, HL. Destrói: flags.
MATH_RandSeed:
    PUSH AF
    PUSH HL
    LD A, H
    OR L
    JR NZ, MATH_RandSeed_Store
    LD HL, 0ACE1h
MATH_RandSeed_Store:
    LD (MATH_RandState), HL
    POP HL
    POP AF
    RET

; MATH_Rand16: proximo numero pseudoaleatorio
; Saída: HL = 1..65535 (o 0 nunca sai)
; Algoritmo xorshift de Marsaglia com o trio (7, 9, 8):
;   x ^= x << 7;  x ^= x >> 9;  x ^= x << 8
; Preserva: A, BC, DE. Destrói: flags.
MATH_Rand16:
    PUSH AF
    PUSH BC
    PUSH DE
    LD HL, (MATH_RandState)
    LD D, H
    LD E, L
    LD B, 07h
MATH_Rand16_Shl7:
    ADD HL, HL
    DJNZ MATH_Rand16_Shl7   ; HL = x << 7
    LD A, H
    XOR D
    LD H, A
    LD A, L
    XOR E
    LD L, A                 ; HL = x ^ (x << 7)
    LD A, H
    SRL A                   ; A = H >> 1 = (HL >> 9)
    XOR L
    LD L, A                 ; x ^= x >> 9
    LD A, H
    XOR L
    LD H, A                 ; x ^= x << 8
    LD (MATH_RandState), HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
