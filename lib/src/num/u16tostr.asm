; =============================================================================
; KIZUNA MSXLIB - num/u16tostr
; inteiro de 16 bits sem sinal para texto decimal
; =============================================================================

MODULE num_u16tostr
BANK 0

PUBLIC NUM_U16ToCStr
EXTERN MATH_DivMod10

; NUM_U16ToCStr: escreve o valor em decimal (sem zeros a esquerda; 0 vira "0") em
; DE, terminado em zero. O destino precisa ter 6 bytes.
; Entrada: HL = valor (0..65535), DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC. Destrói: A, HL (fica 0), flags.
NUM_U16ToCStr:
    PUSH BC
    LD B, 00h           ; B = quantidade de digitos
NUM_U16ToCStr_Div:
    CALL MATH_DivMod10  ; HL = HL / 10, A = resto
    PUSH AF
    INC B
    LD A, H
    OR L
    JR NZ, NUM_U16ToCStr_Div
NUM_U16ToCStr_Out:
    POP AF              ; digitos saem do mais significativo para o menos
    ADD A, 30h
    LD (DE), A
    INC DE
    DJNZ NUM_U16ToCStr_Out
    XOR A
    LD (DE), A
    POP BC
    RET

ENDMOD
