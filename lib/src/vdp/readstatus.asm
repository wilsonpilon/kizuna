; =============================================================================
; KIZUNA MSXLIB - vdp/readstatus
; le um registrador de status do VDP
; =============================================================================

MODULE vdp_readstatus
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_ReadStatus

; VDP_ReadStatus: le o registrador de status S#n (0..9) e deixa R#15 em 0, o que
; o resto do sistema espera. Ler S#0 apaga o flag de vblank (F) e o de colisao (C).
; Entrada: A = numero do status (0..9)
; Saída: A = valor lido
; Preserva: BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_ReadStatus:
    PUSH BC
    DI
    LD B, A
    OUT (PORT_VDP_CMD), A       ; R#15 = n
    LD A, 8Fh
    OUT (PORT_VDP_CMD), A
    NOP
    NOP
    IN A, (PORT_VDP_CMD)
    LD C, A
    XOR A
    OUT (PORT_VDP_CMD), A       ; R#15 = 0
    LD A, 8Fh
    OUT (PORT_VDP_CMD), A
    EI
    LD A, C
    POP BC
    RET

ENDMOD
