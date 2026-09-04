; =============================================================================
; KIZUNA MSXLIB - STRING.ASM
; Rotinas de manipulação de strings e conversão numérica (Hex / Decimal)
; =============================================================================

MODULE STRING
BANK 0

PUBLIC StrLen, StrCopy, StrToUpper, PrintHex8, PrintHex16, PrintDec16
EXTERN BDOS_PrintChar

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

; -----------------------------------------------------------------------------
; StrCopy: Copia string terminada em zero da origem para o destino
; Entrada: HL = origem, DE = destino
; -----------------------------------------------------------------------------
StrCopy:
    PUSH AF
    PUSH HL
    PUSH DE
StrCopy_Loop:
    LD A, (HL)
    LD (DE), A
    OR A
    JR Z, StrCopy_End
    INC HL
    INC DE
    JR StrCopy_Loop
StrCopy_End:
    POP DE
    POP HL
    POP AF
    RET

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

; -----------------------------------------------------------------------------
; PrintHex8: Imprime byte em A como 2 dígitos hexadecimais no console
; Entrada: A = byte
; -----------------------------------------------------------------------------
PrintHex8:
    PUSH AF
    RRA
    RRA
    RRA
    RRA
    CALL PrintNibble
    POP AF
    CALL PrintNibble
    RET

PrintNibble:
    AND 0Fh
    CP 0Ah
    JR C, NibbleDigit
    ADD A, 07h
NibbleDigit:
    ADD A, 30h
    PUSH BC
    LD E, A
    CALL BDOS_PrintChar
    POP BC
    RET

; -----------------------------------------------------------------------------
; PrintHex16: Imprime palavra de 16 bits em HL como 4 dígitos hexadecimais
; Entrada: HL = palavra de 16 bits
; -----------------------------------------------------------------------------
PrintHex16:
    PUSH AF
    LD A, H
    CALL PrintHex8
    LD A, L
    CALL PrintHex8
    POP AF
    RET

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
