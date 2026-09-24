; =============================================================================
; KIZUNA MSXLIB - vdp/setmode
; escolhe o modo de tela ajustando os registradores do VDP
; =============================================================================

MODULE vdp_setmode
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_SetMode, VDP_GetMode
EXTERN VDP_UpdateReg, VDP_SetReg, VDP_Mode

; VDP_SetMode: ajusta os registradores do VDP para um modo de tela, com as tabelas
; de VRAM nos enderecos padrao do MSX-BASIC. Escreve SO os registradores (R#0, R#1,
; R#2..R#6, R#10, R#11): nao limpa a VRAM, nao carrega fonte nem paleta, e mantem a
; tela ligada ou desligada como estava. (Para o SCREEN completo do BASIC, incluindo
; fonte, use BIOS_CHGMOD.)
;   0 = SCREEN 0 (texto 40 colunas)   1 = SCREEN 1        2 = SCREEN 2
;   3 = SCREEN 3 (multicolor)         4 = SCREEN 4        5 = SCREEN 5 (256x212, 16 cores)
;   6 = SCREEN 6 (512x212, 4 cores)   7 = SCREEN 7 (512x212, 16 cores)
;   8 = SCREEN 8 (256x212, 256 cores) 9 = SCREEN 0 com 80 colunas
; Entrada: A = modo (0..9)
; Saída: carry = 0 se ajustou, 1 se o modo nao existe (nada e escrito)
; Preserva: A, BC, DE, HL. Destrói: flags. Reabilita as interrupcoes (EI).
VDP_SetMode:
    CP 0Ah
    JR C, VDP_SetMode_Ok
    SCF
    RET
VDP_SetMode_Ok:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL
    LD (VDP_Mode), A    ; registra o modo
    LD E, A
    LD D, 00h
    LD H, D
    LD L, A
    ADD HL, HL
    ADD HL, HL
    ADD HL, HL          ; HL = modo * 8
    ADD HL, DE          ; HL = modo * 9
    LD DE, VDP_ModeTable
    ADD HL, DE          ; HL = entrada do modo
    LD C, VDP_REG_MODE0
    LD D, 0Eh           ; M3, M4, M5
    LD E, (HL)
    INC HL
    CALL VDP_UpdateReg
    LD C, VDP_REG_MODE1
    LD D, 18h           ; M1, M2
    LD E, (HL)
    INC HL
    CALL VDP_UpdateReg
    LD C, VDP_REG_NAME
VDP_SetMode_Loop:       ; R#2..R#6
    LD B, (HL)
    INC HL
    CALL VDP_SetReg
    INC C
    LD A, C
    CP VDP_REG_SPRPAT + 1
    JR NZ, VDP_SetMode_Loop
    LD C, VDP_REG_COLORH
    LD B, (HL)
    INC HL
    CALL VDP_SetReg
    LD C, VDP_REG_SPRATTRH
    LD B, (HL)
    CALL VDP_SetReg
    POP HL
    POP DE
    POP BC
    POP AF
    OR A                ; carry = 0
    RET

; VDP_GetMode: ultimo modo ajustado por VDP_SetMode (0FFh se nenhum)
; Saída: A
; Preserva: BC, DE, HL. Destrói: nada.
VDP_GetMode:
    LD A, (VDP_Mode)
    RET

; R#0 (bits de modo), R#1 (bits de modo), R#2, R#3, R#4, R#5, R#6, R#10, R#11
VDP_ModeTable:
    DB 00h, 10h, 00h, 00h, 01h, 00h, 00h, 00h, 00h   ; 0: SCREEN 0, texto de 40 colunas (Text 1)
    DB 00h, 00h, 06h, 80h, 00h, 36h, 07h, 00h, 00h   ; 1: SCREEN 1, Graphic 1
    DB 02h, 00h, 06h, FFh, 03h, 36h, 07h, 00h, 00h   ; 2: SCREEN 2, Graphic 2
    DB 00h, 08h, 02h, 00h, 00h, 36h, 07h, 00h, 00h   ; 3: SCREEN 3, Multicolor
    DB 04h, 00h, 06h, FFh, 03h, 3Fh, 07h, 00h, 00h   ; 4: SCREEN 4, Graphic 3
    DB 06h, 00h, 1Fh, 00h, 00h, EFh, 0Fh, 00h, 00h   ; 5: SCREEN 5, Graphic 4 (256x212, 16 cores)
    DB 08h, 00h, 1Fh, 00h, 00h, EFh, 0Fh, 00h, 00h   ; 6: SCREEN 6, Graphic 5 (512x212, 4 cores)
    DB 0Ah, 00h, 1Fh, 00h, 00h, F7h, 1Eh, 00h, 01h   ; 7: SCREEN 7, Graphic 6 (512x212, 16 cores)
    DB 0Eh, 00h, 1Fh, 00h, 00h, F7h, 1Eh, 00h, 01h   ; 8: SCREEN 8, Graphic 7 (256x212, 256 cores)
    DB 04h, 10h, 03h, 27h, 02h, 00h, 00h, 00h, 00h   ; 9: SCREEN 0, texto de 80 colunas (Text 2)

ENDMOD
