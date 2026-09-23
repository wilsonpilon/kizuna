; =============================================================================
; KIZUNA MSXLIB - vdp/boxfill
; retangulo preenchido
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_boxfill
BANK 0

PUBLIC VDP_BoxFill
EXTERN VDP_FillVRAM, VDP_PSet_Raw

; -----------------------------------------------------------------------------
; VDP_BoxFill: Preenche uma região retangular na tela (SCREEN 2)
; Pilha (caller cleanup): X1, Y1, X2, Y2, Color (16-bit cada)
; -----------------------------------------------------------------------------
VDP_BoxFill:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    PUSH BC
    PUSH DE
    PUSH HL

    ; Parâmetros relativos a IX:
    ; IX+4: Color
    ; IX+6: Y2
    ; IX+8: X2
    ; IX+10: Y1
    ; IX+12: X1

    ; Otimização: Se for tela cheia (0,0)-(255,191), usa VDP_FillVRAM
    LD A, (IX+12)
    OR A
    JR NZ, VDP_BoxFill_General
    LD A, (IX+10)
    OR A
    JR NZ, VDP_BoxFill_General
    LD A, (IX+8)
    CP 0FFh
    JR C, VDP_BoxFill_General
    LD A, (IX+6)
    CP 0BFh
    JR C, VDP_BoxFill_General

    ; Limpa tela inteira
    LD HL, 0000h
    LD BC, 1800h
    XOR A
    CALL VDP_FillVRAM

    LD A, (IX+4)
    AND 0Fh
    LD B, A
    SLA A
    SLA A
    SLA A
    SLA A
    OR B
    LD HL, 2000h
    LD BC, 1800h
    CALL VDP_FillVRAM
    JR VDP_BoxFill_Exit

VDP_BoxFill_General:
    ; Preenchimento por linhas e colunas. DI unico cobrindo o preenchimento
    ; inteiro: com muitos pontos em sequencia, deixar cada ponto reabrir
    ; interrupcoes (como o VDP_PSet publico faz) da chance da interrupcao
    ; de ~60Hz do MSX corromper registradores entre um ponto e outro.
    DI
    LD E, (IX+10) ; Y = Y1
VDP_BoxFill_YLoop:
    LD C, (IX+12) ; X = X1
VDP_BoxFill_XLoop:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet_Raw
    INC C
    LD A, C
    CP (IX+8)
    JR C, VDP_BoxFill_XLoop
    JR Z, VDP_BoxFill_XLast
    JR VDP_BoxFill_YNext
VDP_BoxFill_XLast:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet_Raw
VDP_BoxFill_YNext:
    INC E
    LD A, E
    CP (IX+6)
    JR C, VDP_BoxFill_YLoop
    JR Z, VDP_BoxFill_YLast
    JR VDP_BoxFill_Exit
VDP_BoxFill_YLast:
    LD C, (IX+12)
VDP_BoxFill_XLastLoop:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet_Raw
    INC C
    LD A, C
    CP (IX+8)
    JR C, VDP_BoxFill_XLastLoop
    JR Z, VDP_BoxFill_XLastFinal
    JR VDP_BoxFill_Exit
VDP_BoxFill_XLastFinal:
    LD B, 00h
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet_Raw

VDP_BoxFill_Exit:
    EI
    POP HL
    POP DE
    POP BC
    POP IX
    RET

ENDMOD
