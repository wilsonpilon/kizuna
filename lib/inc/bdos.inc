; Entrada do BDOS e codigos de funcao do MSX-DOS 2
BDOS_ENTRY EQU 0005h

; Funcoes de I/O de arquivo baseado em handle do MSX-DOS 2 (nao o FCB antigo
; de DOS 1) -- numeros confirmados contra o protocolo oficial de MSX-DOS 2
; (map.grauw.nl/resources/dos2_functioncalls.php).
BDOS_F_OPEN   EQU 43h
BDOS_F_CREATE EQU 44h
BDOS_F_CLOSE  EQU 45h
BDOS_F_READ   EQU 48h
BDOS_F_WRITE  EQU 49h
BDOS_F_SEEK   EQU 4Ah

; Modo de abertura (byte A de BDOS_FileOpen/BDOS_FileCreate)
BDOS_FMODE_RW     EQU 00h
BDOS_FMODE_RDONLY EQU 01h
BDOS_FMODE_WRONLY EQU 02h

; Atributo de criacao (byte B de BDOS_FileCreate)
BDOS_ATTR_NORMAL     EQU 00h
BDOS_ATTR_CREATE_NEW EQU 80h
