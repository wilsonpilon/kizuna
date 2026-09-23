; =============================================================================
; KIZUNA MSXLIB - bdos/fileopen
; abre arquivo (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_fileopen
BANK 0

PUBLIC BDOS_FileOpen
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_FileOpen: Abre um arquivo existente (Função 43h, MSX-DOS 2)
; Entrada: DE = ponteiro para caminho ASCIIZ, A = modo (BDOS_FMODE_*)
; Saída: A = código de erro (0 = sucesso), B = handle do arquivo
; -----------------------------------------------------------------------------
BDOS_FileOpen:
    LD C, BDOS_F_OPEN
    CALL BDOS_ENTRY
    RET

ENDMOD
