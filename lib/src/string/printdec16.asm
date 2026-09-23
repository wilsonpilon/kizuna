; =============================================================================
; KIZUNA MSXLIB - string/printdec16
; imprime decimal de 16 bits
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_printdec16
BANK 0

PUBLIC PrintDec16
EXTERN BDOS_PrintChar

; -----------------------------------------------------------------------------
; PrintDec16: Imprime número não sinalizado de 16 bits em HL em decimal (0..65535)
; Suprime zeros à esquerda automaticamente.
; Entrada: HL = valor de 16 bits
; Preserva: todos os registradores
; -----------------------------------------------------------------------------
PrintDec16:
    PUSH AF
    PUSH BC
    PUSH DE
    PUSH HL

    LD C, 00h ; C = 0 indica supressão de zeros à esquerda

    ; Dígito de 10000
    LD DE, 2710h ; 10000
    CALL PrintDecDigit

    ; Dígito de 1000
    LD DE, 03E8h ; 1000
    CALL PrintDecDigit

    ; Dígito de 100
    LD DE, 0064h ; 100
    CALL PrintDecDigit

    ; Dígito de 10
    LD DE, 000Ah ; 10
    CALL PrintDecDigit

    ; Dígito das unidades (1): imprime sempre (0..9)
    LD A, L
    ADD A, 30h ; '0'
    LD E, A
    CALL BDOS_PrintChar

    POP HL
    POP DE
    POP BC
    POP AF
    RET

; Subrotina interna: calcula um dígito subtraindo DE repetidamente de HL
PrintDecDigit:
    LD B, 2Fh ; '0' - 1
DecDigit_Loop:
    INC B
    OR A
    SBC HL, DE
    JR NC, DecDigit_Loop
    ADD HL, DE ; Restaura último excesso

    LD A, B
    CP 30h ; '0'
    JR NZ, DecDigit_Print
    ; É '0': verifica se já começamos a imprimir dígitos significativos
    LD A, C
    OR A
    RET Z ; Se C == 0, suprime o zero e retorna

    LD A, 30h
DecDigit_Print:
    LD C, 01h ; Ativa flag: zeros seguintes devem ser impressos
    PUSH DE
    PUSH HL
    LD E, A
    CALL BDOS_PrintChar
    POP HL
    POP DE
    RET

ENDMOD
