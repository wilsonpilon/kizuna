; =============================================================================
; KIZUNA MSXLIB - num/scanhex
; texto hexadecimal para inteiro de 16 bits
; =============================================================================

MODULE num_scanhex
BANK 0

PUBLIC NUM_ScanHex16
EXTERN NUM_SkipBlanks, NUM_ParseHex16

; NUM_ScanHex16: pula espacos/TABs, aceita um prefixo opcional ("0x", "0X", "&H",
; "&h" ou "$") e le os digitos hexadecimais
; Entrada: BC = ponteiro, DE = fim do texto (0 = sem limite)
; Saída: HL = valor, BC = ponteiro depois do numero; carry = 1 se erro (HL = 0)
; Destrói: A, DE, flags.
NUM_ScanHex16:
    PUSH DE
    CALL NUM_SkipBlanks
    POP DE
    LD A, D             ; no fim do texto? entao nao ha o que espiar (e o byte
    OR E                ; seguinte nao pertence a string)
    JR Z, NUM_ScanHex16_Peek
    LD A, B
    CP D
    JR NZ, NUM_ScanHex16_Peek
    LD A, C
    CP E
    JR Z, NUM_ScanHex16_Digits
NUM_ScanHex16_Peek:
    LD A, (BC)
    CP 24h              ; '$'
    JR Z, NUM_ScanHex16_Skip1
    CP 26h              ; '&'
    JR Z, NUM_ScanHex16_Amp
    CP 30h              ; '0'
    JR NZ, NUM_ScanHex16_Digits
    INC BC              ; talvez "0x": olha o proximo
    LD A, (BC)
    DEC BC
    CP 78h              ; 'x'
    JR Z, NUM_ScanHex16_Skip2
    CP 58h              ; 'X'
    JR Z, NUM_ScanHex16_Skip2
    JR NUM_ScanHex16_Digits
NUM_ScanHex16_Amp:
    INC BC
    LD A, (BC)
    DEC BC
    CP 48h              ; 'H'
    JR Z, NUM_ScanHex16_Skip2
    CP 68h              ; 'h'
    JR Z, NUM_ScanHex16_Skip2
    JR NUM_ScanHex16_Digits
NUM_ScanHex16_Skip2:
    INC BC
NUM_ScanHex16_Skip1:
    INC BC
NUM_ScanHex16_Digits:
    JP NUM_ParseHex16

ENDMOD
