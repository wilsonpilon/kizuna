package kaji80

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// =============================================================================
// INCLUDE "arquivo"
//
// Passo em TEXTO BRUTO, antes do lexer: uma linha "INCLUDE \"x.inc\"" é
// substituída pelo conteúdo do arquivo (recursivamente). Existe pra
// compartilhar constantes (EQU) e macros entre módulos sem repetir tudo em
// cada arquivo -- essencial pra uma biblioteca com dezenas/centenas de
// módulos pequenos (um por rotina) que precisam dos mesmos endereços de
// porta/registrador.
//
// O caminho é resolvido relativo ao diretório do arquivo que contém o
// INCLUDE (o arquivo-fonte de nível mais alto usa Assembler.SetBaseDir).
// Inclusão circular e aninhamento excessivo são erro claro.
//
// Limitação conhecida: depois da expansão, "linha N" numa mensagem de erro
// passa a contar linhas do texto JÁ expandido, não do arquivo original --
// só afeta arquivos que usam INCLUDE.
// =============================================================================

const maxIncludeDepth = 16

// parseIncludeLine reconhece uma linha de INCLUDE e devolve o caminho.
func parseIncludeLine(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if len(t) < 9 || !strings.EqualFold(t[:7], "INCLUDE") || (t[7] != ' ' && t[7] != '\t') {
		return "", false
	}
	rest := strings.TrimSpace(t[7:])
	if len(rest) < 2 || (rest[0] != '"' && rest[0] != '\'') {
		return "", false
	}
	q := rest[0]
	end := strings.IndexByte(rest[1:], q)
	if end < 0 {
		return "", false
	}
	tail := strings.TrimSpace(rest[end+2:])
	if tail != "" && tail[0] != ';' {
		return "", false
	}
	return rest[1 : 1+end], true
}

