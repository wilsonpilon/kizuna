package z80sim

// runCommand executa o comando escrito em R#46. Nesta versão do modelo nenhum
// comando está implementado (o motor de comandos entra junto com as rotinas
// de comando da MSXLIB); um comando é aceito e "termina" na hora, sem efeito.
func (v *VDP) runCommand() {
	v.Status[2] &^= 0x01 // CE = 0: nada em execução
}
