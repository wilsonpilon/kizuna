; =============================================================================
; KIZUNA MSXLIB - cstr/copy
; copia de string terminada em zero
; =============================================================================

MODULE cstr_copy
BANK 0

PUBLIC CSTR_Copy

; CSTR_Copy: copia a string de HL (com o terminador) para DE. O destino precisa
; ter espaco para comprimento + 1 bytes; nao pode sobrepor a origem.
; Entrada: HL = origem, DE = destino
; Saída: DE = endereco do terminador no destino (para encadear copias)
; Preserva: BC. Destrói: A, HL, flags.
CSTR_Copy:
    LD A, (HL)
    LD (DE), A
    OR A
    RET Z
    INC HL
    INC DE
    JR CSTR_Copy

ENDMOD
