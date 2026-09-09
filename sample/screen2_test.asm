; KIZUNA - Teste minimo de SCREEN 2
; Testa pontos coloridos usando VDP_PSet.

MODULE Screen2Test
BANK 0
PUBLIC Start
EXTERN VDP_InitScreen2_Tables, VDP_SetWriteAddr
EXTERN BDOS_ReadChar, BDOS_Exit
EXTERN BIOS_CHGMOD

Start:
    ; Inicializa SCREEN 2 e prepara Name, Pattern, Color e Sprite Tables.
    CALL VDP_InitScreen2_Tables

    ; Teste deterministico sem PSet: escreve diretamente o byte do ponto
    ; central (X=128,Y=96): Pattern Table endereco 0C80h, mascara 80h.
    LD HL, 0C80h
    CALL VDP_SetWriteAddr
    LD A, 080h
    OUT (0098h), A

    ; Mantem a imagem na tela ate uma tecla.
    CALL BDOS_ReadChar
    ; Restaura a tela de texto antes de devolver o controle ao DOS.
    XOR A
    CALL BIOS_CHGMOD
    JP BDOS_Exit

ENDMOD
