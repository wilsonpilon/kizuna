; =============================================================================
; KIZUNA MSXLIB - num/cstrtonum
; texto terminado em zero para inteiro
; =============================================================================

MODULE num_cstrtonum
BANK 0

PUBLIC NUM_CStrToU16, NUM_CStrToI16, NUM_CStrToHex16
EXTERN NUM_ScanU16, NUM_ScanI16, NUM_ScanHex16

; As tres leem uma string terminada em zero.
; Entrada: HL = string
; Saída: HL = valor, DE = ponteiro para o primeiro caractere NAO lido; carry = 1
;        se nao havia numero ou ele nao cabe (HL = 0)
; Preserva: BC. Destrói: A, flags.

; NUM_CStrToU16: decimal sem sinal (espacos iniciais e '+' aceitos)
NUM_CStrToU16:
    PUSH BC
    LD B, H
    LD C, L
    LD DE, 0000h
    CALL NUM_ScanU16
    LD D, B
    LD E, C
    POP BC
    RET

; NUM_CStrToI16: decimal com sinal
NUM_CStrToI16:
    PUSH BC
    LD B, H
    LD C, L
    LD DE, 0000h
    CALL NUM_ScanI16
    LD D, B
    LD E, C
    POP BC
    RET

; NUM_CStrToHex16: hexadecimal, com prefixo opcional 0x, &H ou $
NUM_CStrToHex16:
    PUSH BC
    LD B, H
    LD C, L
    LD DE, 0000h
    CALL NUM_ScanHex16
    LD D, B
    LD E, C
    POP BC
    RET

ENDMOD
