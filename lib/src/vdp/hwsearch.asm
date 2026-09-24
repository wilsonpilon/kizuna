; =============================================================================
; KIZUNA MSXLIB - vdp/hwsearch
; busca a borda de uma cor numa linha (SRCH)
; =============================================================================

MODULE vdp_hwsearch
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwSearch
EXTERN VDP_CmdRun, VDP_CmdWait, VDP_ReadStatus
EXTERN VDP_Cmd_SX, VDP_Cmd_SY, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; VDP_HwSearch: anda pela linha Y a partir de X procurando um ponto com a cor dada (ou,
; com EQ, um ponto DIFERENTE dela), para a direita ou para a esquerda, e devolve o X onde
; parou.
; Entrada: BC = X inicial, DE = Y, A = cor, H = argumentos: VDP_ARG_DIX (04h) = para a
;          esquerda, 02h (EQ) = pare no primeiro ponto diferente da cor
; Saída: carry = 1 se achou, 0 se chegou a borda da tela sem achar; HL = X
; Preserva: BC, DE. Destrói: A, flags. Reabilita as interrupcoes (EI).
VDP_HwSearch:
    PUSH BC
    PUSH DE
    LD (VDP_Cmd_SX), BC
    LD (VDP_Cmd_SY), DE
    LD (VDP_Cmd_CLR), A
    LD A, H
    AND 06h
    LD (VDP_Cmd_ARG), A
    LD A, VDP_CMD_SRCH
    LD (VDP_Cmd_CMD), A
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    CALL VDP_CmdWait
    LD A, 08h
    CALL VDP_ReadStatus
    LD C, A
    LD A, 09h
    CALL VDP_ReadStatus
    AND 01h
    LD H, A
    LD L, C
    LD A, 02h
    CALL VDP_ReadStatus
    AND 10h             ; BD: achou
    POP DE
    POP BC
    RET Z
    SCF
    RET

ENDMOD
