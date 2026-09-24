; =============================================================================
; KIZUNA MSXLIB - vdp/spriteaddr
; enderecos dentro das tabelas de sprite (usa os registradores, vale em qualquer modo)
; =============================================================================

MODULE vdp_spriteaddr
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SpriteAttrAddr, VDP_SpriteColorAddr, VDP_SpriteModeIs2
EXTERN VDP_GetSpriteAttrTable, VDP_Mode

; VDP_SpriteModeIs2: 1 se o modo ajustado tem sprites modo 2 (SCREEN 4 a 8), 0 se e
; modo 1 (SCREEN 1 a 3) ou desconhecido
; Saída: A. Preserva: BC, DE, HL. Destrói: flags.
VDP_SpriteModeIs2:
    LD A, (VDP_Mode)
    CP 04h
    JR C, VDP_SpriteModeIs2_No
    CP 09h
    JR NC, VDP_SpriteModeIs2_No
    LD A, 01h
    RET
VDP_SpriteModeIs2_No:
    XOR A
    RET

; VDP_SpriteAttrAddr: endereco do atributo (Y, X, padrao, cor) de um sprite
; Entrada: A = indice do sprite (0..31)
; Saída: A = bit 16, HL = 16 bits baixos (formato de VDP_VramSetWrite)
; Preserva: BC, DE. Destrói: flags.
VDP_SpriteAttrAddr:
    PUSH BC
    PUSH DE
    LD C, A
    CALL VDP_GetSpriteAttrTable
    LD E, A
    LD B, 00h
    SLA C
    RL B
    SLA C
    RL B                ; BC = indice * 4
    ADD HL, BC
    LD A, E
    ADC A, 00h
    AND 01h
    POP DE
    POP BC
    RET

; VDP_SpriteColorAddr: endereco das 16 cores de linha de um sprite (so no modo 2:
; a tabela fica 512 bytes antes da de atributos, 16 bytes por sprite)
; Entrada: A = indice do sprite (0..31)
; Saída: A = bit 16, HL = 16 bits baixos
; Preserva: BC, DE. Destrói: flags.
VDP_SpriteColorAddr:
    PUSH BC
    PUSH DE
    LD C, A
    CALL VDP_GetSpriteAttrTable
    LD E, A
    LD A, H
    SUB 02h
    LD H, A
    LD A, E
    SBC A, 00h          ; A:HL = base - 512
    LD E, A
    LD B, 00h
    SLA C
    RL B
    SLA C
    RL B
    SLA C
    RL B
    SLA C
    RL B                ; BC = indice * 16
    ADD HL, BC
    LD A, E
    ADC A, 00h
    AND 01h
    POP DE
    POP BC
    RET

ENDMOD
