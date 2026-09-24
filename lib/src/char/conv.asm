; =============================================================================
; KIZUNA MSXLIB - char/conv
; conversoes de caractere: maiuscula, minuscula, valor de digito e caractere de digito
; =============================================================================

MODULE char_conv
BANK 0

PUBLIC CHAR_ToUpper, CHAR_ToLower, CHAR_DigitValue, CHAR_HexChar

; CHAR_ToUpper: 'a'..'z' vira 'A'..'Z'; qualquer outro caractere fica igual
; Entrada/Saída: A. Preserva: BC, DE, HL. Destrói: flags.
CHAR_ToUpper:
    CP 61h
    RET C
    CP 7Bh
    RET NC
    SUB 20h
    RET

; CHAR_ToLower: 'A'..'Z' vira 'a'..'z'; qualquer outro caractere fica igual
; Entrada/Saída: A. Preserva: BC, DE, HL. Destrói: flags.
CHAR_ToLower:
    CP 41h
    RET C
    CP 5Bh
    RET NC
    ADD A, 20h
    RET

; CHAR_DigitValue: valor de um digito hexadecimal
; Entrada: A = caractere. Saída: A = 0..15 para '0'..'9', 'A'..'F', 'a'..'f'; 0FFh se nao for digito
; Preserva: BC, DE, HL. Destrói: flags.
CHAR_DigitValue:
    CP 30h
    JR C, CHAR_DigitValue_Bad
    CP 3Ah
    JR C, CHAR_DigitValue_Dec
    CP 41h
    JR C, CHAR_DigitValue_Bad
    CP 47h
    JR C, CHAR_DigitValue_Up
    CP 61h
    JR C, CHAR_DigitValue_Bad
    CP 67h
    JR NC, CHAR_DigitValue_Bad
    SUB 20h                 ; minuscula -> maiuscula
CHAR_DigitValue_Up:
    SUB 07h                 ; 'A' (41h) -> 3Ah, o mesmo trecho dos decimais
CHAR_DigitValue_Dec:
    SUB 30h
    RET
CHAR_DigitValue_Bad:
    LD A, 0FFh
    RET

; CHAR_HexChar: caractere do digito hexadecimal (maiusculo) de um valor
; Entrada: A = valor (so os 4 bits de baixo contam). Saída: A = '0'..'9' ou 'A'..'F'
; Preserva: BC, DE, HL. Destrói: flags.
CHAR_HexChar:
    AND 0Fh
    ADD A, 30h
    CP 3Ah
    RET C
    ADD A, 07h
    RET

ENDMOD
