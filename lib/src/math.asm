; =============================================================================
; KIZUNA MSXLIB - MATH.ASM
; Rotinas aritméticas de 16 bits para Z80 (Multiplicação e Divisão inteira)
; =============================================================================

MODULE MATH
BANK 0

PUBLIC Mul16, Div16

; -----------------------------------------------------------------------------
; Mul16: Multiplicação inteira não sinalizada de 16 bits
; Entrada: HL = multiplicando, DE = multiplicador
; Saída: HL = produto (16 bits mais baixos)
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

; -----------------------------------------------------------------------------
; Div16: Divisão inteira não sinalizada de 16 bits
; Entrada: HL = dividendo, DE = divisor
; Saída: HL = quociente, DE = resto
; Se divisor for 0: retorna HL = 0xFFFF, DE = 0
; -----------------------------------------------------------------------------
Div16:
    PUSH BC
    ; Tratar divisão por zero
    LD A, D
    OR E
    JR Z, Div16_Zero

    LD BC, 0000h ; BC acumula o resto parcial
    LD A, 10h   ; 16 iterações

Div16_Loop:
    ADD HL, HL  ; Desloca HL para a esquerda (MSB vai para Carry)
    RL C
    RL B        ; Desloca resto parcial (BC) incluindo Carry

    PUSH HL
    LD H, B
    LD L, C
    OR A
    SBC HL, DE  ; Resto parcial - Divisor
    JR C, Div16_Skip
    LD B, H
    LD C, L
    POP HL
    INC L       ; Liga bit 0 do quociente
    JR Div16_Next

Div16_Skip:
    POP HL

Div16_Next:
    DEC A
    JR NZ, Div16_Loop

    LD D, B     ; DE recebe o resto final
    LD E, C
    POP BC
    RET

Div16_Zero:
    LD HL, 0FFFFh
    LD DE, 0000h
    POP BC
    RET

ENDMOD
