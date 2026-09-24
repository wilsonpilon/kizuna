; =============================================================================
; KIZUNA MSXLIB - str/val
; string tamanho+dados para numero (VAL)
; =============================================================================

MODULE str_val
BANK 0

PUBLIC STR_Val, STR_HexVal
EXTERN NUM_ScanI16, NUM_ScanHex16

; STR_Val: valor decimal com sinal de uma string tamanho+dados (como o VAL do
; BASIC: espacos iniciais, '+'/'-' e digitos; le ate o primeiro caractere que
; nao serve). String sem numero: carry = 1 e HL = 0.
; Entrada: HL = string
; Saída: HL = valor (-32768..32767); carry = 1 se nao ha numero ou ele nao cabe
; Preserva: BC, DE. Destrói: A, flags.
STR_Val:
    PUSH BC
    PUSH DE
    LD A, (HL)
    INC HL
    LD B, H
    LD C, L             ; BC = primeiro caractere
    LD E, A
    LD D, 00h
    ADD HL, DE
    EX DE, HL           ; DE = fim (depois do ultimo caractere)
    CALL NUM_ScanI16
    POP DE
    POP BC
    RET

; STR_HexVal: valor hexadecimal de uma string tamanho+dados (prefixo opcional 0x, &H ou $)
; Entrada: HL = string. Saída: HL = valor (0..65535); carry = 1 se nao ha numero ou nao cabe
; Preserva: BC, DE. Destrói: A, flags.
STR_HexVal:
    PUSH BC
    PUSH DE
    LD A, (HL)
    INC HL
    LD B, H
    LD C, L
    LD E, A
    LD D, 00h
    ADD HL, DE
    EX DE, HL
    CALL NUM_ScanHex16
    POP DE
    POP BC
    RET

ENDMOD
