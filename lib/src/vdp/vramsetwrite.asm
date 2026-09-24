; =============================================================================
; KIZUNA MSXLIB - vdp/vramsetwrite
; ajusta o ponteiro de escrita da VRAM (17 bits)
; =============================================================================

MODULE vdp_vramsetwrite
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_VramSetWrite

; VDP_VramSetWrite: proximo byte escrito em VDP_VramPut/WriteStream vai para este
; endereco de 17 bits (0..1FFFFh). O ponteiro avanca sozinho a cada byte, inclusive
; atravessando os limites de 16 KB.
; Entrada: A = bit 16 do endereco (0 ou 1), HL = os 16 bits baixos
; R#14 e escrito direto, sem passar pela copia sombra (o VDP o altera sozinho quando o
; ponteiro atravessa 16 KB, entao a copia nao valeria mesmo).
; Preserva: BC, DE, HL, A. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_VramSetWrite:
    PUSH AF
    PUSH BC
    DI
    AND 01h
    ADD A, A
    ADD A, A            ; A16 vai para o bit 2 de R#14
    LD B, A
    LD A, H
    RLCA
    RLCA
    AND 03h             ; A15 e A14
    OR B
    OUT (PORT_VDP_CMD), A       ; R#14 = A16..A14
    LD A, VDP_REG_VRAMHI + 80h
    OUT (PORT_VDP_CMD), A
    LD A, L
    OUT (PORT_VDP_CMD), A
    LD A, H
    AND 3Fh
    OR 40h              ; bit 6 = escrita
    OUT (PORT_VDP_CMD), A
    EI
    POP BC
    POP AF
    RET

ENDMOD
