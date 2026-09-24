MODULE ApiDemo
BANK 0
PUBLIC Main

' Chama rotinas da MSXLIB direto do BASIC, pelos descritores lib/api/*.api:
' nenhum comando dedicado no compilador, nenhum EXTERN.
PROCEDURE Main()
    LOCAL k%

    VDP_SetColor(15, 4)                  ' texto branco, fundo azul
    BIOS_CLS()
    PRINT "KIZUNA - MSXLIB chamada pelo BASIC via .api"
    PRINT "-------------------------------------------"

    PRINT "6 * 7 ="
    PrintDec16(Mul16(6, 7))              ' 42
    BDOS_PrintChar(13)
    BDOS_PrintChar(10)

    PRINT "100 / 7 ="
    PrintDec16(Div16(100, 7))            ' 14
    BDOS_PrintChar(13)
    BDOS_PrintChar(10)

    PRINT "255 em hexa:"
    PrintHex8(255)                       ' FF
    BDOS_PrintChar(13)
    BDOS_PrintChar(10)

    PRINT "Tom de 440 Hz no canal A. Aperte uma tecla para parar."
    PSG_PlayTone(0, 254, 12)             ' periodo 254 = La 4
    k% = BDOS_ReadChar()
    PSG_MuteAll()

    VDP_SetColor(15, 1)                  ' devolve o fundo preto
    PRINT "Fim."
END PROCEDURE
END MODULE
