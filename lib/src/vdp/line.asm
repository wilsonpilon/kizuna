; =============================================================================
; KIZUNA MSXLIB - vdp/line
; linha (Bresenham)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_line
BANK 0

PUBLIC VDP_Line
EXTERN VDP_PSet_Raw

; -----------------------------------------------------------------------------
; VDP_Line: Desenha uma linha reta entre (X1, Y1) e (X2, Y2) via Bresenham
; Pilha (caller cleanup): X1, Y1, X2, Y2, Color (16-bit cada)
; -----------------------------------------------------------------------------
VDP_Line:
    PUSH IX
    LD IX, 0000h
    ADD IX, SP
    PUSH BC
    PUSH DE
    PUSH HL

    ; Se horizontal puro (Y1 == Y2): otimização rápida
    LD A, (IX+10)
    CP (IX+6)
    JR Z, VDP_Line_H_Start

    ; Se vertical puro (X1 == X2): otimização rápida
    LD A, (IX+12)
    CP (IX+8)
    JR Z, VDP_Line_V_Start

    JR VDP_Line_General

VDP_Line_H_Start:
    ; Linha horizontal de min(X1,X2) até max(X1,X2)
    LD C, (IX+12) ; X1
    LD A, (IX+8)  ; X2
    CP C
    JR NC, VDP_Line_H_X1_Le_X2
    LD C, (IX+8)
    LD A, (IX+12)
VDP_Line_H_X1_Le_X2:
    LD (VDP_Line_H_End), A
    LD E, (IX+10) ; Y
    LD D, 00h
    ; DI unico cobrindo a linha inteira (ver comentario em VDP_BoxFill_General).
    DI
VDP_Line_HLoop:
    LD B, 00h
    LD A, (IX+4)  ; Cor
    CALL VDP_PSet_Raw
    LD A, (VDP_Line_H_End)
    CP C
    JP Z, VDP_Line_Exit
    INC C
    JR VDP_Line_HLoop

VDP_Line_V_Start:
    ; Linha vertical de min(Y1,Y2) até max(Y1,Y2)
    LD E, (IX+10) ; Y1
    LD A, (IX+6)  ; Y2
    CP E
    JR NC, VDP_Line_V_Y1_Le_Y2
    LD E, (IX+6)
    LD A, (IX+10)
VDP_Line_V_Y1_Le_Y2:
    LD (VDP_Line_H_End), A
    LD C, (IX+12) ; X
    LD B, 00h
    LD D, 00h
    DI
VDP_Line_VLoop:
    LD A, (IX+4)  ; Cor
    CALL VDP_PSet_Raw
    LD A, (VDP_Line_H_End)
    CP E
    JP Z, VDP_Line_Exit
    INC E
    JR VDP_Line_VLoop

VDP_Line_General:
    ; Bresenham completo
    ; DX e SX
    LD A, (IX+8)  ; X2
    SUB (IX+12)   ; X2 - X1
    JR NC, VDP_Line_SX_Pos
    NEG
    LD (VDP_Line_DX), A
    LD A, 0FFh   ; SX = -1
    LD (VDP_Line_SX), A
    JR VDP_Line_CalcDY
VDP_Line_SX_Pos:
    LD (VDP_Line_DX), A
    LD A, 01h    ; SX = +1
    LD (VDP_Line_SX), A

VDP_Line_CalcDY:
    ; DY e SY: DY será armazenado como valor negativo (-abs(Y2 - Y1)) em 16-bit
    LD A, (IX+6)  ; Y2
    SUB (IX+10)   ; Y2 - Y1
    JR NC, VDP_Line_SY_Pos
    NEG           ; A = abs(Y2 - Y1)
    LD L, A
    LD H, 00h
    XOR A
    SUB L
    LD L, A
    SBC A, A
    LD H, A
    LD (VDP_Line_DY), HL
    LD A, 0FFh   ; SY = -1
    LD (VDP_Line_SY), A
    JR VDP_Line_CalcErr
VDP_Line_SY_Pos:
    LD L, A
    LD H, 00h
    XOR A
    SUB L
    LD L, A
    SBC A, A
    LD H, A
    LD (VDP_Line_DY), HL
    LD A, 01h    ; SY = +1
    LD (VDP_Line_SY), A

VDP_Line_CalcErr:
    ; Err = DX + DY
    LD HL, (VDP_Line_DY)
    EX DE, HL
    LD A, (VDP_Line_DX)
    LD L, A
    LD H, 00h
    ADD HL, DE
    LD (VDP_Line_Err), HL

    ; Ponto inicial (X, Y)
    LD A, (IX+12)
    LD (VDP_Line_CurX), A
    LD A, (IX+10)
    LD (VDP_Line_CurY), A

    DI
VDP_Line_Loop:
    ; Plota ponto atual
    LD A, (VDP_Line_CurX)
    LD C, A
    LD B, 00h
    LD A, (VDP_Line_CurY)
    LD E, A
    LD D, 00h
    LD A, (IX+4)
    CALL VDP_PSet_Raw

    ; Verifica se chegou ao fim: CurX == X2 && CurY == Y2
    LD A, (VDP_Line_CurX)
    CP (IX+8)
    JR NZ, VDP_Line_Step
    LD A, (VDP_Line_CurY)
    CP (IX+6)
    JP Z, VDP_Line_Exit

VDP_Line_Step:
    ; E2 = 2 * Err
    LD HL, (VDP_Line_Err)
    ADD HL, HL
    LD (VDP_Line_E2), HL

    ; Se E2 >= DY: Err += DY, CurX += SX
    ; HL = E2 - DY
    LD HL, (VDP_Line_DY)
    EX DE, HL
    LD HL, (VDP_Line_E2)
    OR A
    SBC HL, DE          ; HL = E2 - DY
    JP M, VDP_Line_CheckE2DX
    LD HL, (VDP_Line_Err)
    ADD HL, DE
    LD (VDP_Line_Err), HL
    LD A, (VDP_Line_SX)
    LD B, A
    LD A, (VDP_Line_CurX)
    ADD A, B
    LD (VDP_Line_CurX), A

VDP_Line_CheckE2DX:
    ; Se E2 <= DX: Err += DX, CurY += SY
    LD A, (VDP_Line_DX)
    LD E, A
    LD D, 00h
    LD HL, (VDP_Line_E2)
    EX DE, HL
    OR A
    SBC HL, DE          ; HL = DX - E2
    JP M, VDP_Line_Loop
    LD HL, (VDP_Line_Err)
    LD A, (VDP_Line_DX)
    LD E, A
    LD D, 00h
    ADD HL, DE
    LD (VDP_Line_Err), HL
    LD A, (VDP_Line_SY)
    LD B, A
    LD A, (VDP_Line_CurY)
    ADD A, B
    LD (VDP_Line_CurY), A
    JR VDP_Line_Loop

VDP_Line_Exit:
    EI
    POP HL
    POP DE
    POP BC
    POP IX
    RET

VDP_Line_H_End: DB 00h
VDP_Line_DX:    DB 00h
VDP_Line_SX:    DB 00h
VDP_Line_DY:    DW 0000h
VDP_Line_SY:    DB 00h
VDP_Line_Err:   DW 0000h
VDP_Line_E2:    DW 0000h
VDP_Line_CurX:  DB 00h
VDP_Line_CurY:  DB 00h

ENDMOD
