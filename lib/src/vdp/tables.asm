; =============================================================================
; KIZUNA MSXLIB - vdp/tables
; enderecos das tabelas de VRAM (nomes, padroes, cores, sprites)
; =============================================================================

MODULE vdp_tables
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetNameTable, VDP_SetPatternTable, VDP_SetColorTable
PUBLIC VDP_SetSpriteAttrTable, VDP_SetSpritePatternTable
EXTERN VDP_SetReg, VDP_AddrShr17, VDP_Mode

; Cada rotina recebe o endereco de 17 bits em A:HL (A = bit 16, HL = os 16 bits
; baixos) e escreve o(s) registrador(es) da tabela, com a copia sombra. O endereco
; tem que ser multiplo do tamanho de alinhamento da tabela (1 KB para nomes, 2 KB
; para padroes, 64 bytes para cores, 128 bytes para atributos de sprite, 2 KB
; para padroes de sprite); os bits baixos que sobrarem sao descartados.
; Os bits "fixos em 1" que o V9938 exige em alguns modos (SCREEN 2/4, bitmap, texto
; de 80 colunas, sprites modo 2) sao acrescentados de acordo com o modo ajustado
; por VDP_SetMode; sem VDP_SetMode antes, os valores saem sem esses bits.
; Preservam: A, BC, DE, HL. Destroem: flags. Reabilitam as interrupcoes (EI).

; interno: A = bit 16, HL = endereco, B = deslocamento, C = registrador baixo,
; D = bits a ligar no valor baixo, E = registrador alto (0 = nao tem)
VDP_SetTable_Do:
    PUSH DE
    CALL VDP_AddrShr17
    LD A, L
    OR D
    LD B, A
    CALL VDP_SetReg
    LD A, E
    OR A
    JR Z, VDP_SetTable_Done
    LD C, A
    LD B, H
    CALL VDP_SetReg
VDP_SetTable_Done:
    POP DE
    RET

; VDP_SetNameTable: tabela de nomes (R#2)
VDP_SetNameTable:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD A, (VDP_Mode)
    LD B, 10
    LD D, 00h
    CP 0Ah
    JR NC, VDP_SetNameTable_Go
    CP 09h
    JR NZ, VDP_SetNameTable_N1
    LD D, 03h           ; texto de 80 colunas: A11-A10 valem 1
    JR VDP_SetNameTable_Go
VDP_SetNameTable_N1:
    CP 05h
    JR C, VDP_SetNameTable_Go
    LD D, 1Fh           ; bitmap: os 5 bits baixos valem 1
    CP 07h
    JR C, VDP_SetNameTable_Go
    LD B, 11            ; SCREEN 7 e 8: o bit 16 fica no bit 5
VDP_SetNameTable_Go:
    LD A, E
    LD E, 00h
    LD C, VDP_REG_NAME
    CALL VDP_SetTable_Do
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SetPatternTable: tabela de padroes/geradora de caracteres (R#4)
VDP_SetPatternTable:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD A, (VDP_Mode)
    LD D, 00h
    CP 02h
    JR Z, VDP_SetPatternTable_G2
    CP 04h
    JR NZ, VDP_SetPatternTable_Go
VDP_SetPatternTable_G2:
    LD D, 03h           ; SCREEN 2 e 4: os 2 bits baixos sao mascara
VDP_SetPatternTable_Go:
    LD A, E
    LD E, 00h
    LD B, 11
    LD C, VDP_REG_PATTERN
    CALL VDP_SetTable_Do
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SetColorTable: tabela de cores (R#3 e R#10); no texto de 80 colunas e a tabela de piscar
VDP_SetColorTable:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD A, (VDP_Mode)
    LD D, 00h
    CP 02h
    JR Z, VDP_SetColorTable_G2
    CP 04h
    JR Z, VDP_SetColorTable_G2
    CP 09h
    JR NZ, VDP_SetColorTable_Go
    LD D, 07h           ; texto de 80 colunas: os 3 bits baixos valem 1
    JR VDP_SetColorTable_Go
VDP_SetColorTable_G2:
    LD D, 7Fh           ; SCREEN 2 e 4: os 7 bits baixos sao mascara
VDP_SetColorTable_Go:
    LD A, E
    LD E, VDP_REG_COLORH
    LD B, 6
    LD C, VDP_REG_COLOR
    CALL VDP_SetTable_Do
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SetSpriteAttrTable: tabela de atributos de sprite (R#5 e R#11). No modo 2 de
; sprites (SCREEN 4 a 8) a tabela de cores dos sprites fica 512 bytes antes dela.
VDP_SetSpriteAttrTable:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD E, A
    LD A, (VDP_Mode)
    LD D, 00h
    CP 04h
    JR C, VDP_SetSpriteAttrTable_Go
    CP 09h
    JR NC, VDP_SetSpriteAttrTable_Go
    LD D, 03h           ; sprites modo 2: os 2 bits baixos valem 1
VDP_SetSpriteAttrTable_Go:
    LD A, E
    LD E, VDP_REG_SPRATTRH
    LD B, 7
    LD C, VDP_REG_SPRATTR
    CALL VDP_SetTable_Do
    POP HL
    POP DE
    POP BC
    POP AF
    RET

; VDP_SetSpritePatternTable: tabela de padroes de sprite (R#6)
VDP_SetSpritePatternTable:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD D, 00h
    LD E, 00h
    LD B, 11
    LD C, VDP_REG_SPRPAT
    CALL VDP_SetTable_Do
    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
