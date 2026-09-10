' ============================================================
' KIZUNA sample -- Teste minimo e isolado de VDP_PSet
' Sem BoxFill, sem LINE, sem loop: so entra em SCREEN 2, planta
' UM pixel e volta. Serve para confirmar se o VDP_PSet reescrito
' (com leitura/escrita da Color Table dentro de um unico DI/EI)
' funciona sozinho, sem qualquer outra rotina no caminho.
' ============================================================

MODULE PSetTest
BANK 0
PUBLIC Main
EXTERN BIOS_CHGET

PROCEDURE Main()
    SCREEN 2
    PSET (128, 96), 15
    BIOS_CHGET()
    SCREEN 0
END PROCEDURE
END MODULE
