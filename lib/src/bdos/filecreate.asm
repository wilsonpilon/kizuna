; =============================================================================
; KIZUNA MSXLIB - bdos/filecreate
; cria arquivo (MSX-DOS 2)
; (um modulo por rotina/familia -- o linker so traz o que o programa usa)
; =============================================================================

MODULE bdos_filecreate
BANK 0

PUBLIC BDOS_FileCreate
INCLUDE "../../inc/bdos.inc"

; -----------------------------------------------------------------------------
; BDOS_FileCreate: Cria (ou trunca) um arquivo (Função 44h, MSX-DOS 2)
; Entrada: DE = ponteiro para caminho ASCIIZ, A = modo (BDOS_FMODE_*),
;          B = atributos (BDOS_ATTR_*)
; Saída: A = código de erro (0 = sucesso), B = handle do arquivo
; -----------------------------------------------------------------------------
BDOS_FileCreate:
    LD C, BDOS_F_CREATE
    CALL BDOS_ENTRY
    RET

ENDMOD
