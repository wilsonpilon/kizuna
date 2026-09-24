; =============================================================================
; KIZUNA MSXLIB - str/instr
; INSTR: posicao de uma string dentro de outra
; =============================================================================

MODULE str_instr
BANK 0

PUBLIC STR_InStr

; STR_InStr: posicao (a primeira e 1) da primeira ocorrencia da string DE (a
; agulha) dentro da string HL (o palheiro). Nao achou: 0. Agulha vazia: 1 se o
; palheiro tem algum caractere, senao 0.
; Entrada: HL = palheiro, DE = agulha (ambas tamanho+dados)
; Saída: A = 0..255
; Preserva: BC, DE, HL. Destrói: flags.
STR_InStr:
    PUSH BC
    PUSH DE
    PUSH HL
    PUSH IX
    LD B, (HL)          ; B = tamanho do palheiro
    INC HL              ; HL = dados
    LD A, (DE)
    LD C, A             ; C = tamanho da agulha
    INC DE              ; DE = dados
    OR A
    JR Z, STR_InStr_Empty
    LD A, B
    SUB C
    JR C, STR_InStr_Zero    ; agulha maior que o palheiro
    INC A               ; A = quantidade de posicoes candidatas
    LD B, A
    LD A, 01h
    PUSH AF             ; frame: (IX+6) = F, (IX+7) = A = indice da posicao atual
    PUSH BC             ;        (IX+4) = C = tamanho da agulha, (IX+5) = B = candidatas restantes
    PUSH DE             ;        (IX+2..3) = ponteiro da agulha
    PUSH HL             ;        (IX+0..1) = ponteiro do candidato
    LD IX, 0000h
    ADD IX, SP
STR_InStr_Outer:
    LD L, (IX+0)
    LD H, (IX+1)
    LD E, (IX+2)
    LD D, (IX+3)
    LD B, (IX+4)        ; B = caracteres a comparar
STR_InStr_Cmp:
    LD A, (DE)
    CP (HL)
    JR NZ, STR_InStr_Next
    INC HL
    INC DE
    DJNZ STR_InStr_Cmp
    LD A, (IX+7)        ; casou: A = indice da posicao
    JR STR_InStr_Found
STR_InStr_Next:
    LD L, (IX+0)
    LD H, (IX+1)
    INC HL
    LD (IX+0), L
    LD (IX+1), H        ; proximo candidato
    INC (IX+7)          ; indice++
    DEC (IX+5)          ; candidatas restantes--
    JR NZ, STR_InStr_Outer
    XOR A               ; nao achou
STR_InStr_Found:
    LD HL, 0008h
    ADD HL, SP
    LD SP, HL           ; descarta o frame de 8 bytes
    JR STR_InStr_Done
STR_InStr_Empty:
    LD A, B
    OR A
    JR Z, STR_InStr_Zero
    LD A, 01h
    JR STR_InStr_Done
STR_InStr_Zero:
    XOR A
STR_InStr_Done:
    POP IX
    POP HL
    POP DE
    POP BC
    RET

ENDMOD
