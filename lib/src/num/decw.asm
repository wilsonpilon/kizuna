; =============================================================================
; KIZUNA MSXLIB - num/decw
; inteiro sem sinal em decimal com largura fixa e preenchimento
; =============================================================================

MODULE num_decw
BANK 0

PUBLIC NUM_U16ToDecW
EXTERN NUM_U16ToCStr

; NUM_U16ToDecW: escreve o valor em decimal ALINHADO A DIREITA numa largura fixa,
; preenchendo a esquerda com o caractere dado (espaco ou '0', por exemplo), e
; termina com zero. Se o numero tem mais digitos que a largura, ficam so os
; digitos menos significativos.
; Entrada: HL = valor, DE = destino (largura + 1 bytes), B = largura (0 = string vazia),
;          A = caractere de preenchimento
; Saída: DE = endereco do terminador
; Destrói: A, BC, HL, flags. Nao reentrante (usa um buffer proprio).
NUM_U16ToDecW:
    PUSH AF             ; preenchimento
    PUSH BC             ; largura em B
    PUSH DE             ; destino
    LD DE, NUM_U16ToDecW_Buf
    CALL NUM_U16ToCStr  ; DE = fim dos digitos
    LD HL, NUM_U16ToDecW_Buf
    LD A, E
    SUB L
    LD C, A             ; C = quantidade de digitos
    POP DE              ; DE = destino
    POP HL              ; H = largura (o BC empilhado; POP BC apagaria o C recem calculado)
    LD B, H
    POP AF              ; A = preenchimento
    LD HL, NUM_U16ToDecW_Buf
    PUSH AF
    LD A, B
    SUB C               ; A = largura - digitos (carry se sobram digitos)
    JR C, NUM_U16ToDecW_Trunc
    LD B, A             ; B = quantos caracteres de preenchimento
    POP AF              ; A = preenchimento
    INC B
NUM_U16ToDecW_Pad:
    DEC B
    JR Z, NUM_U16ToDecW_Digits
    LD (DE), A
    INC DE
    JR NUM_U16ToDecW_Pad
NUM_U16ToDecW_Trunc:
    POP AF              ; descarta o preenchimento
    LD A, C
    SUB B               ; A = digitos a pular (os mais significativos)
    ADD A, L
    LD L, A
    JR NC, NUM_U16ToDecW_Cut
    INC H
NUM_U16ToDecW_Cut:
    LD C, B             ; copia so `largura` digitos
NUM_U16ToDecW_Digits:
    LD A, C
    OR A
    JR Z, NUM_U16ToDecW_End     ; largura 0
    LD B, 00h
    LDIR
NUM_U16ToDecW_End:
    XOR A
    LD (DE), A
    RET

NUM_U16ToDecW_Buf:
    DS 8

ENDMOD
