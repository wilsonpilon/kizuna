; Portas e enderecos do VDP (TMS9918 / V9938 / V9958)
VDP_DATA EQU 0098h
VDP_CMD  EQU 0099h

; Tabelas de sprite em SCREEN 2 -- mesmos enderecos-padrao do MSX-BASIC
; (confirmado contra resource/MSXgl/engine/src/vdp_reg.h: R#5 = 36h e
; R#6 = 07h para SCREEN 2)
VDP_SPRITE_ATTR_TABLE    EQU 1B00h
VDP_SPRITE_PATTERN_TABLE EQU 3800h
