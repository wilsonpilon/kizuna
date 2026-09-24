; =============================================================================
; KIZUNA MSXLIB - codigos de controle e de teclas (ASCII + extensoes do MSX)
; =============================================================================
ASCII_NUL  EQU 00h
ASCII_BEL  EQU 07h  ; beep
ASCII_BS   EQU 08h  ; backspace
ASCII_TAB  EQU 09h
ASCII_LF   EQU 0Ah
ASCII_HOME EQU 0Bh  ; cursor para o canto (o MSX usa 0Bh, nao o "VT" do ASCII)
ASCII_CLS  EQU 0Ch  ; limpa a tela
ASCII_CR   EQU 0Dh
ASCII_INS  EQU 12h
ASCII_ESC  EQU 1Bh
ASCII_RIGHT EQU 1Ch
ASCII_LEFT  EQU 1Dh
ASCII_UP    EQU 1Eh
ASCII_DOWN  EQU 1Fh
ASCII_SPACE EQU 20h
ASCII_DEL   EQU 7Fh
; Teclas de funcao devolvidas por CHGET/tabela de teclado do MSX: F1-F5 = 0F1h..0F5h
; nao sao ASCII; use a BIOS (GTSTCK/CHGET) para le-las.
