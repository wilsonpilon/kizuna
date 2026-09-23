; =============================================================================
; KIZUNA MSXLIB - bdos/filewrite
; escreve num handle (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_filewrite
BANK 0

PUBLIC BDOS_FileWrite
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_FileWrite: Escreve bytes num handle de arquivo (Função 49h, MSX-DOS 2)
; Entrada: B = handle, DE = buffer de origem, HL = quantidade de bytes a
;          escrever
; Saída: A = código de erro (0 = sucesso), HL = bytes realmente escritos
; -----------------------------------------------------------------------------
BDOS_FileWrite:
    LD C, BDOS_F_WRITE
    CALL BDOS_ENTRY
    RET

ENDMOD
