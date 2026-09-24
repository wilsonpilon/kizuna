; =============================================================================
; KIZUNA MSXLIB - num/parseu16
; leitura de digitos decimais (nucleo)
; =============================================================================

MODULE num_parseu16
BANK 0

PUBLIC NUM_ParseU16

; NUM_ParseU16: le digitos decimais a partir de BC ate o primeiro caractere que
; nao seja digito (ou ate o fim, se DE nao for 0).
; Entrada: BC = ponteiro para o primeiro digito, DE = fim do texto (0 = sem limite)
; Saída: HL = valor, BC = ponteiro depois do ultimo digito; carry = 1 se nao havia
;        nenhum digito ou o valor passou de 65535 (HL = 0 nesse caso)
; Destrói: A, DE, flags.
NUM_ParseU16:
    PUSH BC             ; inicio: para saber se leu algum digito
    PUSH DE             ; fim do texto
    LD HL, 0000h
NUM_ParseU16_Loop:
    POP DE
    PUSH DE             ; DE = fim
    LD A, D
    OR E
    JR Z, NUM_ParseU16_Look
    LD A, B
    CP D
    JR NZ, NUM_ParseU16_Look
    LD A, C
    CP E
    JR Z, NUM_ParseU16_Done     ; chegou ao fim do texto
NUM_ParseU16_Look:
    LD A, (BC)
    SUB 30h
    JR C, NUM_ParseU16_Done
    CP 0Ah
    JR NC, NUM_ParseU16_Done
    LD E, A
    LD D, 00h
    PUSH DE             ; digito
    ADD HL, HL          ; valor * 2
    JR C, NUM_ParseU16_OvfD
    LD D, H
    LD E, L
    ADD HL, HL          ; * 4
    JR C, NUM_ParseU16_OvfD
    ADD HL, HL          ; * 8
    JR C, NUM_ParseU16_OvfD
    ADD HL, DE          ; * 10
    JR C, NUM_ParseU16_OvfD
    POP DE
    ADD HL, DE          ; + digito
    JR C, NUM_ParseU16_Ovf
    INC BC
    JR NUM_ParseU16_Loop
NUM_ParseU16_OvfD:
    POP DE              ; descarta o digito
NUM_ParseU16_Ovf:
    POP DE              ; descarta o fim
    POP DE              ; descarta o inicio
    LD HL, 0000h
    SCF
    RET
NUM_ParseU16_Done:
    POP DE              ; descarta o fim
    POP DE              ; DE = inicio
    LD A, B
    CP D
    JR NZ, NUM_ParseU16_Ok
    LD A, C
    CP E
    JR Z, NUM_ParseU16_None     ; BC nao andou: nenhum digito
NUM_ParseU16_Ok:
    OR A                ; carry = 0
    RET
NUM_ParseU16_None:
    LD HL, 0000h
    SCF
    RET

ENDMOD
