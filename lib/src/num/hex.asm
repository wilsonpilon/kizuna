; =============================================================================
; KIZUNA MSXLIB - num/hex
; inteiro para texto hexadecimal (maiusculo, largura fixa)
; =============================================================================

MODULE num_hex
BANK 0

PUBLIC NUM_U8ToHex, NUM_U16ToHex
EXTERN CHAR_HexChar

; NUM_U8ToHex: A em hexadecimal, 2 digitos, terminado em zero (destino de 3 bytes)
; Entrada: A = valor, DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC, HL. Destrói: A, flags.
NUM_U8ToHex:
    PUSH AF
    RRCA
    RRCA
    RRCA
    RRCA                ; nibble alto para baixo
    CALL CHAR_HexChar
    LD (DE), A
    INC DE
    POP AF
    CALL CHAR_HexChar
    LD (DE), A
    INC DE
    XOR A
    LD (DE), A
    RET

; NUM_U16ToHex: HL em hexadecimal, 4 digitos, terminado em zero (destino de 5 bytes)
; Entrada: HL = valor, DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC, HL. Destrói: A, flags.
NUM_U16ToHex:
    LD A, H
    CALL NUM_U8ToHex
    LD A, L
    JP NUM_U8ToHex      ; o segundo par sobrescreve o terminador do primeiro

ENDMOD
