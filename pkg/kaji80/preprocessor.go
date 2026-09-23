package kaji80

import (
	"fmt"
	"strings"
)

// =============================================================================
// KIZUNA KAJI80 - Pré-processador (rótulos locais nesta leva; IF/REPT/MACRO
// entram em leva futura, ver plano em
// C:\Users\wilso\.claude\plans\robust-discovering-lemon.md)
//
// Nota de arquitetura: o plano original descrevia isto como um passo em
// texto bruto, rodando ANTES do lexer. Rótulos locais, na prática, não
// precisam disso -- o lexer já trata '.' como caractere de identificador
// (isIdentStart/isIdentPart), então ".loop" já chega como UM token
// IDENTIFIER só, sem ambiguidade com comentário ou string (que o lexer já
// separa corretamente). Processar no nível de TOKEN em vez de texto bruto
// evita reimplementar a lógica de "ignorar comentário/string" que o lexer
// já tem, e continua compatível com a ordem de execução do plano: REPT/
// MACRO (quando existirem) vão rodar em texto bruto ANTES do lexer, então
// um rótulo local dentro de um bloco expandido por eles já chega aqui como
// texto final (com qualquer sufixo de expansão única já embutido no nome)
// -- o mangle "por rótulo global" abaixo continua sendo, na prática, o
// ÚLTIMO passo do pipeline, como o plano pedia.
// =============================================================================

// resolveLocalLabels percorre o fluxo de tokens inteiro (já agrupado por
// linha) e renomeia toda definição/referência de rótulo local (".nome")
// para "<RótuloGlobalAtual>_nome", em lugar (mutando Value dos tokens).
// Rótulo global é qualquer "IDENT:" sem ponto na frente -- vale a partir
// da linha em que é definido até o próximo rótulo global. Um rótulo local
// usado antes de qualquer rótulo global no arquivo é erro de compilação.
func resolveLocalLabels(lineTokens [][]Token) error {
	currentGlobal := ""
	for _, line := range lineTokens {
		if len(line) == 0 {
			continue
		}

		// Uma definição de rótulo GLOBAL (sem ponto, seguida de ':') abre
		// um novo escopo a partir desta linha em diante.
		if line[0].Type == TokenIdentifier && !strings.HasPrefix(line[0].Value, ".") &&
			len(line) > 1 && line[1].Type == TokenColon {
			currentGlobal = line[0].Value
		}

		for i := range line {
			if line[i].Type != TokenIdentifier || !strings.HasPrefix(line[i].Value, ".") {
				continue
			}
			if currentGlobal == "" {
				return fmt.Errorf("linha %d: rótulo local '%s' usado antes de qualquer rótulo global no arquivo", line[i].Line, line[i].Value)
			}
			line[i].Value = currentGlobal + "_" + line[i].Value[1:]
		}
	}
	return nil
}
