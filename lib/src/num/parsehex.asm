; =============================================================================
; KIZUNA MSXLIB - num/parsehex
; leitura de digitos hexadecimais (nucleo)
; =============================================================================

MODULE num_parsehex
BANK 0

PUBLIC NUM_ParseHex16
EXTERN CHAR_DigitValue

; NUM_ParseHex16: le digitos hexadecimais (0-9, A-F, a-f) a partir de BC
; Entrada: BC = ponteiro para o primeiro digito, DE = fim do texto (0 = sem limite)
; Saída: HL = valor, BC = ponteiro depois do ultimo digito; carry = 1 se nao havia
;        digito ou o valor passou de 65535 (HL = 0 nesse caso)
; Destrói: A, DE, flags.
NUM_ParseHex16:
    PUSH BC
    PUSH DE
    LD HL, 0000h
NUM_ParseHex16_Loop:
    POP DE
    PUSH DE
    LD A, D
    OR E
    JR Z, NUM_ParseHex16_Look
    LD A, B
    CP D
    JR NZ, NUM_ParseHex16_Look
    LD A, C
    CP E
    JR Z, NUM_ParseHex16_Done
NUM_ParseHex16_Look:
    LD A, (BC)
    CALL CHAR_DigitValue
    CP 0FFh
    JR Z, NUM_ParseHex16_Done
    LD D, A             ; D = digito
    LD A, H
    AND 0F0h
    JR NZ, NUM_ParseHex16_Ovf   ; o proximo deslocamento perderia bits
    ADD HL, HL
    ADD HL, HL
    ADD HL, HL
    ADD HL, HL
    LD A, L
    OR D
    LD L, A
    INC BC
    JR NUM_ParseHex16_Loop
NUM_ParseHex16_Ovf:
    POP DE
    POP DE
    LD HL, 0000h
    SCF
    RET
NUM_ParseHex16_Done:
    POP DE
    POP DE              ; DE = inicio
    LD A, B
    CP D
    JR NZ, NUM_ParseHex16_Ok
    LD A, C
    CP E
    JR Z, NUM_ParseHex16_None
NUM_ParseHex16_Ok:
    OR A
    RET
NUM_ParseHex16_None:
    LD HL, 0000h
    SCF
    RET

ENDMOD
