; =============================================================================
; KIZUNA MSXLIB - string/strcopy
; copia string
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE string_strcopy
BANK 0

PUBLIC StrCopy

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

ENDMOD
