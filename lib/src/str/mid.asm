; =============================================================================
; KIZUNA MSXLIB - str/mid
; MID$: um trecho no meio
; =============================================================================

MODULE str_mid
BANK 0

PUBLIC STR_Mid

; STR_Mid: B caracteres da string de HL a partir da posicao A (a primeira e 1; 0
; vale 1). Se a posicao passa do fim, o resultado e vazio; se faltam caracteres,
; devolve os que existirem.
; Entrada: HL = origem, A = posicao inicial, B = quantidade, DE = destino
;          (ate 256 bytes; nao pode ser a propria origem)
; Destrói: A, BC, DE, HL, flags.
STR_Mid:
    OR A
    JR NZ, STR_Mid_Start
    INC A
STR_Mid_Start:
    LD C, (HL)          ; C = tamanho
    INC HL              ; HL = dados
    CP C
    JR Z, STR_Mid_Ok
    JR C, STR_Mid_Ok
    XOR A               ; posicao > tamanho: resultado vazio
    LD (DE), A
    RET
STR_Mid_Ok:
    DEC A               ; A = deslocamento a partir do inicio
    PUSH AF
    ADD A, L
    LD L, A
    JR NC, STR_Mid_Ptr
    INC H
STR_Mid_Ptr:
    POP AF              ; A = deslocamento de novo
    SUB C
    NEG                 ; A = tamanho - deslocamento = quantos existem daqui
    CP B
    JR C, STR_Mid_Use   ; existem menos que o pedido: vale o que existe
    LD A, B
STR_Mid_Use:
    LD (DE), A
    INC DE
    LD C, A
    LD B, 00h
    OR A
    RET Z
    LDIR
    RET

ENDMOD
