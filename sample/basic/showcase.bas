' ============================================================
' KIZUNA sample -- Showcase de DIGNAC: sprite, musica PSG e
' I/O de arquivo, tudo direto em MSX-BASIC Dignified (sem cair
' pra Assembly puro).
' Compilador: DIGNAC
' ============================================================

MODULE Showcase
BANK 0
PUBLIC Main
EXTERN BIOS_CHGET

PROCEDURE Main()
    SCREEN 2

    ' Bolinha 16x16 (4 quadrantes 8x8, mesma tabela do sample/sprites)
    SPRITE PATTERN 0, 3, 15, 31, 63, 127, 127, 255, 255, 255, 255, 127, 127, 63, 31, 15, 3, 192, 240, 248, 252, 254, 254, 255, 255, 255, 255, 254, 254, 252, 248, 240, 192
    PUT SPRITE 0, (120, 88), 15, 0

    ' Arpejo de Do maior, subindo uma oitava no final
    PLAY "O4 L4 CEG O5 C"

    ' Grava um arquivo de "pontuacao" e confirma escrevendo/lendo de volta
    OPEN "SCORE.TXT" FOR OUTPUT AS #1
    PRINT #1, "KIZUNA DIGNAC -- pontuacao"
    PRINT #1, "Sprites: ", 1
    PRINT #1, "Notas tocadas: ", 4
    CLOSE #1

    PRINT "Demo concluida! Confira SCORE.TXT no disco."

    BIOS_CHGET()

    SPRITE OFF
    SCREEN 0
END PROCEDURE
END MODULE
