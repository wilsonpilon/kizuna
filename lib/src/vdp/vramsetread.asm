; =============================================================================
; KIZUNA MSXLIB - vdp/vramsetread
; ajusta o ponteiro de leitura da VRAM (17 bits)
; =============================================================================

MODULE vdp_vramsetread
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramSetRead

; VDP_VramSetRead: proximo byte lido por VDP_VramGet/ReadStream vem deste endereco
; de 17 bits (0..1FFFFh); o ponteiro avanca sozinho.
; Entrada: A = bit 16 do endereco (0 ou 1), HL = os 16 bits baixos
; Preserva: BC, DE, HL, A. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VramSetRead:
    PUSH AF
    PUSH BC
    DI
    AND 01h
    ADD A, A
    ADD A, A
    LD B, A
    LD A, H
    RLCA
    RLCA
    AND 03h
    OR B
    OUT (PORT_VDP_CMD), A
    LD A, VDP_REG_VRAMHI + 80h
    OUT (PORT_VDP_CMD), A
    LD A, L
    OUT (PORT_VDP_CMD), A
    LD A, H
    AND 3Fh             ; bit 6 = 0: leitura
    OUT (PORT_VDP_CMD), A
    EI
    POP BC
    POP AF
    RET

ENDMOD
