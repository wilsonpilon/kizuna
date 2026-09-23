; =============================================================================
; KIZUNA MSXLIB - string/strtoupper
; converte pra maiusculas
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_strtoupper
BANK 0

PUBLIC StrToUpper

; -----------------------------------------------------------------------------
; StrToUpper: Converte caracteres minúsculos ('a'..'z') para maiúsculas in-place
; Entrada: HL = ponteiro para a string
; -----------------------------------------------------------------------------
StrToUpper:
    PUSH AF
    PUSH HL
StrUp_Loop:
    LD A, (HL)
    OR A
    JR Z, StrUp_End
    CP 61h ; 'a'
    JR C, StrUp_Next
    CP 7Bh ; 'z' + 1
    JR NC, StrUp_Next
    SUB 20h
    LD (HL), A
StrUp_Next:
    INC HL
    JR StrUp_Loop
StrUp_End:
    POP HL
    POP AF
    RET

ENDMOD
