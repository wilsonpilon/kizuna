package dignac

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser converte o fluxo de tokens em uma AST de MSX-BASIC Dignified
type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
}

// NewParser cria uma nova instância de Parser
func NewParser(lexer *Lexer) (*Parser, error) {
	p := &Parser{lexer: lexer}
	var err error
	p.curToken, err = p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	p.peekToken, err = p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Parser) nextToken() error {
	p.curToken = p.peekToken
	var err error
	p.peekToken, err = p.lexer.NextToken()
	return err
}

func (p *Parser) skipNewlines() error {
	for p.curToken.Type == TokenNewline {
		if err := p.nextToken(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) expect(t TokenType) error {
	if p.curToken.Type != t {
		return fmt.Errorf("esperado token %v, obteve %v (%q) na linha %d:%d", t, p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Column)
	}
	return p.nextToken()
}

// ParseModule analisa um módulo completo Dignified BASIC
func (p *Parser) ParseModule() (*ModuleNode, error) {
	if err := p.skipNewlines(); err != nil {
		return nil, err
	}

	module := &ModuleNode{
		Name:    "MainModule",
		Bank:    0,
		Publics: make([]string, 0),
		Externs: make([]string, 0),
	}

	// Cabeçalho opcional: MODULE <Nome>
	if p.curToken.Type == TokenModule {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type != TokenIdent {
			return nil, fmt.Errorf("esperado nome do módulo após MODULE na linha %d", p.curToken.Line)
		}
		module.Name = p.curToken.Value
		if err := p.nextToken(); err != nil {
			return nil, err
		}
	}

	for p.curToken.Type != TokenEOF {
		if err := p.skipNewlines(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenEOF {
			break
		}

		// END MODULE
		if p.curToken.Type == TokenEnd && (p.peekToken.Type == TokenModule || p.peekToken.Type == TokenEOF) {
			_ = p.nextToken() // consome END
			if p.curToken.Type == TokenModule {
				_ = p.nextToken()
			}
			break
		}

		switch p.curToken.Type {
		case TokenBank:
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			if p.curToken.Type != TokenNumber {
				return nil, fmt.Errorf("esperado número do banco após BANK na linha %d", p.curToken.Line)
			}
			bankNum, err := strconv.Atoi(p.curToken.Value)
			if err != nil {
				return nil, fmt.Errorf("número de banco inválido: %v", err)
			}
			module.Bank = bankNum
			if err := p.nextToken(); err != nil {
				return nil, err
			}

		case TokenPublic:
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			for {
				if p.curToken.Type != TokenIdent {
					return nil, fmt.Errorf("esperado nome de símbolo após PUBLIC na linha %d", p.curToken.Line)
				}
				module.Publics = append(module.Publics, p.curToken.Value)
				if err := p.nextToken(); err != nil {
					return nil, err
				}
				if p.curToken.Type == TokenComma {
					if err := p.nextToken(); err != nil {
						return nil, err
					}
					continue
				}
				break
			}

		case TokenExtern:
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			for {
				if p.curToken.Type != TokenIdent {
					return nil, fmt.Errorf("esperado nome de símbolo após EXTERN na linha %d", p.curToken.Line)
				}
				module.Externs = append(module.Externs, p.curToken.Value)
				if err := p.nextToken(); err != nil {
					return nil, err
				}
				if p.curToken.Type == TokenComma {
					if err := p.nextToken(); err != nil {
						return nil, err
					}
					continue
				}
				break
			}

		case TokenDim:
			dimDecl, err := p.parseDimDecl()
			if err != nil {
				return nil, err
			}
			module.Globals = append(module.Globals, dimDecl)

		case TokenProcedure, TokenSub, TokenFunction:
			proc, err := p.parseProcedure()
			if err != nil {
				return nil, err
			}
			// Verifica se está marcado como público
			for _, pub := range module.Publics {
				if strings.EqualFold(pub, proc.Name) {
					proc.IsPublic = true
					break
				}
			}
			module.Procedures = append(module.Procedures, proc)

		default:
			return nil, fmt.Errorf("declaração inesperada no módulo: %v (%q) na linha %d:%d", p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Column)
		}
	}

	return module, nil
}

func (p *Parser) parseDimDecl() (*DimDeclNode, error) {
	if err := p.nextToken(); err != nil { // consome DIM
		return nil, err
	}

	decl := &DimDeclNode{Type: "INTEGER"}
	for {
		if p.curToken.Type != TokenIdent {
			return nil, fmt.Errorf("esperado nome de variável em DIM na linha %d", p.curToken.Line)
		}
		decl.Vars = append(decl.Vars, p.curToken.Value)
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenComma {
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	if p.curToken.Type == TokenAs {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenInteger || p.curToken.Type == TokenStringKw || p.curToken.Type == TokenBoolean {
			decl.Type = strings.ToUpper(p.curToken.Value)
			if err := p.nextToken(); err != nil {
				return nil, err
			}
		}
	}

	return decl, nil
}

func (p *Parser) parseProcedure() (*ProcedureNode, error) {
	isFunc := p.curToken.Type == TokenFunction
	kindName := p.curToken.Value
	if err := p.nextToken(); err != nil { // consome PROCEDURE / SUB / FUNCTION
		return nil, err
	}

	if p.curToken.Type != TokenIdent {
		return nil, fmt.Errorf("esperado nome após %s na linha %d", kindName, p.curToken.Line)
	}

	proc := &ProcedureNode{
		Name:       p.curToken.Value,
		IsFunction: isFunc,
		ReturnType: "INTEGER",
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	// Parâmetros opcionais: (p1%, p2$ ...)
	if p.curToken.Type == TokenLParen {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		for p.curToken.Type != TokenRParen && p.curToken.Type != TokenEOF {
			if p.curToken.Type != TokenIdent {
				return nil, fmt.Errorf("esperado identificador de parâmetro na linha %d", p.curToken.Line)
			}
			pName := p.curToken.Value
			pType := "INTEGER"
			if strings.HasSuffix(pName, "$") {
				pType = "STRING"
			} else if strings.HasSuffix(pName, "!") {
				pType = "BOOLEAN"
			}

			if err := p.nextToken(); err != nil {
				return nil, err
			}

			if p.curToken.Type == TokenAs {
				if err := p.nextToken(); err != nil {
					return nil, err
				}
				pType = strings.ToUpper(p.curToken.Value)
				if err := p.nextToken(); err != nil {
					return nil, err
				}
			}

			proc.Params = append(proc.Params, ParamDecl{Name: pName, Type: pType})

			if p.curToken.Type == TokenComma {
				if err := p.nextToken(); err != nil {
					return nil, err
				}
				continue
			}
			break
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	// Tipo de retorno opcional da FUNCTION: AS INTEGER
	if isFunc && p.curToken.Type == TokenAs {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		proc.ReturnType = strings.ToUpper(p.curToken.Value)
		if err := p.nextToken(); err != nil {
			return nil, err
		}
	}

	// Corpo do procedimento
	for {
		if err := p.skipNewlines(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenEOF {
			return nil, fmt.Errorf("fim de arquivo inesperado antes do término do procedimento %s", proc.Name)
		}

		// Verifica fechamento: END PROCEDURE / END SUB / END FUNCTION
		if p.curToken.Type == TokenEnd {
			if p.peekToken.Type == TokenProcedure || p.peekToken.Type == TokenSub || p.peekToken.Type == TokenFunction {
				_ = p.nextToken() // consome END
				_ = p.nextToken() // consome PROCEDURE / SUB / FUNCTION
				break
			}
		}

		// Declaração de variáveis locais: LOCAL x%, y%
		if p.curToken.Type == TokenLocal {
			localDecl, err := p.parseLocalDecl()
			if err != nil {
				return nil, err
			}
			proc.Locals = append(proc.Locals, localDecl)
			continue
		}

		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			proc.Body = append(proc.Body, stmt)
		}
	}

	return proc, nil
}

func (p *Parser) parseLocalDecl() (*LocalDeclNode, error) {
	if err := p.nextToken(); err != nil { // consome LOCAL
		return nil, err
	}

	decl := &LocalDeclNode{Type: "INTEGER"}
	for {
		if p.curToken.Type != TokenIdent {
			return nil, fmt.Errorf("esperado nome de variável local após LOCAL na linha %d", p.curToken.Line)
		}
		decl.Vars = append(decl.Vars, p.curToken.Value)
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenComma {
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	if p.curToken.Type == TokenAs {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		decl.Type = strings.ToUpper(p.curToken.Value)
		if err := p.nextToken(); err != nil {
			return nil, err
		}
	}

	return decl, nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	switch p.curToken.Type {
	case TokenFor:
		return p.parseForStmt()
	case TokenIf:
		return p.parseIfStmt()
	case TokenWhile:
		return p.parseWhileStmt()
	case TokenDo:
		return p.parseDoLoopStmt()
	case TokenLine:
		return p.parseLineStmt()
	case TokenPset:
		return p.parsePsetStmt()
	case TokenPrint:
		return p.parsePrintStmt()
	case TokenCls:
		_ = p.nextToken()
		return &ClsStmt{}, nil
	case TokenBeep:
		_ = p.nextToken()
		return &BeepStmt{}, nil
	case TokenScreen:
		_ = p.nextToken()
		mode, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ScreenStmt{Mode: mode}, nil
	case TokenReturn:
		_ = p.nextToken()
		var val Expr
		if p.curToken.Type != TokenNewline && p.curToken.Type != TokenEOF {
			var err error
			val, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
		return &ReturnStmt{Value: val}, nil
	case TokenIdent:
		// Atribuição: var = expr
		// Ou chamada de procedimento: Proc(args)
		if p.peekToken.Type == TokenEqual {
			varName := p.curToken.Value
			_ = p.nextToken() // consome ident
			_ = p.nextToken() // consome =
			val, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &AssignStmt{VarName: varName, Value: val}, nil
		}

		// Chamada de procedimento
		procName := p.curToken.Value
		_ = p.nextToken() // consome ident
		var args []Expr
		if p.curToken.Type == TokenLParen {
			_ = p.nextToken()
			for p.curToken.Type != TokenRParen && p.curToken.Type != TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.curToken.Type == TokenComma {
					_ = p.nextToken()
					continue
				}
				break
			}
			if err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
		}
		return &CallStmt{Name: procName, Args: args}, nil

	default:
		return nil, fmt.Errorf("instrução desconhecida %v (%q) na linha %d:%d", p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Column)
	}
}

// LINE (x1,y1)-(x2,y2)[, color][, B | BF]
func (p *Parser) parseLineStmt() (*LineStmt, error) {
	if err := p.nextToken(); err != nil { // consome LINE
		return nil, err
	}

	// Ponto 1: (x1, y1)
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}
	x1, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenComma); err != nil {
		return nil, err
	}
	y1, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	// Separador '-'
	if err := p.expect(TokenMinus); err != nil {
		return nil, err
	}

	// Ponto 2: (x2, y2)
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}
	x2, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenComma); err != nil {
		return nil, err
	}
	y2, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	stmt := &LineStmt{X1: x1, Y1: y1, X2: x2, Y2: y2}

	// Cor opcional: , color
	if p.curToken.Type == TokenComma {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type != TokenB && p.curToken.Type != TokenBf && p.curToken.Type != TokenComma {
			c, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			stmt.Color = c
		}
	}

	// Modificador opcional: , B ou , BF
	if p.curToken.Type == TokenComma {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenB {
			stmt.Box = true
			_ = p.nextToken()
		} else if p.curToken.Type == TokenBf {
			stmt.BoxFill = true
			_ = p.nextToken()
		}
	}

	return stmt, nil
}

// PSET (x, y)[, color]
func (p *Parser) parsePsetStmt() (*PsetStmt, error) {
	if err := p.nextToken(); err != nil { // consome PSET
		return nil, err
	}

	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}
	x, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenComma); err != nil {
		return nil, err
	}
	y, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	stmt := &PsetStmt{X: x, Y: y}

	// Cor opcional: , color
	if p.curToken.Type == TokenComma {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		c, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Color = c
	}

	return stmt, nil
}

// FOR var = start TO end [STEP step] ... NEXT [var]
func (p *Parser) parseForStmt() (*ForStmt, error) {
	if err := p.nextToken(); err != nil { // consome FOR
		return nil, err
	}

	if p.curToken.Type != TokenIdent {
		return nil, fmt.Errorf("esperado identificador de variável no FOR na linha %d", p.curToken.Line)
	}
	varName := p.curToken.Value
	if err := p.nextToken(); err != nil {
		return nil, err
	}

	if err := p.expect(TokenEqual); err != nil {
		return nil, err
	}

	startExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(TokenTo); err != nil {
		return nil, err
	}

	endExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var stepExpr Expr
	if p.curToken.Type == TokenStep {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		stepExpr, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	stmt := &ForStmt{
		VarName: varName,
		Start:   startExpr,
		End:     endExpr,
		Step:    stepExpr,
	}

	// Corpo do laço
	for {
		if err := p.skipNewlines(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenEOF {
			return nil, fmt.Errorf("esperado NEXT para o laço FOR %s", varName)
		}

		if p.curToken.Type == TokenNext {
			_ = p.nextToken()
			// Nome de variável opcional após NEXT
			if p.curToken.Type == TokenIdent && strings.EqualFold(p.curToken.Value, varName) {
				_ = p.nextToken()
			}
			break
		}

		child, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if child != nil {
			stmt.Body = append(stmt.Body, child)
		}
	}

	return stmt, nil
}

// IF cond THEN stmt [ELSE stmt] OU IF cond THEN \n ... [ELSE ...] END IF
func (p *Parser) parseIfStmt() (*IfStmt, error) {
	if err := p.nextToken(); err != nil { // consome IF
		return nil, err
	}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expect(TokenThen); err != nil {
		return nil, err
	}

	stmt := &IfStmt{Condition: cond}

	// Se houver quebra de linha após THEN, é um bloco multi-linha fechado por END IF
	if p.curToken.Type == TokenNewline {
		for {
			if err := p.skipNewlines(); err != nil {
				return nil, err
			}
			if p.curToken.Type == TokenEOF {
				return nil, fmt.Errorf("esperado END IF para fechar bloco IF")
			}

			if p.curToken.Type == TokenEnd && p.peekToken.Type == TokenIf {
				_ = p.nextToken() // consome END
				_ = p.nextToken() // consome IF
				break
			}

			if p.curToken.Type == TokenElse {
				_ = p.nextToken() // consome ELSE
				for {
					if err := p.skipNewlines(); err != nil {
						return nil, err
					}
					if p.curToken.Type == TokenEnd && p.peekToken.Type == TokenIf {
						_ = p.nextToken() // consome END
						_ = p.nextToken() // consome IF
						break
					}
					elseChild, err := p.parseStmt()
					if err != nil {
						return nil, err
					}
					if elseChild != nil {
						stmt.ElseBody = append(stmt.ElseBody, elseChild)
					}
				}
				break
			}

			child, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if child != nil {
				stmt.ThenBody = append(stmt.ThenBody, child)
			}
		}
	} else {
		// IF de linha única
		thenStmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if thenStmt != nil {
			stmt.ThenBody = append(stmt.ThenBody, thenStmt)
		}

		if p.curToken.Type == TokenElse {
			_ = p.nextToken()
			elseStmt, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			if elseStmt != nil {
				stmt.ElseBody = append(stmt.ElseBody, elseStmt)
			}
		}
	}

	return stmt, nil
}

// WHILE cond ... WEND
func (p *Parser) parseWhileStmt() (*WhileStmt, error) {
	if err := p.nextToken(); err != nil { // consome WHILE
		return nil, err
	}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	stmt := &WhileStmt{Condition: cond}

	for {
		if err := p.skipNewlines(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenEOF {
			return nil, fmt.Errorf("esperado WEND para fechar laço WHILE")
		}

		if p.curToken.Type == TokenWend {
			_ = p.nextToken()
			break
		}

		child, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if child != nil {
			stmt.Body = append(stmt.Body, child)
		}
	}

	return stmt, nil
}

// DO [WHILE cond] ... LOOP
func (p *Parser) parseDoLoopStmt() (*DoLoopStmt, error) {
	if err := p.nextToken(); err != nil { // consome DO
		return nil, err
	}

	stmt := &DoLoopStmt{IsWhile: true}
	if p.curToken.Type == TokenWhile {
		_ = p.nextToken()
		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Condition = cond
	}

	for {
		if err := p.skipNewlines(); err != nil {
			return nil, err
		}
		if p.curToken.Type == TokenEOF {
			return nil, fmt.Errorf("esperado LOOP para fechar laço DO")
		}

		if p.curToken.Type == TokenLoop {
			_ = p.nextToken()
			break
		}

		child, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if child != nil {
			stmt.Body = append(stmt.Body, child)
		}
	}

	return stmt, nil
}

// PRINT expr1, expr2; ...
func (p *Parser) parsePrintStmt() (*PrintStmt, error) {
	if err := p.nextToken(); err != nil { // consome PRINT
		return nil, err
	}

	stmt := &PrintStmt{}

	for p.curToken.Type != TokenNewline && p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenSemi {
			stmt.TrailingSemicolon = true
			_ = p.nextToken()
			continue
		}
		if p.curToken.Type == TokenComma {
			stmt.TrailingSemicolon = false
			_ = p.nextToken()
			continue
		}

		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Args = append(stmt.Args, arg)
		stmt.TrailingSemicolon = false

		if p.curToken.Type == TokenSemi {
			stmt.TrailingSemicolon = true
			_ = p.nextToken()
		} else if p.curToken.Type == TokenComma {
			_ = p.nextToken()
		}
	}

	return stmt, nil
}

// --- Análise de Expressões (Precedência de Operadores) ---

func (p *Parser) parseExpression() (Expr, error) {
	return p.parseRelational()
}

// Operadores relacionais: =, <>, <, <=, >, >=
func (p *Parser) parseRelational() (Expr, error) {
	left, err := p.parseLogical()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TokenEqual || p.curToken.Type == TokenNotEqual ||
		p.curToken.Type == TokenLess || p.curToken.Type == TokenLessEq ||
		p.curToken.Type == TokenGreater || p.curToken.Type == TokenGreaterEq {
		op := p.curToken.Type
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		right, err := p.parseLogical()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}

	return left, nil
}

// Operadores lógicos: OR, XOR, AND
func (p *Parser) parseLogical() (Expr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TokenOr || p.curToken.Type == TokenXor || p.curToken.Type == TokenAnd {
		op := p.curToken.Type
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}

	return left, nil
}

// Adição e Subtração: +, -
func (p *Parser) parseAdditive() (Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TokenPlus || p.curToken.Type == TokenMinus {
		op := p.curToken.Type
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}

	return left, nil
}

// Multiplicação, Divisão e MOD: *, /, \, MOD
func (p *Parser) parseMultiplicative() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TokenMul || p.curToken.Type == TokenDiv || p.curToken.Type == TokenMod {
		op := p.curToken.Type
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}

	return left, nil
}

// Unários: -, NOT
func (p *Parser) parseUnary() (Expr, error) {
	if p.curToken.Type == TokenMinus || p.curToken.Type == TokenNot {
		op := p.curToken.Type
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Expr: expr}, nil
	}

	return p.parsePrimary()
}

// Primários: número, string, variável, chamada ou (expr)
func (p *Parser) parsePrimary() (Expr, error) {
	switch p.curToken.Type {
	case TokenNumber:
		valStr := p.curToken.Value
		var val int
		var err error
		if strings.HasPrefix(valStr, "&H") || strings.HasPrefix(valStr, "&h") {
			var parsed int64
			parsed, err = strconv.ParseInt(valStr[2:], 16, 32)
			val = int(parsed)
		} else {
			val, err = strconv.Atoi(valStr)
		}
		if err != nil {
			return nil, fmt.Errorf("número inválido: %v", err)
		}
		_ = p.nextToken()
		return &NumberExpr{Value: val}, nil

	case TokenString:
		str := p.curToken.Value
		_ = p.nextToken()
		return &StringExpr{Value: str}, nil

	case TokenIdent:
		ident := p.curToken.Value
		_ = p.nextToken()

		// Se seguido por parênteses, é uma chamada de função
		if p.curToken.Type == TokenLParen {
			_ = p.nextToken()
			var args []Expr
			for p.curToken.Type != TokenRParen && p.curToken.Type != TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.curToken.Type == TokenComma {
					_ = p.nextToken()
					continue
				}
				break
			}
			if err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
			return &CallExpr{Name: ident, Args: args}, nil
		}

		return &VarExpr{Name: ident}, nil

	case TokenLParen:
		_ = p.nextToken() // consome (
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, fmt.Errorf("expressão inválida com token %v (%q) na linha %d:%d", p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Column)
	}
}
