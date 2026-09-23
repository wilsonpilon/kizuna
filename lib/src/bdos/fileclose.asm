; =============================================================================
; KIZUNA MSXLIB - bdos/fileclose
; fecha handle (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_fileclose
BANK 0

PUBLIC BDOS_FileClose
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_FileClose: Fecha um handle de arquivo aberto (Função 45h, MSX-DOS 2)
; Entrada: B = handle do arquivo
; Saída: A = código de erro (0 = sucesso)
; -----------------------------------------------------------------------------
BDOS_FileClose:
    LD C, BDOS_F_CLOSE
    CALL BDOS_ENTRY
    RET

ENDMOD
