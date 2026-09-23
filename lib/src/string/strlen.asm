; =============================================================================
; KIZUNA MSXLIB - string/strlen
; comprimento de string
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_strlen
BANK 0

PUBLIC StrLen

; -----------------------------------------------------------------------------
; StrLen: Calcula o tamanho de uma string terminada em zero (\0)
; Entrada: HL = ponteiro para a string
; Saída: BC = comprimento da string em bytes
; -----------------------------------------------------------------------------
StrLen:
    PUSH HL
    LD BC, 0000h
StrLen_Loop:
    LD A, (HL)
    OR A
    JR Z, StrLen_End
    INC BC
    INC HL
    JR StrLen_Loop
StrLen_End:
    POP HL
    RET

ENDMOD
