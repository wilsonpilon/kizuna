; =============================================================================
; KIZUNA MSXLIB - vdp/initscreen2
; tabelas VRAM do SCREEN 2
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE vdp_initscreen2
BANK 0

PUBLIC VDP_InitScreen2_Tables
EXTERN VDP_FillVRAM, VDP_WriteReg
INCLUDE "../../inc/vdp.inc"

; -----------------------------------------------------------------------------
; VDP_InitScreen2_Tables: Inicializa tabelas VRAM essenciais para SCREEN 2:
; 1. Pattern Name Table (1800h..1AFFh): 3 blocos de 00h..FFh (768 bytes)
; 2. Pattern Generator Table (0000h..17FFh): limpa com 00h (6144 bytes)
; 3. Color Table (2000h..37FFh): preenche com 0F1h (Frente Branco, Fundo Preto)
; 4. Sprite Attribute Table (1B00h): desativa sprites colocando Y=208
; -----------------------------------------------------------------------------
VDP_InitScreen2_Tables:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    ; Reaplica a configuração de SCREEN 2 diretamente no VDP.
    ; A MSXgl usa estes mesmos endereços: NT=1800h, CT=2000h, PT=0000h.
    LD B, 02h
    LD C, 00h
    CALL VDP_WriteReg
    LD B, 0E2h
    LD C, 01h
    CALL VDP_WriteReg
    LD B, 06h
    LD C, 02h
    CALL VDP_WriteReg
    LD B, 0FFh
    LD C, 03h
    CALL VDP_WriteReg
    LD B, 03h
    LD C, 04h
    CALL VDP_WriteReg
    LD B, 36h
    LD C, 05h
    CALL VDP_WriteReg
    LD B, 07h
    LD C, 06h
    CALL VDP_WriteReg
    LD B, 0F1h
    LD C, 07h
    CALL VDP_WriteReg

    ; 1. Inicializa a Pattern Name Table com 00h..FFh em cada faixa.
    LD HL, 1800h
    DI
    LD A, L
    OUT (VDP_CMD), A
    NOP
    NOP
    LD A, H
    AND 3Fh
    OR 40h
    OUT (VDP_CMD), A
    LD D, 03h
VDP_Init2_BlockLoop:
    XOR A
VDP_Init2_ByteLoop:
    OUT (VDP_DATA), A
    NOP
    NOP
    NOP
    NOP
    INC A
    JR NZ, VDP_Init2_ByteLoop
    DEC D
    JR NZ, VDP_Init2_BlockLoop
    EI

    ; 2. Limpa Pattern Generator Table (0000h..17FFh, 6144 bytes) com 00h
    LD HL, 0000h
    LD BC, 1800h
    XOR A
    CALL VDP_FillVRAM

    ; 3. Preenche Color Table (2000h..37FFh, 6144 bytes) com 0F1h (Frente 15 Branco, Fundo 1 Preto)
    LD HL, 2000h
    LD BC, 1800h
    LD A, 0F1h
    CALL VDP_FillVRAM

    ; 4. Desativa todos os sprites (Y=208 em toda a SAT)
    LD HL, 1B00h
    LD BC, 0080h
    LD A, 0D0h
    CALL VDP_FillVRAM

    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
