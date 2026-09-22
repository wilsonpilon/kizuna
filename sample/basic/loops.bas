' ============================================================
' KIZUNA sample -- Exercita dois bugs reais corrigidos na v4.8.0:
' FOR com STEP negativo (o teste de termino agora considera a
' direcao do STEP) e OPEN...FOR APPEND (agora de fato posiciona
' no fim do arquivo antes de escrever, em vez de truncar como
' FOR OUTPUT fazia por engano).
' Compilador: DIGNAC
' ============================================================

MODULE Loops
BANK 0
PUBLIC Main

PROCEDURE Main()
    PRINT "Contagem regressiva:"
    LOCAL i%
    FOR i% = 5 TO 1 STEP -1
        PRINT i%
    NEXT i%
    PRINT "Fim!"

    ' Duas aberturas em APPEND -- a segunda deve ACRESCENTAR, nao
    ' sobrescrever a primeira linha
    OPEN "LOOPS.TXT" FOR OUTPUT AS #1
    PRINT #1, "primeira linha"
    CLOSE #1

    OPEN "LOOPS.TXT" FOR APPEND AS #1
    PRINT #1, "segunda linha (append)"
    CLOSE #1
END PROCEDURE
END MODULE
