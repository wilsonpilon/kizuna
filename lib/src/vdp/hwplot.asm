; =============================================================================
; KIZUNA MSXLIB - vdp/hwplot
; ponto pelo motor de comandos (PSET) e leitura de ponto (POINT)
; =============================================================================

MODULE vdp_hwplot
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwPlot, VDP_HwPoint
EXTERN VDP_CmdRun, VDP_CmdWait, VDP_ReadStatus
EXTERN VDP_Cmd_SX, VDP_Cmd_SY, VDP_Cmd_DX, VDP_Cmd_DY, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; Os comandos do motor so valem nos modos bitmap (SCREEN 5 a 8).

; VDP_HwPlot: desenha um ponto (comando PSET)
; Entrada: BC = X, DE = Y, A = cor (no formato do modo), H = operacao logica (VDP_LOP_*:
;          0 = copia, 1 AND, 2 OR, 3 XOR, 4 NOT; +8 = transparente para a cor 0)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwPlot:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_DX), BC
    LD (VDP_Cmd_DY), DE
    LD (VDP_Cmd_CLR), A
    XOR A
    LD (VDP_Cmd_ARG), A
    LD A, H
    AND 0Fh
    OR VDP_CMD_PSET
    LD (VDP_Cmd_CMD), A
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwPoint: le a cor de um ponto (comando POINT; espera o resultado)
; Entrada: BC = X, DE = Y
; Saída: A = cor
; Preserva: BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_HwPoint:
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Cmd_SX), BC
    LD (VDP_Cmd_SY), DE
    XOR A
    LD (VDP_Cmd_ARG), A
    LD A, VDP_CMD_POINT
    LD (VDP_Cmd_CMD), A
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    CALL VDP_CmdWait
    LD A, 07h
    CALL VDP_ReadStatus
    POP HL
    POP DE
    POP BC
    RET

ENDMOD
