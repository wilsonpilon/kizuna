; =============================================================================
; KIZUNA MSXLIB - V9990 (GFX9000 / Video9000): portas
; O mapa de registradores (R#0-R#55), modos e comandos do V9990 entra na fase
; dedicada a ele.
; =============================================================================
PORT_V9990_VRAM    EQU 60h  ; P#0: dados da VRAM
PORT_V9990_PAL     EQU 61h  ; P#1: dados da paleta
PORT_V9990_CMD     EQU 62h  ; P#2: dados de comando
PORT_V9990_REGDATA EQU 63h  ; P#3: dado de registrador
PORT_V9990_REGSEL  EQU 64h  ; P#4: selecao de registrador
PORT_V9990_STATUS  EQU 65h  ; P#5: status
PORT_V9990_INTFLAG EQU 66h  ; P#6: flags de interrupcao
PORT_V9990_SYSCTRL EQU 67h  ; P#7: controle de sistema (reset, modo)
PORT_V9990_P15     EQU 6Fh  ; P#15: seletor de joystick/saida (GFX9000 e Video9000)
