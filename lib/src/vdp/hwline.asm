; =============================================================================
; KIZUNA MSXLIB - vdp/hwline
; linha pelo motor de comandos (LINE)
; =============================================================================

MODULE vdp_hwline
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwLine
EXTERN VDP_CmdRun, VDP_HwSpan
EXTERN VDP_Cmd_SX, VDP_Cmd_DX, VDP_Cmd_DY, VDP_Cmd_NX, VDP_Cmd_NY, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; VDP_HwLine: linha reta de (x1,y1) a (x2,y2), os dois extremos incluidos. Calcula
; o lado maior, o lado menor e os sentidos, e dispara o comando LINE (NX = lado maior,
; NY = lado menor; o motor desenha NX+1 pontos). Coordenadas de 0 a 511 (X) e 0 a 1023 (Y).
; Entrada: HL = pontos em RAM: x1(2) y1(2) x2(2) y2(2), A = cor, B = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwLine:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_CLR), A
    LD A, B
    AND 0Fh
    OR VDP_CMD_LINE
    LD (VDP_Cmd_CMD), A
    XOR A
    LD (VDP_Cmd_ARG), A
    LD DE, VDP_HwLine_X1
    LD BC, 0008h
    LDIR
    LD HL, (VDP_HwLine_X1)
    LD (VDP_Cmd_DX), HL
    LD HL, (VDP_HwLine_Y1)
    LD (VDP_Cmd_DY), HL
    LD HL, (VDP_HwLine_X2)
    LD DE, (VDP_HwLine_X1)
    CALL VDP_HwSpan     ; HL = |dx|
    LD (VDP_Cmd_NX), HL
    JR NC, VDP_HwLine_Dy
    LD A, (VDP_Cmd_ARG)
    OR VDP_ARG_DIX
    LD (VDP_Cmd_ARG), A
VDP_HwLine_Dy:
    LD HL, (VDP_HwLine_Y2)
    LD DE, (VDP_HwLine_Y1)
    CALL VDP_HwSpan     ; HL = |dy|
    LD (VDP_Cmd_NY), HL
    JR NC, VDP_HwLine_Major
    LD A, (VDP_Cmd_ARG)
    OR VDP_ARG_DIY
    LD (VDP_Cmd_ARG), A
VDP_HwLine_Major:
    LD DE, (VDP_Cmd_NX) ; DE = |dx|, HL = |dy|
    OR A
    SBC HL, DE          ; |dy| - |dx|
    JR C, VDP_HwLine_Go ; |dy| < |dx|: lado maior em X (NX = |dx|, NY = |dy|)
    JR Z, VDP_HwLine_Go ; iguais: tanto faz
    LD HL, (VDP_Cmd_NY) ; lado maior em Y: NX = |dy|, NY = |dx|, MAJ = 1
    LD (VDP_Cmd_NX), HL
    LD (VDP_Cmd_NY), DE
    LD A, (VDP_Cmd_ARG)
    OR 01h
    LD (VDP_Cmd_ARG), A
VDP_HwLine_Go:
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET
VDP_HwLine_X1:
    DS 2
VDP_HwLine_Y1:
    DS 2
VDP_HwLine_X2:
    DS 2
VDP_HwLine_Y2:
    DS 2

ENDMOD
