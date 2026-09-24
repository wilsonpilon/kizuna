; =============================================================================
; KIZUNA MSXLIB - vdp/hwtransfer
; transfere pixels entre a RAM e a VRAM pelo motor de comandos (LMMC, HMMC, LMCM)
; =============================================================================

MODULE vdp_hwtransfer
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_HwLoadRect, VDP_HwLoadFast, VDP_HwReadRect
EXTERN VDP_CmdRun, VDP_Cmd_SX, VDP_Cmd_SY, VDP_Cmd_DX, VDP_Cmd_NX, VDP_Cmd_CLR, VDP_Cmd_ARG, VDP_Cmd_CMD

; Estas rotinas falam com o motor ponto a ponto: o primeiro dado vai junto com o comando
; (R#44) e cada dado seguinte so e enviado/lido quando S#2 mostra TR = 1. Rodam com as
; interrupcoes desligadas do inicio ao fim (S#2 fica selecionado em R#15 durante a
; transferencia) e reabilitam (EI) no fim.

; VDP_HwLoadRect: RAM -> VRAM, um byte de dados por PONTO, com operacao logica (LMMC)
; Entrada: HL = retangulo em RAM: X(2) Y(2) largura(2) altura(2), DE = dados na RAM,
;          BC = quantidade de bytes (largura * altura), A = operacao logica
; Preserva: A, BC, DE, HL. Destrói: flags.
VDP_HwLoadRect:
    PUSH AF
    AND 0Fh
    OR VDP_CMD_LMMC
    LD (VDP_Cmd_CMD), A
    POP AF
    JR VDP_HwLoad_Go

; VDP_HwLoadFast: RAM -> VRAM, um byte de dados por BYTE de VRAM (HMMC): SCREEN 5 leva 2
; pontos por byte, SCREEN 6 leva 4, SCREEN 7 leva 2, SCREEN 8 leva 1; X e largura tem que
; ser multiplos disso. Sem operacao logica.
; Entrada: HL = retangulo em RAM: X(2) Y(2) largura(2) altura(2), DE = dados na RAM,
;          BC = quantidade de bytes
; Preserva: A, BC, DE, HL. Destrói: flags.
VDP_HwLoadFast:
    PUSH AF
    LD A, VDP_CMD_HMMC
    LD (VDP_Cmd_CMD), A
    POP AF
VDP_HwLoad_Go:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD A, B
    OR C
    JR Z, VDP_HwLoad_Done
    PUSH DE
    PUSH BC
    XOR A
    LD (VDP_Cmd_ARG), A
    LD DE, VDP_Cmd_DX
    LD BC, 0008h
    LDIR
    POP BC
    POP DE
    LD A, (DE)
    LD (VDP_Cmd_CLR), A
    INC DE
    DEC BC
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    LD A, B
    OR C
    JR Z, VDP_HwLoad_Done
    DI
    LD A, 0ACh          ; R#17 = 44 sem auto-incremento: cada OUT (9Bh) escreve em R#44
    OUT (PORT_VDP_CMD), A
    LD A, 91h
    OUT (PORT_VDP_CMD), A
    LD A, 02h
    OUT (PORT_VDP_CMD), A
    LD A, 8Fh           ; R#15 = 2: le S#2
    OUT (PORT_VDP_CMD), A
VDP_HwLoad_Loop:
    IN A, (PORT_VDP_CMD)
    BIT 7, A
    JR NZ, VDP_HwLoad_Send
    RRCA
    JR C, VDP_HwLoad_Loop
    JR VDP_HwLoad_Restore
VDP_HwLoad_Send:
    LD A, (DE)
    OUT (PORT_VDP_IREG), A
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_HwLoad_Loop
VDP_HwLoad_Restore:
    XOR A
    OUT (PORT_VDP_CMD), A
    LD A, 8Fh           ; R#15 = 0
    OUT (PORT_VDP_CMD), A
VDP_HwLoad_Done:
    EI
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_HwReadRect: VRAM -> RAM, um byte por PONTO (LMCM); a cor vem no formato do modo
; Entrada: HL = retangulo em RAM: X(2) Y(2) largura(2) altura(2), DE = destino na RAM,
;          BC = quantidade de bytes (largura * altura)
; Preserva: A, BC, DE, HL. Destrói: flags.
VDP_HwReadRect:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD A, B
    OR C
    JR Z, VDP_HwRead_Done
    PUSH DE
    PUSH BC
    XOR A
    LD (VDP_Cmd_ARG), A
    LD A, VDP_CMD_LMCM
    LD (VDP_Cmd_CMD), A
    LD DE, VDP_Cmd_SX
    LD BC, 0004h
    LDIR                ; X, Y -> SX, SY
    LD DE, VDP_Cmd_NX
    LD C, 04h
    LDIR                ; largura, altura -> NX, NY
    POP BC
    POP DE
    LD HL, VDP_Cmd_SX
    CALL VDP_CmdRun
    DI
VDP_HwRead_Loop:
    LD A, 02h
    OUT (PORT_VDP_CMD), A
    LD A, 8Fh           ; R#15 = 2
    OUT (PORT_VDP_CMD), A
    IN A, (PORT_VDP_CMD)
    BIT 7, A
    JR NZ, VDP_HwRead_Get
    RRCA
    JR C, VDP_HwRead_Loop
    JR VDP_HwRead_Restore
VDP_HwRead_Get:
    LD A, 07h
    OUT (PORT_VDP_CMD), A
    LD A, 8Fh           ; R#15 = 7
    OUT (PORT_VDP_CMD), A
    IN A, (PORT_VDP_CMD)
    LD (DE), A
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, VDP_HwRead_Loop
VDP_HwRead_Restore:
    XOR A
    OUT (PORT_VDP_CMD), A
    LD A, 8Fh           ; R#15 = 0
    OUT (PORT_VDP_CMD), A
    LD A, 00h
    OUT (PORT_VDP_CMD), A
    LD A, 0AEh          ; R#46 = 0: cancela o que sobrar (se a quantidade pedida foi menor que a area)
    OUT (PORT_VDP_CMD), A
VDP_HwRead_Done:
    EI
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
