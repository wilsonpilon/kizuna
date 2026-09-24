; =============================================================================
; KIZUNA MSXLIB - mem/compare
; comparacao de dois blocos de memoria
; =============================================================================

MODULE mem_compare
BANK 0

PUBLIC MEM_Compare

; MEM_Compare: compara BC bytes de HL com os de DE, como bytes SEM sinal
; Entrada: HL = bloco a, DE = bloco b, BC = quantidade (0 = iguais)
; Saída: A = 0 se iguais, 1 se a > b (no primeiro byte diferente), 0FFh se a < b;
;        flag Z = 1 quando iguais
; Destrói: BC, DE, HL, flags.
MEM_Compare:
    LD A, B
    OR C
    JR Z, MEM_Compare_Equal
MEM_Compare_Loop:
    LD A, (DE)
    CP (HL)             ; b - a
    JR NZ, MEM_Compare_Diff
    INC HL
    INC DE
    DEC BC
    LD A, B
    OR C
    JR NZ, MEM_Compare_Loop
MEM_Compare_Equal:
    XOR A
    RET
MEM_Compare_Diff:
    JR C, MEM_Compare_Greater   ; b < a
    LD A, 0FFh                  ; b > a  =>  a < b
    RET
MEM_Compare_Greater:
    LD A, 01h
    RET

ENDMOD
