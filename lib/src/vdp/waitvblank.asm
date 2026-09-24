; =============================================================================
; KIZUNA MSXLIB - vdp/waitvblank
; espera o proximo retraco vertical
; =============================================================================

MODULE vdp_waitvblank
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_WaitVBlank, VDP_WaitFrames

; A BIOS incrementa o contador JIFFY (FC9Eh) a cada interrupcao de vblank (50/60 Hz).
; Esperar pelo JIFFY e mais seguro que ler o flag F do status: a propria BIOS ja o
; le e apaga dentro da interrupcao.

; VDP_WaitVBlank: espera o proximo vblank. Habilita as interrupcoes (EI) -- se elas
; ficarem desligadas, o laco nunca termina.
; Preserva: A, BC, DE, HL. Destrói: flags.
VDP_WaitVBlank:
    PUSH AF
    PUSH BC
    EI
    LD A, (0FC9Eh)
    LD B, A
VDP_WaitVBlank_Loop:
    LD A, (0FC9Eh)
    CP B
    JR Z, VDP_WaitVBlank_Loop
    POP BC
    POP AF
    RET

; VDP_WaitFrames: espera B quadros (0 nao espera)
; Entrada: B = quantidade de quadros
; Preserva: A, BC, DE, HL. Destrói: flags.
VDP_WaitFrames:
    PUSH BC
    INC B
VDP_WaitFrames_Loop:
    DEC B
    JR Z, VDP_WaitFrames_Done
    CALL VDP_WaitVBlank
    JR VDP_WaitFrames_Loop
VDP_WaitFrames_Done:
    POP BC
    RET

ENDMOD
