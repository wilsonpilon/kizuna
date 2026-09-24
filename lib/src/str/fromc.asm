; =============================================================================
; KIZUNA MSXLIB - str/fromc
; converte string terminada em zero para tamanho+dados
; =============================================================================

MODULE str_fromc
BANK 0

PUBLIC STR_FromC

; STR_FromC: copia a string terminada em zero de HL para DE no formato
; tamanho+dados. Trunca em 255 caracteres.
; Entrada: HL = string terminada em zero, DE = destino (ate 256 bytes)
; Preserva: DE. Destrói: A, BC, HL, flags.
STR_FromC:
    PUSH DE             ; endereco do byte de tamanho
    INC DE
    LD B, 00h           ; B = quantos ja copiados
STR_FromC_Loop:
    LD A, B
    CP 0FFh
    JR Z, STR_FromC_End
    LD A, (HL)
    OR A
    JR Z, STR_FromC_End
    LD (DE), A
    INC HL
    INC DE
    INC B
    JR STR_FromC_Loop
STR_FromC_End:
    POP DE
    LD A, B
    LD (DE), A          ; grava o tamanho
    RET

ENDMOD
