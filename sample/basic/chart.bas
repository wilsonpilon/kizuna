' ============================================================
' KIZUNA sample -- Modulo Chart em MSX-BASIC Dignified
' Compilador: DIGNAC
' ============================================================

MODULE Chart
BANK 0
PUBLIC Main, Desenhar
EXTERN BIOS_CHGET

' Ponto de entrada para demonstracao grafica standalone
PROCEDURE Main()
    PRINT "[1] ENTRANDO NA SCREEN 2"
    SCREEN 2
    PRINT "[2] SCREEN 2 ATIVA"
    PRINT "[3] TRACANDO GRAFICO"
    ' Desenhar(10)   ' Diagnostico: testar SCREEN 2 sem executar as primitivas
    PRINT "[4] GRAFICO CONCLUIDO"
    BIOS_CHGET()
    PRINT "[5] SAINDO DA SCREEN 2"
    SCREEN 0
    PRINT "[6] SCREEN 0 ATIVA"
END PROCEDURE

' Desenhar: recebe um valor e traça um gráfico com moldura,
' eixos cartesianos, grade e curva calculada.
PROCEDURE Desenhar(valor%)
    LOCAL x%, y%

    PRINT "[3.1] LIMPANDO TELA"
    ' 1. Limpa a tela com fundo preto
    LINE (0,0)-(255,191), 1, BF

    PRINT "[3.2] TRACANDO MOLDURA"
    ' 2. Moldura retangular externa branca (cor 15)
    LINE (8, 8)-(247, 8), 15
    LINE (247, 8)-(247, 183), 15
    LINE (247, 183)-(8, 183), 15
    LINE (8, 183)-(8, 8), 15

    PRINT "[3.3] TRACANDO EIXOS"
    ' 3. Eixos cartesianos em Ciano (cor 7)
    LINE (24, 20)-(24, 165), 7
    LINE (24, 165)-(236, 165), 7

    PRINT "[3.4] TRACANDO GRADE"
    ' 4. Linhas de grade horizontais em Cinza (cor 14)
    LINE (24, 130)-(236, 130), 14
    LINE (24, 95)-(236, 95), 14
    LINE (24, 60)-(236, 60), 14

    PRINT "[3.5] TRACANDO CURVA"
    ' 5. Curva de pontos do grafico em Amarelo (cor 10)
    FOR x% = 25 TO 235
        y% = 160 - (x% MOD (valor% + 1)) * 9
        PSET (x%, y%), 10
    NEXT x%

END PROCEDURE
END MODULE
