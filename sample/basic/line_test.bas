' ============================================================
' KIZUNA sample -- Teste minimo e isolado de VDP_Line (horizontal)
' Sem BoxFill, sem checkpoints, sem loop de espera repetido: so
' entra em SCREEN 2, traca UMA linha horizontal e volta.
' ============================================================

MODULE LineTest
BANK 0
PUBLIC Main
EXTERN BIOS_CHGET

PROCEDURE Main()
    SCREEN 2
    LINE (8, 8)-(247, 8), 15
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE
END MODULE
