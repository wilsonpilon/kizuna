package kaji80

import (
	"fmt"
	"math"
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

// filterConditionals implementa IF condição / ELSE / ENDIF: remove do fluxo
// de linhas tudo que estiver num ramo morto, ANTES de qualquer outra coisa
// (rótulos locais, e futuramente REPT/MACRO) enxergar essas linhas --
// evita, por exemplo, uma macro dentro de um ramo morto sendo registrada à
// toa, e garante que um rótulo local dentro de um ramo morto não force um
// escopo/erro que nunca deveria existir.
//
// Roda ANTES do Pass 1 de verdade, mas condições IF frequentemente
// referenciam uma constante EQU ou variável já definida mais acima no
// arquivo (ex. clássico: "FORMATO=1 / IF FORMATO==1 ..."). Pass 1 é quem
// normalmente constrói a.constants/a.variables, só que Pass 1 só roda
// DEPOIS deste filtro (só vê as linhas já filtradas). Pra resolver isso
// sem duplicar toda a lógica de EQU/ASSIGN, este filtro também processa
// EQU/ASSIGN sequencialmente (só nos ramos ativos, reaproveitando
// a.parseLine e a.EvalExpr) só pra ter o valor certo disponível na hora de
// avaliar um IF -- e reseta a.constants/a.variables no final, pra Pass 1
// reconstruir do zero normalmente sobre a lista já filtrada, sem herdar
// nada deste passo.
func (a *Assembler) filterConditionals(lineTokens [][]Token) ([][]Token, error) {
	a.constants = make(map[string]int64)
	a.variables = make(map[string]float64)

	var out [][]Token
	var stack []bool // true = ramo vivo neste nível de aninhamento
	active := func() bool {
		for _, v := range stack {
			if !v {
				return false
			}
		}
		return true
	}

	for _, line := range lineTokens {
		if len(line) == 0 {
			continue
		}
		first := line[0]

		if first.Type == TokenIdentifier {
			switch strings.ToUpper(first.Value) {
			case "IF":
				if len(line) < 2 {
					return nil, fmt.Errorf("linha %d: IF requer uma condição", first.Line)
				}
				if !active() {
					// Não avalia a condição -- pode depender de algo não
					// definido neste ramo já morto. Só empilha um nível
					// falso pra casar com o ENDIF/ELSE correspondente.
					stack = append(stack, false)
					continue
				}
				pl, err := a.parseLine(line)
				if err != nil {
					return nil, err
				}
				if len(pl.operands) == 0 {
					return nil, fmt.Errorf("linha %d: IF requer uma condição", first.Line)
				}
				val, ok, err := a.EvalExpr(pl.operands[0])
				if err != nil {
					return nil, fmt.Errorf("linha %d: IF %s: %w", first.Line, pl.operands[0], err)
				}
				if !ok {
					return nil, fmt.Errorf("linha %d: IF %s: condição não é uma expressão numérica válida", first.Line, pl.operands[0])
				}
				stack = append(stack, val != 0)
				continue
			case "ELSE":
				if len(stack) == 0 {
					return nil, fmt.Errorf("linha %d: ELSE sem IF correspondente", first.Line)
				}
				stack[len(stack)-1] = !stack[len(stack)-1]
				continue
			case "ENDIF":
				if len(stack) == 0 {
					return nil, fmt.Errorf("linha %d: ENDIF sem IF correspondente", first.Line)
				}
				stack = stack[:len(stack)-1]
				continue
			}
		}

		if !active() {
			continue
		}

		// EQU/ASSIGN processadas aqui só pra condições IF mais adiante no
		// arquivo enxergarem o valor certo -- qualquer erro real de
		// verdade (sintaxe malformada, símbolo desconhecido) é reportado
		// de novo, com a mensagem completa, pelo Pass 1 de verdade
		// depois -- este passo só ignora silenciosamente em caso de erro
		// aqui, pra não duplicar toda a lógica de mensagem de erro.
		pl, err := a.parseLine(line)
		if err != nil {
			return nil, err
		}
		switch strings.ToUpper(pl.mnemonic) {
		case "EQU":
			if pl.label != "" && len(pl.operands) > 0 {
				if val, ok, err := a.EvalExpr(pl.operands[0]); err == nil && ok {
					a.constants[pl.label] = int64(math.Round(val))
				}
			}
		case "ASSIGN":
			if pl.label != "" && len(pl.operands) > 0 {
				if val, ok, err := a.EvalExpr(pl.operands[0]); err == nil && ok {
					a.variables[pl.label] = val
				}
			}
		}

		out = append(out, line)
	}

	if len(stack) != 0 {
		return nil, fmt.Errorf("%d IF sem ENDIF correspondente", len(stack))
	}

	a.constants = make(map[string]int64)
	a.variables = make(map[string]float64)
	return out, nil
}

// expandRept implementa REPT n / ENDR: duplica o bloco de linhas entre REPT
// e o ENDR correspondente n vezes. n precisa ser um literal inteiro (mesma
// restrição do asMSX -- não pode vir de expressão), então REPT não usa o
// avaliador de expressões da Fase 1 -- só o campo Number, já calculado pelo
// lexer pra qualquer literal numérico (decimal/hex/binário). Aninhamento
// permitido: o corpo de um REPT é expandido recursivamente ANTES de cada
// cópia externa ser duplicada, então cada combinação (iteração externa,
// iteração interna) recebe seu próprio ID de expansão único, compartilhado
// via ponteiro entre todas as chamadas recursivas.
//
// Uma linha de atribuição de variável (Fase 1, "X=X+1") dentro do bloco
// não é avaliada aqui -- REPT só duplica texto/tokens; a atribuição de
// verdade acontece depois, no Pass 1/2 de verdade, na ordem sequencial das
// cópias já expandidas (é assim que o exemplo canônico do asMSX --
// "X=0 / REPT 10 / DB X*Y / X=X+1 / ENDR" -- consegue emitir um valor
// diferente em cada DB).
func expandRept(lineTokens [][]Token, expCounter *int) ([][]Token, error) {
	var out [][]Token
	i := 0
	for i < len(lineTokens) {
		line := lineTokens[i]
		if len(line) == 0 {
			i++
			continue
		}
		if line[0].Type == TokenIdentifier && strings.EqualFold(line[0].Value, "REPT") {
			if len(line) < 2 || line[1].Type != TokenNumber {
				return nil, fmt.Errorf("linha %d: REPT requer um número inteiro literal (ex.: REPT 10) -- não pode vir de uma expressão nesta leva", line[0].Line)
			}
			n := line[1].Number
			if n < 0 {
				return nil, fmt.Errorf("linha %d: REPT requer um número >= 0", line[0].Line)
			}
			endIdx, err := findMatchingEndr(lineTokens, i+1)
			if err != nil {
				return nil, fmt.Errorf("linha %d: %w", line[0].Line, err)
			}
			rawBody := lineTokens[i+1 : endIdx]
			for iter := int64(0); iter < n; iter++ {
				*expCounter++
				expanded, err := expandRept(rawBody, expCounter)
				if err != nil {
					return nil, err
				}
				out = append(out, tagLocalLabelsForExpansion(expanded, *expCounter)...)
			}
			i = endIdx + 1
			continue
		}
		out = append(out, line)
		i++
	}
	return out, nil
}

// findMatchingEndr acha o índice do ENDR que fecha o REPT cujo corpo começa
// em "start", respeitando aninhamento (um REPT dentro do corpo incrementa
// a profundidade, seu próprio ENDR decrementa).
func findMatchingEndr(lineTokens [][]Token, start int) (int, error) {
	depth := 1
	for j := start; j < len(lineTokens); j++ {
		if len(lineTokens[j]) == 0 || lineTokens[j][0].Type != TokenIdentifier {
			continue
		}
		switch strings.ToUpper(lineTokens[j][0].Value) {
		case "REPT":
			depth++
		case "ENDR":
			depth--
			if depth == 0 {
				return j, nil
			}
		}
	}
	return -1, fmt.Errorf("REPT sem ENDR correspondente")
}

// tagLocalLabelsForExpansion devolve uma CÓPIA de "lines" com um sufixo
// único (baseado em "tag") acrescentado a toda definição/referência de
// rótulo local (".nome" -> ".nome__expN") -- preserva o "." na frente, pra
// resolveLocalLabels continuar reconhecendo e mesclando normalmente pelo
// rótulo global de verdade, só que agora cada cópia expandida (por REPT
// nesta leva; por MACRO numa leva futura, reaproveitando esta mesma
// função) tem um nome de rótulo local funcionalmente distinto, evitando
// colisão entre cópias. Nunca muta "lines" -- o corpo bruto de um REPT é
// reusado uma vez por iteração, então mutar em lugar corromperia as
// iterações seguintes.
func tagLocalLabelsForExpansion(lines [][]Token, tag int) [][]Token {
	suffix := fmt.Sprintf("__exp%d", tag)
	out := make([][]Token, len(lines))
	for li, line := range lines {
		newLine := make([]Token, len(line))
		copy(newLine, line)
		for i := range newLine {
			if newLine[i].Type == TokenIdentifier && strings.HasPrefix(newLine[i].Value, ".") {
				newLine[i].Value = newLine[i].Value + suffix
			}
		}
		out[li] = newLine
	}
	return out
}

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