func expandIncludes(source, baseDir string, stack []string) (string, error) {
	if len(stack) > maxIncludeDepth {
		return "", fmt.Errorf("INCLUDE aninhado além de %d níveis -- provável inclusão circular", maxIncludeDepth)
	}
	if !strings.Contains(strings.ToUpper(source), "INCLUDE") {
		return source, nil
	}
	lines := strings.Split(source, "\n")
	var out strings.Builder
	for i, line := range lines {
		path, ok := parseIncludeLine(line)
		if !ok {
			out.WriteString(line)
		} else {
			resolved := path
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(baseDir, path)
			}
			abs, err := filepath.Abs(resolved)
			if err != nil {
				abs = resolved
			}
			for _, s := range stack {
				if s == abs {
					return "", fmt.Errorf("INCLUDE circular: '%s' já está sendo incluído", path)
				}
			}
			data, err := os.ReadFile(resolved)
			if err != nil {
				return "", fmt.Errorf("INCLUDE: erro ao ler '%s': %w", path, err)
			}
			sub, err := expandIncludes(string(data), filepath.Dir(resolved), append(stack, abs))
			if err != nil {
				return "", err
			}
			out.WriteString(strings.TrimRight(sub, "\r\n"))
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String(), nil
}

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

// =============================================================================
// MACRO @param, .../ENDM
//
// Marcador de parâmetro: "@nome", não "#nome" como no asMSX -- "#" já é o
// prefixo de literal hexadecimal do KAJI80 (#100 = 256, documentado), e
// "@nome" já é tokenizado pelo lexer como UM identificador só (isIdentStart/
// isIdentPart já incluem '@', mesmo padrão usado pro '.' dos rótulos
// locais), sem precisar de nenhuma mudança no lexer. Mesmo espírito da
// escolha de MOD em vez de '%' na Fase 1: "nossa própria sintaxe", não uma
// cópia literal do asMSX.
//
// A substituição de parâmetro precisa funcionar mesmo DENTRO de um
// identificador maior (ex. real da documentação do asMSX, só trocando '#'
// por '@': ".noreset_@VARIABLE:" chamado com VARNAME vira
// ".noreset_VARNAME:") -- como '.', letras, dígitos, '_' e '@' são TODOS
// caracteres válidos de identificador pro lexer, ".noreset_@VARIABLE"
// chega como UM token só, não vários. Por isso, diferente de REPT/IF/
// rótulos locais (que operam só reescrevendo Value de tokens inteiros já
// existentes), a expansão de macro reconstrói cada linha do corpo de volta
// em texto (depois de substituir @param por dentro do Value de cada
// token), e relexa esse texto do zero -- garante que um argumento
// multi-token (ex.: uma expressão "10+20" passada como argumento) também
// vira tokens corretos de verdade (NUMBER, PLUS, NUMBER), não um Value
// só com texto misturado.
// =============================================================================

const maxMacroExpansionDepth = 64

type macroDef struct {
	params []string // nomes SEM o '@' na frente
	body   [][]Token
}

// expandMacros processa "Nome: MACRO @p1, @p2, .../ENDM" (registra, não
// expande na hora) e expande cada invocação "Nome arg1, arg2, ..."
// encontrada depois -- exige declarar antes de usar, mesma regra já usada
// no resto do projeto (DIGNAC/WIRTH80). O corpo de uma macro pode conter
// outra chamada de macro (expandida recursivamente, com guarda de
// profundidade contra auto-referência infinita) -- IF/REPT dentro do
// corpo já funcionam de graça, porque MACRO roda depois deles no pipeline
// (ver tokenizeLines em assembler.go), então já chegam aqui como texto
// final.
func (a *Assembler) expandMacros(lineTokens [][]Token, defs map[string]*macroDef, expCounter *int, depth int) ([][]Token, error) {
	if depth > maxMacroExpansionDepth {
		return nil, fmt.Errorf("expansão de macro excedeu %d níveis -- provável auto-referência infinita", maxMacroExpansionDepth)
	}

	var out [][]Token
	i := 0
	for i < len(lineTokens) {
		line := lineTokens[i]
		if len(line) == 0 {
			i++
			continue
		}

		// Definição: "Nome: MACRO @p1, @p2, ..."
		if len(line) >= 3 && line[0].Type == TokenIdentifier && line[1].Type == TokenColon &&
			line[2].Type == TokenIdentifier && strings.EqualFold(line[2].Value, "MACRO") {
			var params []string
			for _, t := range line[3:] {
				if t.Type == TokenIdentifier && strings.HasPrefix(t.Value, "@") {
					params = append(params, t.Value[1:])
				}
			}
			endIdx, err := findMatchingEndm(lineTokens, i+1)
			if err != nil {
				return nil, fmt.Errorf("linha %d: %w", line[0].Line, err)
			}
			defs[strings.ToUpper(line[0].Value)] = &macroDef{params: params, body: lineTokens[i+1 : endIdx]}
			i = endIdx + 1
			continue
		}

		// Chamada: "Nome arg1, arg2, ..."
		if line[0].Type == TokenIdentifier {
			if def, ok := defs[strings.ToUpper(line[0].Value)]; ok {
				pl, err := a.parseLine(line)
				if err != nil {
					return nil, err
				}
				if len(pl.operands) != len(def.params) {
					return nil, fmt.Errorf("linha %d: macro '%s' espera %d argumento(s), recebeu %d", line[0].Line, line[0].Value, len(def.params), len(pl.operands))
				}
				*expCounter++
				substituted := substituteMacroParams(def.body, def.params, pl.operands)
				relexed, err := relexLines(substituted)
				if err != nil {
					return nil, fmt.Errorf("linha %d: macro '%s': %w", line[0].Line, line[0].Value, err)
				}
				tagged := tagLocalLabelsForExpansion(relexed, *expCounter)
				expanded, err := a.expandMacros(tagged, defs, expCounter, depth+1)
				if err != nil {
					return nil, err
				}
				out = append(out, expanded...)
				i++
				continue
			}
		}

		out = append(out, line)
		i++
	}
	return out, nil
}

// findMatchingEndm acha o índice do ENDM que fecha a declaração de macro
// cujo corpo começa em "start". Declarações de MACRO não aninham (só
// chamadas aninham, tratadas por recursão em expandMacros), então não
// precisa rastrear profundidade como findMatchingEndr faz pro REPT.
func findMatchingEndm(lineTokens [][]Token, start int) (int, error) {
	for j := start; j < len(lineTokens); j++ {
		if len(lineTokens[j]) == 0 {
			continue
		}
		if lineTokens[j][0].Type == TokenIdentifier && strings.EqualFold(lineTokens[j][0].Value, "ENDM") {
			return j, nil
		}
	}
	return -1, fmt.Errorf("MACRO sem ENDM correspondente")
}

// substituteMacroParams devolve uma CÓPIA do corpo com cada "@param"
// substituído pelo texto do argumento correspondente, em qualquer lugar
// que apareça dentro do Value de um token (inclusive no meio de um
// identificador maior). Substitui os parâmetros de nome mais LONGO
// primeiro, pra um parâmetro cujo nome é prefixo de outro (ex.: @VAR e
// @VARIABLE declarados juntos) não ser trocado por engano dentro do nome
// maior antes da vez dele.
func substituteMacroParams(body [][]Token, params []string, args []string) [][]Token {
	order := make([]int, len(params))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(x, y int) bool { return len(params[order[x]]) > len(params[order[y]]) })

	out := make([][]Token, len(body))
	for li, line := range body {
		newLine := make([]Token, len(line))
		for ti, tok := range line {
			v := tok.Value
			for _, idx := range order {
				marker := "@" + params[idx]
				if strings.Contains(v, marker) {
					v = strings.ReplaceAll(v, marker, args[idx])
				}
			}
			newTok := tok
			newTok.Value = v
			newLine[ti] = newTok
		}
		out[li] = newLine
	}
	return out
}

// relexLines reconstrói cada linha (já com os parâmetros substituídos) de
// volta em texto e relexa do zero -- necessário pra um argumento
// multi-token (ex.: uma expressão passada como argumento) virar tokens de
// verdade, não um Value de token só com texto misturado.
func relexLines(lines [][]Token) ([][]Token, error) {
	out := make([][]Token, 0, len(lines))
	for _, line := range lines {
		text := tokensToText(line)
		if strings.TrimSpace(text) == "" {
			continue
		}
		lexer := NewLexer(text)
		var toks []Token
		for {
			tok, err := lexer.NextToken()
			if err != nil {
				return nil, fmt.Errorf("relexando %q: %w", text, err)
			}
			if tok.Type == TokenEOF || tok.Type == TokenNewline {
				break
			}
			toks = append(toks, tok)
		}
		if len(toks) > 0 {
			out = append(out, toks)
		}
	}
	return out, nil
}

// tokensToText reconstrói o texto-fonte aproximado de uma linha de tokens
// -- espaçamento exato não importa (o lexer ignora espaço em branco fora
// de string), só recolocar as aspas de um literal de string/caractere
// (TokenString já vem sem elas, igual ao resto do assembler já faz em
// parseLine).
func tokensToText(line []Token) string {
	var sb strings.Builder
	for i, t := range line {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if t.Type == TokenString {
			sb.WriteByte('"')
			sb.WriteString(t.Value)
			sb.WriteByte('"')
		} else {
			sb.WriteString(t.Value)
		}
	}
	return sb.String()
}
