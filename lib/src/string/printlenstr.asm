; =============================================================================
; KIZUNA MSXLIB - string/printlenstr
; imprime string de comprimento dado
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_printlenstr
BANK 0

PUBLIC BDOS_PrintLenStr
EXTERN BDOS_PrintChar

; -----------------------------------------------------------------------------
; BDOS_PrintLenStr: Imprime no console uma string no formato "curto" (1 byte
; de tamanho + dados) -- diferente de BDOS_PrintString, que espera um
; terminador '$' e não sabe nada sobre esse formato.
; Entrada: HL = ponteiro para a string (tamanho + dados)
; -----------------------------------------------------------------------------
BDOS_PrintLenStr:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    LD A, (HL)
    LD B, A
    INC HL

    LD A, B
    OR A
    JR Z, BDOS_PrintLenStr_Done
BDOS_PrintLenStr_Loop:
    LD A, (HL)
    LD E, A
    PUSH HL
    PUSH BC
    CALL BDOS_PrintChar
    POP BC
    POP HL
    INC HL
    DJNZ BDOS_PrintLenStr_Loop
BDOS_PrintLenStr_Done:

    POP HL
    POP DE
    POP BC
    POP AF
    RET

ENDMOD
