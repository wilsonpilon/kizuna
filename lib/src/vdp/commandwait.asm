; =============================================================================
; KIZUNA MSXLIB - vdp/commandwait
; aguarda o motor de comando do V9938/V9958
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_commandwait
BANK 0

PUBLIC VDP_CommandWait_Raw
EXTERN VDP_WriteReg_Raw
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_CommandWait_Raw: aguarda o motor de comando do V9938/V9958 ficar
; livre (bit CE=0 do registrador de status S#2). Sem DI/EI proprios.
; -----------------------------------------------------------------------------
VDP_CommandWait_Raw:
    PUSH AF
    PUSH BC
    LD B, 02h
    LD C, 0Fh
    CALL VDP_WriteReg_Raw      ; R#15 = 2 (seleciona S#2 para leitura)
VDP_CommandWait_Loop:
    IN A, (VDP_CMD)
    AND 01h                    ; bit0 = CE (Command Executing)
    JR NZ, VDP_CommandWait_Loop
    LD B, 00h
    LD C, 0Fh
    CALL VDP_WriteReg_Raw      ; restaura R#15 = 0 (leitura de S#0 padrao)
    POP BC
    POP AF
    RET

ENDMOD
