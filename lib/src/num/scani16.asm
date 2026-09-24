; =============================================================================
; KIZUNA MSXLIB - num/scani16
; texto para inteiro de 16 bits com sinal
; =============================================================================

MODULE num_scani16
BANK 0

PUBLIC NUM_ScanI16
EXTERN NUM_SkipBlanks, NUM_ParseU16, MATH_Neg16

; NUM_ScanI16: pula espacos/TABs, aceita '+' ou '-' e le os digitos decimais
; Entrada: BC = ponteiro, DE = fim do texto (0 = sem limite)
; Saída: HL = valor (-32768..32767), BC = ponteiro depois do numero;
;        carry = 1 se nao havia digitos ou o valor nao cabe (HL = 0)
; Destrói: A, DE, flags.
NUM_ScanI16:
    PUSH DE
    CALL NUM_SkipBlanks
    POP DE
    LD A, D             ; no fim do texto? entao nao ha o que espiar (e o byte
    OR E                ; seguinte nao pertence a string)
    JR Z, NUM_ScanI16_Peek
    LD A, B
    CP D
    JR NZ, NUM_ScanI16_Peek
    LD A, C
    CP E
    JR Z, NUM_ScanI16_Pos
NUM_ScanI16_Peek:
    LD A, (BC)
    CP 2Dh              ; '-'
    JR NZ, NUM_ScanI16_Plus
    INC BC
    LD A, 01h
    JR NUM_ScanI16_Sign
NUM_ScanI16_Plus:
    CP 2Bh              ; '+'
    JR NZ, NUM_ScanI16_Pos
    INC BC
NUM_ScanI16_Pos:
    XOR A
NUM_ScanI16_Sign:
    PUSH AF             ; A = 1 se negativo
    CALL NUM_ParseU16
    JR C, NUM_ScanI16_Err
    POP AF
    OR A
    JR NZ, NUM_ScanI16_Neg
    BIT 7, H
    JR NZ, NUM_ScanI16_Range   ; positivo acima de 32767
    OR A                ; carry = 0
    RET
NUM_ScanI16_Neg:
    LD A, H
    CP 80h
    JR C, NUM_ScanI16_NegOk
    JR NZ, NUM_ScanI16_Range   ; acima de 32768
    LD A, L
    OR A
    JR NZ, NUM_ScanI16_Range
NUM_ScanI16_NegOk:
    CALL MATH_Neg16
    OR A                ; carry = 0
    RET
NUM_ScanI16_Range:
    LD HL, 0000h
    SCF
    RET
NUM_ScanI16_Err:
    POP DE              ; descarta o sinal (POP DE nao mexe nos flags; POP AF apagaria o carry)
    RET                 ; carry = 1 e HL = 0, vindos de NUM_ParseU16

ENDMOD
