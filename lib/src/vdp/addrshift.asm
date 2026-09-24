; =============================================================================
; KIZUNA MSXLIB - vdp/addrshift
; desloca um endereco de VRAM de 17 bits (usado pelas rotinas de tabela)
; =============================================================================

MODULE vdp_addrshift
BANK 0


PUBLIC VDP_AddrShr17, VDP_AddrShl17

; Um endereco de VRAM de 17 bits viaja em A:HL -- o bit 16 e o bit 0 de A, os 16
; bits baixos ficam em HL (o mesmo formato de VDP_VramSetWrite).

; VDP_AddrShr17: A:HL >> B
; Entrada: A (so o bit 0 vale), HL, B = quantidade de posicoes (0 nao muda nada)
; Saída: A (0 ou 1), HL. Preserva: BC, DE. Destrói: flags.
VDP_AddrShr17:
    PUSH BC
    AND 01h
    INC B
    DEC B
    JR Z, VDP_AddrShr17_Done
VDP_AddrShr17_Loop:
    SRL A
    RR H
    RR L
    DJNZ VDP_AddrShr17_Loop
VDP_AddrShr17_Done:
    POP BC
    RET

; VDP_AddrShl17: A:HL << B (o que passa de 17 bits se perde)
; Entrada: A (so o bit 0 vale), HL, B = quantidade de posicoes
; Saída: A (0 ou 1), HL. Preserva: BC, DE. Destrói: flags.
VDP_AddrShl17:
    PUSH BC
    AND 01h
    INC B
    DEC B
    JR Z, VDP_AddrShl17_Done
VDP_AddrShl17_Loop:
    ADD HL, HL
    RLA
    DJNZ VDP_AddrShl17_Loop
VDP_AddrShl17_Done:
    AND 01h
    POP BC
    RET

ENDMOD
