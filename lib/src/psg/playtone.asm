; =============================================================================
; KIZUNA MSXLIB - psg/playtone
; toca um tom num canal
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE psg_playtone
BANK 0

PUBLIC PSG_PlayTone
EXTERN PSG_Write
INCLUDE "../../inc/psg.inc"

; -----------------------------------------------------------------------------
; PSG_PlayTone: Configura frequência e toca tom em um canal (0, 1 ou 2)
; Entrada: A = canal (0=A, 1=B, 2=C)
;          HL = período da nota (12 bits: menor valor = frequência mais alta)
;          E = volume (0..15)
; -----------------------------------------------------------------------------
PSG_PlayTone:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    LD B, A ; B = canal (0, 1 ou 2)
    ADD A, A
    LD C, A ; C = reg fino (0, 2 ou 4)

    ; 1. Escrever parte baixa do período (8 bits)
    LD A, C
    OUT (PSG_REG_SEL), A
    LD A, L
    OUT (PSG_DATA_WR), A

    ; 2. Escrever parte alta do período (4 bits)
    INC C
    LD A, C
    OUT (PSG_REG_SEL), A
    LD A, H
    AND 0Fh
    OUT (PSG_DATA_WR), A

    ; 3. Escrever volume no canal correspondente (8 + canal)
    LD A, 08h
    ADD A, B
    OUT (PSG_REG_SEL), A
    LD A, E
    AND 0Fh
    OUT (PSG_DATA_WR), A

    ; 4. Ativar tom no misturador (Reg 7):
    ; Canal 0 -> 0xBE (Tom A), Canal 1 -> 0xBD (Tom B), Canal 2 -> 0xBB (Tom C)
    LD A, B
    LD E, 0BEh
    OR A
    JR Z, PSG_SetMixer
    LD E, 0BDh
    DEC A
    JR Z, PSG_SetMixer
    LD E, 0BBh
PSG_SetMixer:
    LD A, 07h
    CALL PSG_Write

    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
