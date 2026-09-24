; =============================================================================
; KIZUNA MSXLIB - vdp/hwbox
; contorno de retangulo (4 linhas) e retangulo cheio entre dois cantos
; =============================================================================

MODULE vdp_hwbox
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwBox, VDP_HwBoxFill
EXTERN VDP_HwLine, VDP_HwSpan, VDP_CmdRun
EXTERN VDP_Cmd_SX, VDP_Cmd_DX, VDP_Cmd_DY, VDP_Cmd_NX, VDP_Cmd_NY, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; VDP_HwBox: contorno de um retangulo com cantos opostos (x1,y1) e (x2,y2), em qualquer
; ordem, desenhado com 4 linhas. Os cantos sao desenhados duas vezes: com a operacao
; XOR eles ficam apagados (use a copia, VDP_LOP_IMP, ou VDP_HwBoxFill).
; Entrada: HL = cantos em RAM: x1(2) y1(2) x2(2) y2(2), A = cor, B = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwBox:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_HwBox_Color), A
    LD A, B
    LD (VDP_HwBox_Op), A
    LD DE, VDP_HwBox_X1
    LD BC, 0008h
    LDIR
    LD HL, (VDP_HwBox_X1)
    LD (VDP_HwBox_LX1), HL
    LD HL, (VDP_HwBox_Y1)
    LD (VDP_HwBox_LY1), HL
    LD HL, (VDP_HwBox_X2)
    LD (VDP_HwBox_LX2), HL
    LD HL, (VDP_HwBox_Y1)
    LD (VDP_HwBox_LY2), HL
    LD HL, VDP_HwBox_LX1
    LD A, (VDP_HwBox_Op)
    LD B, A
    LD A, (VDP_HwBox_Color)
    CALL VDP_HwLine
    LD HL, (VDP_HwBox_X2)
    LD (VDP_HwBox_LX1), HL
    LD HL, (VDP_HwBox_Y1)
    LD (VDP_HwBox_LY1), HL
    LD HL, (VDP_HwBox_X2)
    LD (VDP_HwBox_LX2), HL
    LD HL, (VDP_HwBox_Y2)
    LD (VDP_HwBox_LY2), HL
    LD HL, VDP_HwBox_LX1
    LD A, (VDP_HwBox_Op)
    LD B, A
    LD A, (VDP_HwBox_Color)
    CALL VDP_HwLine
    LD HL, (VDP_HwBox_X2)
    LD (VDP_HwBox_LX1), HL
    LD HL, (VDP_HwBox_Y2)
    LD (VDP_HwBox_LY1), HL
    LD HL, (VDP_HwBox_X1)
    LD (VDP_HwBox_LX2), HL
    LD HL, (VDP_HwBox_Y2)
    LD (VDP_HwBox_LY2), HL
    LD HL, VDP_HwBox_LX1
    LD A, (VDP_HwBox_Op)
    LD B, A
    LD A, (VDP_HwBox_Color)
    CALL VDP_HwLine
    LD HL, (VDP_HwBox_X1)
    LD (VDP_HwBox_LX1), HL
    LD HL, (VDP_HwBox_Y2)
    LD (VDP_HwBox_LY1), HL
    LD HL, (VDP_HwBox_X1)
    LD (VDP_HwBox_LX2), HL
    LD HL, (VDP_HwBox_Y1)
    LD (VDP_HwBox_LY2), HL
    LD HL, VDP_HwBox_LX1
    LD A, (VDP_HwBox_Op)
    LD B, A
    LD A, (VDP_HwBox_Color)
    CALL VDP_HwLine
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwBoxFill: retangulo cheio entre dois cantos opostos, em qualquer ordem (LMMV)
; Entrada: HL = cantos em RAM: x1(2) y1(2) x2(2) y2(2), A = cor, B = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwBoxFill:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_CLR), A
    LD A, B
    AND 0Fh
    OR VDP_CMD_LMMV
    LD (VDP_Cmd_CMD), A
    XOR A
    LD (VDP_Cmd_ARG), A
    LD DE, VDP_HwBox_X1
    LD BC, 0008h
    LDIR
    LD HL, (VDP_HwBox_X1)
    LD (VDP_Cmd_DX), HL
    LD HL, (VDP_HwBox_Y1)
    LD (VDP_Cmd_DY), HL
    LD HL, (VDP_HwBox_X2)
    LD DE, (VDP_HwBox_X1)
    CALL VDP_HwSpan
    INC HL
    LD (VDP_Cmd_NX), HL
    JR NC, VDP_HwBoxFill_Dy
    LD A, (VDP_Cmd_ARG)
    OR VDP_ARG_DIX
    LD (VDP_Cmd_ARG), A
VDP_HwBoxFill_Dy:
    LD HL, (VDP_HwBox_Y2)
    LD DE, (VDP_HwBox_Y1)
    CALL VDP_HwSpan
    INC HL
    LD (VDP_Cmd_NY), HL
    JR NC, VDP_HwBoxFill_Go
    LD A, (VDP_Cmd_ARG)
    OR VDP_ARG_DIY
    LD (VDP_Cmd_ARG), A
VDP_HwBoxFill_Go:
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET
VDP_HwBox_X1:
    DS 2
VDP_HwBox_Y1:
    DS 2
VDP_HwBox_X2:
    DS 2
VDP_HwBox_Y2:
    DS 2
VDP_HwBox_LX1:
    DS 2
VDP_HwBox_LY1:
    DS 2
VDP_HwBox_LX2:
    DS 2
VDP_HwBox_LY2:
    DS 2
VDP_HwBox_Color:
    DS 1
VDP_HwBox_Op:
    DS 1

ENDMOD
