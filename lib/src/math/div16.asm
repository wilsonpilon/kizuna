; =============================================================================
; KIZUNA MSXLIB - math/div16
; divisao 16/16
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE math_div16
BANK 0

PUBLIC Div16

; -----------------------------------------------------------------------------
; Div16: Divisão inteira não sinalizada de 16 bits
; Entrada: HL = dividendo, DE = divisor
; Saída: HL = quociente, DE = resto
; Se divisor for 0: retorna HL = 0xFFFF, DE = 0
; Preserva: BC. Destrói: A, flags.
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
