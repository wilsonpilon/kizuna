; =============================================================================
; KIZUNA MSXLIB - vdp/gettables
; le de volta os enderecos das tabelas de VRAM (a partir da copia sombra)
; =============================================================================

MODULE vdp_gettables
BANK 0

INCLUDE "../../inc/ports.inc"
INCLUDE "../../inc/vdpreg.inc"

PUBLIC VDP_GetNameTable, VDP_GetPatternTable, VDP_GetColorTable
PUBLIC VDP_GetSpriteAttrTable, VDP_GetSpritePatternTable
EXTERN VDP_GetReg, VDP_AddrShl17, VDP_Mode

; Devolvem o endereco de 17 bits em A:HL (A = bit 16), calculado a partir da copia
; sombra dos registradores e do modo ajustado por VDP_SetMode (os bits de mascara de
; cada modo sao ignorados).
; Preservam: BC, DE. Destroem: A, HL, flags.

VDP_GetNameTable:
    PUSH BC
    PUSH DE
    LD A, 02h
    CALL VDP_GetReg
    LD E, A
    LD B, 10
    LD A, (VDP_Mode)
    CP 0Ah
    JR NC, VDP_GetNameTable_Go
    CP 09h
    JR Z, VDP_GetNameTable_T2
    CP 05h
    JR C, VDP_GetNameTable_Go
    CP 07h
    JR NC, VDP_GetNameTable_M78
    LD A, E
    AND 60h             ; SCREEN 5 e 6: A16 e A15
    LD E, A
    JR VDP_GetNameTable_Go
VDP_GetNameTable_M78:
    LD A, E
    AND 20h             ; SCREEN 7 e 8: so A16
    LD E, A
    LD B, 11
    JR VDP_GetNameTable_Go
VDP_GetNameTable_T2:
    LD A, E
    AND 0FCh            ; texto de 80 colunas: os 2 bits baixos valem 1
    LD E, A
VDP_GetNameTable_Go:
    LD H, 00h
    LD L, E
    XOR A
    CALL VDP_AddrShl17
    POP DE
    POP BC
    RET

VDP_GetPatternTable:
    PUSH BC
    PUSH DE
    LD A, 04h
    CALL VDP_GetReg
    LD E, A
    LD A, (VDP_Mode)
    CP 02h
    JR Z, VDP_GetPatternTable_G2
    CP 04h
    JR NZ, VDP_GetPatternTable_Go
VDP_GetPatternTable_G2:
    LD A, E
    AND 0FCh
    LD E, A
VDP_GetPatternTable_Go:
    LD A, E
    AND 3Fh
    LD L, A
    LD H, 00h
    LD B, 11
    XOR A
    CALL VDP_AddrShl17
    POP DE
    POP BC
    RET

VDP_GetColorTable:
    PUSH BC
    PUSH DE
    LD A, 0Ah
    CALL VDP_GetReg
    AND 07h
    LD D, A
    LD A, 03h
    CALL VDP_GetReg
    LD E, A
    LD A, (VDP_Mode)
    CP 02h
    JR Z, VDP_GetColorTable_G2
    CP 04h
    JR Z, VDP_GetColorTable_G2
    CP 09h
    JR NZ, VDP_GetColorTable_Go
    LD A, E
    AND 0F8h            ; texto 80 colunas: os 3 bits baixos valem 1
    LD E, A
    JR VDP_GetColorTable_Go
VDP_GetColorTable_G2:
    LD A, E
    AND 80h             ; SCREEN 2 e 4: so A13
    LD E, A
VDP_GetColorTable_Go:
    LD H, D
    LD L, E
    LD B, 6
    XOR A
    CALL VDP_AddrShl17
    POP DE
    POP BC
    RET

VDP_GetSpriteAttrTable:
    PUSH BC
    PUSH DE
    LD A, 0Bh
    CALL VDP_GetReg
    AND 03h
    LD D, A
    LD A, 05h
    CALL VDP_GetReg
    LD E, A
    LD A, (VDP_Mode)
    CP 04h
    JR C, VDP_GetSpriteAttrTable_Go
    CP 09h
    JR NC, VDP_GetSpriteAttrTable_Go
    LD A, E
    AND 0FCh            ; sprites modo 2: os 2 bits baixos valem 1
    LD E, A
VDP_GetSpriteAttrTable_Go:
    LD H, D
    LD L, E
    LD B, 7
    XOR A
    CALL VDP_AddrShl17
    POP DE
    POP BC
    RET

VDP_GetSpritePatternTable:
    PUSH BC
    PUSH DE
    LD A, 06h
    CALL VDP_GetReg
    AND 3Fh
    LD L, A
    LD H, 00h
    LD B, 11
    XOR A
    CALL VDP_AddrShl17
    POP DE
    POP BC
    RET

ENDMOD
