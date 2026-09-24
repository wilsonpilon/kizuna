; =============================================================================
; KIZUNA MSXLIB - cstr/findstr
; busca de uma substring
; =============================================================================

MODULE cstr_findstr
BANK 0

PUBLIC CSTR_FindStr

; CSTR_FindStr: primeira ocorrencia da string DE (agulha) dentro da string HL
; (palheiro). Agulha vazia acha no inicio do palheiro (como o strstr do C).
; Entrada: HL = palheiro, DE = agulha
; Saída: HL = endereco da ocorrencia, ou 0 se nao ha; flag Z = 1 se achou
; Preserva: BC, DE. Destrói: A, flags.
CSTR_FindStr:
CSTR_FindStr_Outer:
    PUSH HL
    PUSH DE
CSTR_FindStr_Inner:
    LD A, (DE)
    OR A
    JR Z, CSTR_FindStr_Found    ; agulha toda casou
    CP (HL)
    JR NZ, CSTR_FindStr_Miss
    INC HL
    INC DE
    JR CSTR_FindStr_Inner
CSTR_FindStr_Miss:
    POP DE
    POP HL
    LD A, (HL)
    OR A
    JR Z, CSTR_FindStr_None     ; acabou o palheiro
    INC HL
    JR CSTR_FindStr_Outer
CSTR_FindStr_Found:
    POP DE
    POP HL                      ; HL = inicio desta tentativa
    XOR A                       ; Z = 1
    RET
CSTR_FindStr_None:
    LD HL, 0000h
    LD A, 01h
    OR A                        ; Z = 0
    RET

ENDMOD
