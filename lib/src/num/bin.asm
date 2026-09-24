; =============================================================================
; KIZUNA MSXLIB - num/bin
; inteiro para texto binario (largura fixa)
; =============================================================================

MODULE num_bin
BANK 0

PUBLIC NUM_U8ToBin, NUM_U16ToBin

; NUM_U8ToBin: A em binario, 8 digitos, terminado em zero (destino de 9 bytes)
; Entrada: A = valor, DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC, HL. Destrói: A, flags.
NUM_U8ToBin:
    PUSH BC
    LD B, 08h
NUM_U8ToBin_Loop:
    RLCA                ; o bit mais alto vai para o carry
    PUSH AF
    LD A, 30h
    ADC A, 00h          ; '0' + carry
    LD (DE), A
    INC DE
    POP AF
    DJNZ NUM_U8ToBin_Loop
    XOR A
    LD (DE), A
    POP BC
    RET

; NUM_U16ToBin: HL em binario, 16 digitos, terminado em zero (destino de 17 bytes)
; Entrada: HL = valor, DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC, HL. Destrói: A, flags.
NUM_U16ToBin:
    LD A, H
    CALL NUM_U8ToBin
    LD A, L
    JP NUM_U8ToBin

ENDMOD
