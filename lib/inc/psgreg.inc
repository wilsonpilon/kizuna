; =============================================================================
; KIZUNA MSXLIB - registradores do PSG (AY-3-8910 / YM2149) e joystick
; =============================================================================
PSG_R_TONE_A_LO  EQU 0      ; periodo do canal A, byte baixo
PSG_R_TONE_A_HI  EQU 1      ;                     nibble alto (4 bits)
PSG_R_TONE_B_LO  EQU 2
PSG_R_TONE_B_HI  EQU 3
PSG_R_TONE_C_LO  EQU 4
PSG_R_TONE_C_HI  EQU 5
PSG_R_NOISE      EQU 6      ; periodo do ruido (5 bits)
PSG_R_MIXER      EQU 7      ; mixer + direcao das portas de E/S
PSG_R_VOL_A      EQU 8      ; volume do canal A (bit 4 = usa o envelope)
PSG_R_VOL_B      EQU 9
PSG_R_VOL_C      EQU 10
PSG_R_ENV_LO     EQU 11     ; periodo do envelope, byte baixo
PSG_R_ENV_HI     EQU 12     ;                      byte alto
PSG_R_ENV_SHAPE  EQU 13     ; forma do envelope (4 bits)
PSG_R_IO_A       EQU 14     ; porta A (joysticks): leitura
PSG_R_IO_B       EQU 15     ; porta B: seleciona o joystick, LED Kana

; Mixer (R#7): bit em 1 = DESLIGADO
PSG_MIX_TONE_A_OFF  EQU 01h
PSG_MIX_TONE_B_OFF  EQU 02h
PSG_MIX_TONE_C_OFF  EQU 04h
PSG_MIX_NOISE_A_OFF EQU 08h
PSG_MIX_NOISE_B_OFF EQU 10h
PSG_MIX_NOISE_C_OFF EQU 20h
PSG_MIX_IO_A_OUT    EQU 40h ; porta A como saida (no MSX deve ficar em 0 = entrada)
PSG_MIX_IO_B_OUT    EQU 80h ; porta B como saida (no MSX deve ficar em 1)
PSG_MIX_ALL_OFF     EQU 0BFh ; nenhum canal audivel, portas no estado normal do MSX

PSG_VOL_ENVELOPE    EQU 10h ; bit do registrador de volume: usa o envelope

; Bits do R#13 (forma do envelope)
PSG_SHAPE_HOLD     EQU 01h  ; mantem o ultimo valor ao fim do primeiro ciclo
PSG_SHAPE_ALTERNATE EQU 02h ; inverte o sentido a cada ciclo
PSG_SHAPE_ATTACK   EQU 04h  ; comeca subindo (0 = comeca descendo)
PSG_SHAPE_CONTINUE EQU 08h  ; repete/continua depois do primeiro ciclo

; Formas prontas, desenhadas como o volume evolui ao longo do tempo
PSG_ENV_DECAY_LOW   EQU 09h ; \___  cai e fica em silencio
PSG_ENV_ATTACK_LOW  EQU 0Fh ; /___  sobe e cai de uma vez
PSG_ENV_SAW_DOWN    EQU 08h ; \\\\  serra decrescente repetida
PSG_ENV_SAW_UP      EQU 0Ch ; ////  serra crescente repetida
PSG_ENV_TRI_DOWN    EQU 0Ah ; \/\/  triangulo, comeca descendo
PSG_ENV_TRI_UP      EQU 0Eh ; /\/\  triangulo, comeca subindo
PSG_ENV_DECAY_HIGH  EQU 0Bh ; \---  cai e salta para o maximo
PSG_ENV_ATTACK_HIGH EQU 0Dh ; /---  sobe e fica no maximo

; Joystick: leitura de R#14 (0 = pressionado)
JOY_UP       EQU 01h
JOY_DOWN     EQU 02h
JOY_LEFT     EQU 04h
JOY_RIGHT    EQU 08h
JOY_BUTTON_A EQU 10h
JOY_BUTTON_B EQU 20h
; Escrita em R#15
PSG_R15_PORT2    EQU 40h    ; 0 = joystick da porta 1, 1 = porta 2
PSG_R15_KANA_LED EQU 80h    ; 0 = LED Kana aceso
