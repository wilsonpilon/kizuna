; ==============================================================================
; KIZUNA sample -- OBI: multi-modulo (KAJI80 + DIGNAC), multi-banco, com
; resource embutido, tudo orquestrado declarativamente por um Obifile.
; Compilador: KAJI80 (dono do Start)
; Orquestrado por: obi build sample/obi/Obifile
; ==============================================================================

MODULE OBIMAIN
BANK 0

PUBLIC Start
EXTERN BIOS_CHGMOD, BIOS_CHGET, BDOS_PrintString, BDOS_Exit
EXTERN Desenhar        ; PROCEDURE do modulo DIGNAC (chart_lib.bas, banco 2)
EXTERN Res_Banner      ; resource embutido pelo OBI a partir de banner.txt

Start:
    ; Mensagem de abertura (resource embutido, terminado em '$')
    LD DE, Res_Banner
    CALL BDOS_PrintString

    ; SCREEN 2
    LD A, 2
    CALL BIOS_CHGMOD

    ; Desenhar(6) -- chama o modulo DIGNAC no banco 2; o MUSUBI ja gerou o
    ; trampolim de troca de banco automaticamente, o codigo aqui nao precisa
    ; se preocupar com isso.
    LD HL, 6
    PUSH HL
    CALL Desenhar
    POP HL

    ; Aguarda uma tecla antes de voltar ao MSX-DOS
    CALL BIOS_CHGET

    ; Volta para SCREEN 0 (texto)
    XOR A
    CALL BIOS_CHGMOD

    CALL BDOS_Exit
    RET

ENDMOD
