; =============================================================================
; KIZUNA MSXLIB - PSG.ASM
; Rotinas de controle do gerador de som PSG (AY-3-8910 / YM2149)
; =============================================================================

MODULE PSG
BANK 0

PUBLIC PSG_Write, PSG_Read, PSG_MuteAll, PSG_PlayTone

PSG_REG_SEL EQU 00A0h
PSG_DATA_WR EQU 00A1h
PSG_DATA_RD EQU 00A2h

; -----------------------------------------------------------------------------
; PSG_Write: Escreve um valor em um registrador do PSG
; Entrada: A = número do registrador (0..15), E = valor (8 bits)
; -----------------------------------------------------------------------------
PSG_Write:
    OUT (PSG_REG_SEL), A
    LD A, E
    OUT (PSG_DATA_WR), A
    RET

; -----------------------------------------------------------------------------
; PSG_Read: Lê o valor atual de um registrador do PSG
; Entrada: A = número do registrador (0..15)
; Saída: A = valor lido
; -----------------------------------------------------------------------------
PSG_Read:
    OUT (PSG_REG_SEL), A
    IN A, (PSG_DATA_RD)
    RET

; -----------------------------------------------------------------------------
; PSG_MuteAll: Zera o volume dos 3 canais de som e silencia o PSG
; -----------------------------------------------------------------------------
PSG_MuteAll:
    PUSH AF
    PUSH DE
    LD E, 00h
    LD A, 08h ; Canal A volume
    CALL PSG_Write
    LD A, 09h ; Canal B volume
    CALL PSG_Write
    LD A, 0Ah ; Canal C volume
    CALL PSG_Write
    ; Desativa tons no mixer de forma segura para portas de I/O do MSX (0xBF)
    LD A, 07h
    LD E, 0BFh
    CALL PSG_Write
    POP DE
    POP AF
    RET

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
