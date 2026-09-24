; =============================================================================
; KIZUNA MSXLIB - mem/sysinfo
; informacoes de memoria do sistema (pilha e topo da TPA)
; =============================================================================

MODULE mem_sysinfo
BANK 0

PUBLIC MEM_GetSP, MEM_TPATop

; MEM_GetSP: valor que SP tera no chamador depois do RET (ou seja, o topo da
; pilha do programa, sem contar o endereco de retorno desta chamada)
; Saída: HL
; Preserva: A, BC, DE. Destrói: flags.
MEM_GetSP:
    LD HL, 0002h
    ADD HL, SP
    RET

; MEM_TPATop: endereco guardado em 0006h, o inicio do BDOS (a TPA -- a memoria
; livre do programa -- vai ate este endereco menos 1)
; Saída: HL
; Preserva: A, BC, DE. Destrói: nada.
MEM_TPATop:
    LD HL, (0006h)
    RET

ENDMOD
