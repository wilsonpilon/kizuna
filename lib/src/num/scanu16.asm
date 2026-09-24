; =============================================================================
; KIZUNA MSXLIB - num/scanu16
; texto para inteiro de 16 bits sem sinal (espacos, '+' e digitos)
; =============================================================================

MODULE num_scanu16
BANK 0

PUBLIC NUM_ScanU16
EXTERN NUM_SkipBlanks, NUM_ParseU16

; NUM_ScanU16: pula espacos/TABs, aceita um '+' opcional e le os digitos decimais
; Entrada: BC = ponteiro, DE = fim do texto (0 = sem limite)
; Saída: HL = valor, BC = ponteiro depois do numero; carry = 1 se erro (HL = 0)
; Destrói: A, DE, flags.
NUM_ScanU16:
    PUSH DE
    CALL NUM_SkipBlanks
    POP DE
    LD A, D             ; no fim do texto? entao nao ha o que espiar (e o byte
    OR E                ; seguinte nao pertence a string)
    JR Z, NUM_ScanU16_Peek
    LD A, B
    CP D
    JR NZ, NUM_ScanU16_Peek
    LD A, C
    CP E
    JR Z, NUM_ScanU16_Digits
NUM_ScanU16_Peek:
    LD A, (BC)
    CP 2Bh              ; '+'
    JR NZ, NUM_ScanU16_Digits
    INC BC
NUM_ScanU16_Digits:
    JP NUM_ParseU16

ENDMOD
