; =============================================================================
; KIZUNA MSXLIB - vdp/cmdrun
; dispara um comando do motor a partir de um bloco de 15 bytes
; =============================================================================

MODULE vdp_cmdrun
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_CmdRun
EXTERN VDP_CmdWait

; VDP_CmdRun: espera o comando anterior acabar, grava os 15 bytes em R#32..R#46 (pela
; escrita indireta, R#17 com auto-incremento) e assim dispara o novo comando. Nao
; espera o novo comando terminar: use VDP_CmdWait antes de ler resultados ou de mexer
; na VRAM com a CPU.
; Entrada: HL = bloco: SX(2) SY(2) DX(2) DY(2) NX(2) NY(2) CLR ARG CMD
; Preserva: A, BC, HL (DE nao e usado). Destrói: flags. Reabilita as interrupcoes (EI).
VDP_CmdRun:
    PUSH AF
    PUSH BC
    PUSH HL
    CALL VDP_CmdWait
    DI
    LD A, VDP_REG_SXL
    OUT (PORT_VDP_CMD), A
    LD A, 91h           ; R#17 = 32, com auto-incremento
    OUT (PORT_VDP_CMD), A
    LD C, PORT_VDP_IREG
    LD B, 15
    OTIR
    EI
    POP HL
    POP BC
    POP AF
    RET

ENDMOD
