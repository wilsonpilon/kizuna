; =============================================================================
; KIZUNA MSXLIB - vdp/cmdwait
; espera, consulta e cancela o motor de comandos
; =============================================================================

MODULE vdp_cmdwait
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_CmdWait, VDP_CmdBusy, VDP_CmdStop
EXTERN VDP_ReadStatus, VDP_SetReg

; VDP_CmdWait: espera o comando em andamento terminar (S#2, bit CE = 0)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_CmdWait:
    PUSH AF
VDP_CmdWait_Loop:
    LD A, 02h
    CALL VDP_ReadStatus
    RRCA
    JR C, VDP_CmdWait_Loop
    POP AF
    RET

; VDP_CmdBusy: 1 se ha um comando em andamento, 0 se o motor esta livre
; Saída: A. Preserva: BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_CmdBusy:
    LD A, 02h
    CALL VDP_ReadStatus
    AND 01h
    RET

; VDP_CmdStop: cancela o comando em andamento (R#46 = 0)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_CmdStop:
    PUSH AF
    PUSH BC
    LD B, 00h
    LD C, VDP_REG_CMD
    CALL VDP_SetReg
    POP BC
    POP AF
    RET

ENDMOD
