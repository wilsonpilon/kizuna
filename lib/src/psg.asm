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
    POP DE
    POP AF
    RET

; -----------------------------------------------------------------------------
; PSG_PlayTone: Configura frequência e toca tom contínuo em um canal (0, 1 ou 2)
; Entrada: A = canal (0=A, 1=B, 2=C)
;          HL = período da nota (12 bits: menor valor = frequência mais alta)
;          E = volume (0..15)
; -----------------------------------------------------------------------------
PSG_PlayTone:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    ; Determina registrador de tom fino (canal * 2)
    LD B, A ; B = canal
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

    ; 4. Ativar canal no misturador (Reg 7): bit = 0 ativa o tom
    LD A, 07h
    CALL PSG_Read
    ; Criar máscara para zerar bit B
    LD D, A ; D = valor atual do Reg 7
    LD A, 01h
    LD E, B
    INC E
PSG_Mask_Loop:
    DEC E
    JR Z, PSG_Mask_Done
    SLA A
    JR PSG_Mask_Loop
PSG_Mask_Done:
    CPL ; Inverte máscara: 0 no canal desejado, 1 nos outros
    AND D
    LD E, A
    LD A, 07h
    CALL PSG_Write

    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
