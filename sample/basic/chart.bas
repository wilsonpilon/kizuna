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
    SCREEN 2
    Desenhar(10)
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE

' Desenhar: recebe um valor e traça um gráfico com moldura,
' eixos cartesianos, grade e curva calculada.
PROCEDURE Desenhar(valor%)
    LOCAL x%, y%

    ' Checkpoints: um PSET de cor unica por etapa, na coluna x=2 (fora da
    ' area da moldura/eixos/grade/curva, ninguem mais escreve ali), cada um
    ' numa linha Y diferente para nao dar color clash entre eles.
    ' Ordem/cores: 15 branco, 8 vermelho medio, 5 azul claro, 11 amarelo
    ' claro, 13 magenta, 7 ciano, 3 verde claro, 10 amarelo escuro,
    ' 6 vermelho escuro, 12 verde escuro, 9 rosa/vermelho claro.

    ' 1. Limpa a tela com fundo preto
    LINE (0,0)-(255,191), 1, BF
    PSET (2, 2), 15   ' checkpoint 1: BoxFill (limpar tela) OK
    BIOS_CHGET()

    ' 2. Moldura retangular externa branca (cor 15)
    LINE (8, 8)-(247, 8), 15
    PSET (2, 6), 8    ' checkpoint 2: borda topo OK
    BIOS_CHGET()
    LINE (247, 8)-(247, 183), 15
    PSET (2, 10), 5   ' checkpoint 3: borda direita OK
    BIOS_CHGET()
    LINE (247, 183)-(8, 183), 15
    PSET (2, 14), 11  ' checkpoint 4: borda baixo OK
    BIOS_CHGET()
    LINE (8, 183)-(8, 8), 15
    PSET (2, 18), 13  ' checkpoint 5: borda esquerda OK
    BIOS_CHGET()

    ' 3. Eixos cartesianos em Ciano (cor 7)
    LINE (24, 20)-(24, 165), 7
    PSET (2, 22), 7   ' checkpoint 6: eixo vertical OK
    BIOS_CHGET()
    LINE (24, 165)-(236, 165), 7
    PSET (2, 26), 3   ' checkpoint 7: eixo horizontal OK
    BIOS_CHGET()

    ' 4. Linhas de grade horizontais em Cinza (cor 14)
    LINE (24, 130)-(236, 130), 14
    PSET (2, 30), 10  ' checkpoint 8: grade 1 OK
    BIOS_CHGET()
    LINE (24, 95)-(236, 95), 14
    PSET (2, 34), 6   ' checkpoint 9: grade 2 OK
    BIOS_CHGET()
    LINE (24, 60)-(236, 60), 14
    PSET (2, 38), 12  ' checkpoint 10: grade 3 OK
    BIOS_CHGET()

    ' 5. Curva de pontos do grafico em Amarelo (cor 10)
    FOR x% = 25 TO 235
        y% = 160 - (x% MOD (valor% + 1)) * 9
        PSET (x%, y%), 10
    NEXT x%
    PSET (2, 42), 9   ' checkpoint 11: curva (loop completo) OK
    BIOS_CHGET()

END PROCEDURE
END MODULE
