; =============================================================================
; KIZUNA MSXLIB - str/toc
; converte string tamanho+dados para terminada em zero
; =============================================================================

MODULE str_toc
BANK 0

PUBLIC STR_ToC

; STR_ToC: copia a string tamanho+dados de HL para DE e termina com zero. O
; destino precisa ter tamanho + 1 bytes.
; Entrada: HL = string tamanho+dados, DE = destino
; Saída: DE = endereco do terminador
; Destrói: A, BC, HL, flags.
STR_ToC:
    LD C, (HL)
    LD B, 00h
    INC HL
    LD A, C
    OR A
    JR Z, STR_ToC_Term
    LDIR
STR_ToC_Term:
    XOR A
    LD (DE), A
    RET

ENDMOD
