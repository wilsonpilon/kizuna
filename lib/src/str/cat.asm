; =============================================================================
; KIZUNA MSXLIB - str/cat
; concatenacao de strings tamanho+dados
; =============================================================================

MODULE str_cat
BANK 0

PUBLIC STR_Cat

; STR_Cat: anexa a string de HL ao fim da string de DE. Se passar de 255
; caracteres, a parte que nao cabe e descartada.
; Entrada: HL = origem (o que anexar), DE = destino (ate 256 bytes)
; Destrói: A, BC, DE, HL, flags.
STR_Cat:
    LD C, (HL)          ; C = tamanho do que sera anexado
    INC HL              ; HL = dados da origem
    LD A, (DE)          ; A = tamanho atual do destino
    PUSH AF
    ADD A, C
    JR NC, STR_Cat_Fit
    LD A, 0FFh          ; passou de 255: trunca
STR_Cat_Fit:
    LD (DE), A          ; novo tamanho
    POP BC              ; B = tamanho antigo do destino
    SUB B               ; A = quantos caracteres realmente entram
    LD C, A
    LD A, B
    LD B, 00h           ; BC = quantidade a copiar
    INC DE              ; DE = dados do destino
    ADD A, E
    LD E, A
    JR NC, STR_Cat_Pos
    INC D               ; DE = fim dos dados atuais
STR_Cat_Pos:
    LD A, B
    OR C
    RET Z
    LDIR
    RET

ENDMOD
