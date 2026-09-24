; =============================================================================
; KIZUNA MSXLIB - str/fromnum
; numero para string tamanho+dados (STR$ e HEX$)
; =============================================================================

MODULE str_fromnum
BANK 0

PUBLIC STR_FromU16, STR_FromI16, STR_Hex8, STR_Hex16
EXTERN NUM_U16ToCStr, NUM_I16ToCStr, NUM_U8ToHex, NUM_U16ToHex

; Todas escrevem em DE uma string tamanho+dados (o destino precisa ter 8 bytes
; para FromI16, 7 para FromU16, 3 para Hex8 e 5 para Hex16).
; Destrói: A, HL, flags. Preservam: BC.

; STR_FromU16: HL em decimal sem sinal (STR$ de um valor sem sinal)
STR_FromU16:
    PUSH DE
    INC DE
    CALL NUM_U16ToCStr
    JR STR_FromNum_Len

; STR_FromI16: HL em decimal com sinal
STR_FromI16:
    PUSH DE
    INC DE
    CALL NUM_I16ToCStr
    JR STR_FromNum_Len

; STR_Hex8: A em hexadecimal maiusculo, 2 digitos
STR_Hex8:
    PUSH DE
    INC DE
    CALL NUM_U8ToHex
    JR STR_FromNum_Len

; STR_Hex16: HL em hexadecimal maiusculo, 4 digitos
STR_Hex16:
    PUSH DE
    INC DE
    CALL NUM_U16ToHex
STR_FromNum_Len:            ; DE = terminador; a pilha guarda o inicio da string
    POP HL
    LD A, E
    SUB L
    DEC A               ; tamanho = terminador - inicio - 1
    LD (HL), A
    RET

ENDMOD
