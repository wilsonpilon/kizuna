; =============================================================================
; KIZUNA MSXLIB - num/i16tostr
; inteiro de 16 bits com sinal para texto decimal
; =============================================================================

MODULE num_i16tostr
BANK 0

PUBLIC NUM_I16ToCStr
EXTERN NUM_U16ToCStr, MATH_Neg16

; NUM_I16ToCStr: como NUM_U16ToCStr, mas com sinal ("-" na frente dos negativos).
; O destino precisa ter 7 bytes.
; Entrada: HL = valor (-32768..32767), DE = destino
; Saída: DE = endereco do terminador
; Preserva: BC. Destrói: A, HL, flags.
NUM_I16ToCStr:
    BIT 7, H
    JP Z, NUM_U16ToCStr
    LD A, 2Dh           ; '-'
    LD (DE), A
    INC DE
    CALL MATH_Neg16     ; -32768 continua 8000h, que lido sem sinal e 32768
    JP NUM_U16ToCStr

ENDMOD
