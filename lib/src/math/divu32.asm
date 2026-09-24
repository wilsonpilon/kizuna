; =============================================================================
; KIZUNA MSXLIB - math/divu32
; divisao 32 / 16 bits, sem sinal
; =============================================================================

MODULE math_divu32
BANK 0

PUBLIC MATH_DivU32By16
EXTERN Div16

; MATH_DivU32By16: divide um numero de 32 bits por um de 16, sem sinal
; Entrada: DE:HL = dividendo (DE = 16 bits altos), BC = divisor
; Saída: DE:HL = quociente de 32 bits, BC = resto (0..divisor-1)
; Divisor zero: DE:HL = 0FFFFFFFFh, BC = 0.
; Destrói: A, flags.
; Metodo: primeiro divide a palavra alta (quociente alto + resto parcial) e depois
; divide (resto parcial : palavra baixa) por 16 passos de subtrai-e-desloca, sem
; laco (REPT), o que evita gastar um registrador com contador.
MATH_DivU32By16:
    LD A, B
    OR C
    JP Z, MATH_DivU32By16_Zero
    PUSH HL             ; guarda a palavra baixa do dividendo
    EX DE, HL           ; HL = palavra alta
    LD D, B
    LD E, C             ; DE = divisor
    CALL Div16          ; HL = quociente alto, DE = resto parcial
    EX (SP), HL         ; HL = palavra baixa; pilha = quociente alto
    ; DE = resto (<divisor), HL = dividendo baixo -> vira quociente baixo
REPT 16
    ADD HL, HL          ; desloca o dividendo; o bit que sai entra no resto
    RL E
    RL D
    JR C, .sub          ; resto estourou 16 bits: com certeza >= divisor
    LD A, D
    CP B
    JR C, .next
    JR NZ, .sub
    LD A, E
    CP C
    JR C, .next
.sub:
    LD A, E
    SUB C
    LD E, A
    LD A, D
    SBC A, B
    LD D, A             ; resto -= divisor
    INC L               ; bit do quociente
.next:
ENDR
    LD B, D
    LD C, E             ; BC = resto final
    POP DE              ; DE = quociente alto; HL = quociente baixo
    RET

MATH_DivU32By16_Zero:
    LD DE, 0FFFFh
    LD HL, 0FFFFh
    LD BC, 0000h
    RET

ENDMOD
