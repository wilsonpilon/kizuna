package wirth80

import (
	"fmt"
	"strings"
)

// Parser analisa os tokens e constrói a árvore sintática (AST)
type Parser struct {
	lexer   *Lexer
	current Token
	peekTok Token
}

// NewParser cria um novo Parser para a fonte fornecida
func NewParser(lexer *Lexer) (*Parser, error) {
	p := &Parser{lexer: lexer}
	// Inicializa current e peekTok
	var err error
	p.current, err = p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	p.peekTok, err = p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Parser) advance() error {
	p.current = p.peekTok
	var err error
	p.peekTok, err = p.lexer.NextToken()
	return err
}

func (p *Parser) expect(tt TokenType) (Token, error) {
	if p.current.Type != tt {
		return Token{}, fmt.Errorf("esperado '%s', encontrado '%s' na linha %d:%d",
			tt, p.current.Value, p.current.Line, p.current.Column)
	}
	tok := p.current
	err := p.advance()
	return tok, err
}

func (p *Parser) match(tt TokenType) bool {
	if p.current.Type == tt {
		_ = p.advance()
		return true
	}
	return false
}

// ParseProgram analisa o programa Pascal completo
func (p *Parser) ParseProgram() (*ProgramNode, error) {
	prog := &ProgramNode{
		Line:   p.current.Line,
		Column: p.current.Column,
	}

	// 1. Cabeçalho opcional: program Nome;
	if p.match(TokenProgram) {
		nameTok, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		prog.Name = nameTok.Value
		if _, err := p.expect(TokenSemi); err != nil {
			return nil, err
		}
	} else {
		prog.Name = "Program"
	}

	// 2. BANK <n> opcional, mesma posição do KAJI80/DIGNAC (logo após o
	// cabeçalho, antes de PUBLIC/EXTERN) -- terminado em ';' como qualquer
	// declaração Pascal. Sem BANK, o módulo cai no banco comum (0), como
	// sempre foi.
	if p.match(TokenBank) {
		if p.current.Type != TokenNumber {
			return nil, fmt.Errorf("esperado número do banco após 'BANK' na linha %d:%d", p.current.Line, p.current.Column)
		}
		prog.Bank = int(p.current.Number)
		if err := p.advance(); err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenSemi); err != nil {
			return nil, err
		}
	}

	// 3. PUBLIC/EXTERN opcionais, na mesma posição e formato do KAJI80/DIGNAC
	// (mas terminados em ';', como qualquer declaração Pascal) -- só nomes de
	// procedure/function fazem sentido aqui, já que só rotina é chamável
	// entre módulos.
	for p.current.Type == TokenPublic || p.current.Type == TokenExtern {
		isPublic := p.current.Type == TokenPublic
		if err := p.advance(); err != nil {
			return nil, err
		}
		for {
			nameTok, err := p.expect(TokenIdent)
			if err != nil {
				return nil, err
			}
			if isPublic {
				prog.Publics = append(prog.Publics, nameTok.Value)
			} else {
				prog.Externs = append(prog.Externs, nameTok.Value)
			}
			if p.match(TokenComma) {
				continue
			}
			break
		}
		if _, err := p.expect(TokenSemi); err != nil {
			return nil, err
		}
	}

	// 4. Declarações de variáveis opcionais: var ...
	if p.match(TokenVar) {
		for p.current.Type == TokenIdent {
			decl, err := p.parseVarDecl()
			if err != nil {
				return nil, err
			}
			prog.Vars = append(prog.Vars, decl)
		}
	}

	// 5. procedure/function definidas pelo usuário, na ordem declarada
	// (declare-antes-de-usar: uma só pode chamar outra já declarada antes
	// dela, sem forward declarations nesta leva).
	for p.current.Type == TokenProcedure || p.current.Type == TokenFunction {
		proc, err := p.parseProcDecl()
		if err != nil {
			return nil, err
		}
		prog.Procs = append(prog.Procs, proc)
	}

	// 6. Bloco principal: begin ... end.
	if p.current.Type != TokenBegin {
		return nil, fmt.Errorf("esperado 'begin' do programa na linha %d:%d", p.current.Line, p.current.Column)
	}
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	prog.Block = block

	// 7. Ponto final '.'
	if _, err := p.expect(TokenDot); err != nil {
		return nil, fmt.Errorf("esperado '.' no final do programa: %w", err)
	}

	return prog, nil
}

// parseVarDecl analisa: a, b, c: Integer;
func (p *Parser) parseVarDecl() (*VarDecl, error) {
	decl := &VarDecl{
		Line:   p.current.Line,
		Column: p.current.Column,
	}

	for {
		nameTok, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		decl.Names = append(decl.Names, nameTok.Value)

		if p.match(TokenComma) {
			continue
		}
		break
	}

	if _, err := p.expect(TokenColon); err != nil {
		return nil, err
	}

	// Tipo (Integer, Char, Boolean, String)
	typeTok := p.current
	if typeTok.Type != TokenIdent && typeTok.Type != TokenInteger && typeTok.Type != TokenChar &&
		typeTok.Type != TokenBoolean && typeTok.Type != TokenStringKw {
		return nil, fmt.Errorf("tipo inválido '%s' na declaração de variável na linha %d:%d",
			typeTok.Value, typeTok.Line, typeTok.Column)
	}
	decl.Type = typeTok.Value
	_ = p.advance()

	if _, err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return decl, nil
}

// parseProcDecl analisa uma declaração completa de procedure/function:
//
//	procedure Nome(a, b: Integer; c: Char);
//	var local: Integer;
//	begin ... end;
//
// function segue o mesmo formato com ": Tipo" antes do ';' final do
// cabeçalho, e devolve valor por atribuição ao próprio nome dentro do
// corpo (Nome := expr;) -- estilo Turbo Pascal clássico, não "return".
func (p *Parser) parseProcDecl() (*ProcDecl, error) {
	decl := &ProcDecl{
		Line:       p.current.Line,
		Column:     p.current.Column,
		IsFunction: p.current.Type == TokenFunction,
	}
	if err := p.advance(); err != nil { // consome procedure/function
		return nil, err
	}

	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	decl.Name = nameTok.Value

	// Lista de parâmetros opcional: grupos "nomes: Tipo" separados por ';'
	if p.match(TokenLParen) {
		for p.current.Type != TokenRParen {
			var names []string
			for {
				nTok, err := p.expect(TokenIdent)
				if err != nil {
					return nil, err
				}
				names = append(names, nTok.Value)
				if p.match(TokenComma) {
					continue
				}
				break
			}
			if _, err := p.expect(TokenColon); err != nil {
				return nil, err
			}
			typeTok := p.current
			if typeTok.Type != TokenIdent && typeTok.Type != TokenInteger && typeTok.Type != TokenChar &&
				typeTok.Type != TokenBoolean && typeTok.Type != TokenStringKw {
				return nil, fmt.Errorf("tipo inválido '%s' em parâmetro na linha %d:%d",
					typeTok.Value, typeTok.Line, typeTok.Column)
			}
			if err := p.advance(); err != nil {
				return nil, err
			}
			for _, n := range names {
				decl.Params = append(decl.Params, ParamDecl{Name: n, Type: typeTok.Value})
			}
			if p.match(TokenSemi) {
				continue
			}
			break
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	if decl.IsFunction {
		if _, err := p.expect(TokenColon); err != nil {
			return nil, err
		}
		typeTok := p.current
		if typeTok.Type != TokenIdent && typeTok.Type != TokenInteger && typeTok.Type != TokenChar &&
			typeTok.Type != TokenBoolean && typeTok.Type != TokenStringKw {
			return nil, fmt.Errorf("tipo de retorno inválido '%s' na linha %d:%d",
				typeTok.Value, typeTok.Line, typeTok.Column)
		}
		decl.ReturnType = typeTok.Value
		if err := p.advance(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	// var local opcional
	if p.match(TokenVar) {
		for p.current.Type == TokenIdent {
			local, err := p.parseVarDecl()
			if err != nil {
				return nil, err
			}
			decl.Locals = append(decl.Locals, local)
		}
	}

	if p.current.Type != TokenBegin {
		return nil, fmt.Errorf("esperado 'begin' no corpo de '%s' na linha %d:%d", decl.Name, p.current.Line, p.current.Column)
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	decl.Body = body

	if _, err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return decl, nil
}

// parseBlock analisa: begin stmt; stmt; ... end
func (p *Parser) parseBlock() (*BlockStmt, error) {
	block := &BlockStmt{
		Line:   p.current.Line,
		Column: p.current.Column,
	}

	if _, err := p.expect(TokenBegin); err != nil {
		return nil, err
	}

	for p.current.Type != TokenEnd && p.current.Type != TokenEOF {
		// Pular ponto-e-vírgula avulso
		if p.match(TokenSemi) {
			continue
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// Ponto e vírgula após statement é opcional logo antes de 'end'
		if p.current.Type != TokenEnd {
			if _, err := p.expect(TokenSemi); err != nil {
				return nil, err
			}
		}
	}

	if _, err := p.expect(TokenEnd); err != nil {
		return nil, err
	}

	return block, nil
}

func (p *Parser) parseStatement() (Stmt, error) {
	switch p.current.Type {
	case TokenBegin:
		return p.parseBlock()

	case TokenWrite, TokenWriteLn:
		isNewLine := p.current.Type == TokenWriteLn
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		var args []Expr
		if p.match(TokenLParen) {
			for p.current.Type != TokenRParen && p.current.Type != TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)

				if p.match(TokenComma) {
					continue
				}
				break
			}
			if _, err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
		}
		return &WriteStmt{Args: args, NewLine: isNewLine, Line: line, Column: col}, nil

	case TokenIdent:
		name := p.current.Value
		line := p.current.Line
		col := p.current.Column

		if p.peekTok.Type == TokenLParen {
			// Chamada de procedure como comando: Nome(args)
			if err := p.advance(); err != nil { // consome ident
				return nil, err
			}
			_ = p.advance() // consome '('
			var args []Expr
			for p.current.Type != TokenRParen && p.current.Type != TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.match(TokenComma) {
					continue
				}
				break
			}
			if _, err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
			return &CallStmt{Name: name, Args: args, Line: line, Column: col}, nil
		}

		// Atribuição: var := expr (ou NomeDaFuncao := expr dentro do corpo
		// de uma function, mesma sintaxe -- o codegen decide o significado)
		_ = p.advance()

		if _, err := p.expect(TokenAssign); err != nil {
			return nil, err
		}

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &AssignStmt{VarName: name, Expr: expr, Line: line, Column: col}, nil

	case TokenIf:
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenThen); err != nil {
			return nil, err
		}

		thenStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		var elseStmt Stmt
		if p.match(TokenElse) {
			elseStmt, err = p.parseStatement()
			if err != nil {
				return nil, err
			}
		}
		return &IfStmt{Cond: cond, Then: thenStmt, Else: elseStmt, Line: line, Column: col}, nil

	case TokenWhile:
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenDo); err != nil {
			return nil, err
		}

		body, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return &WhileStmt{Cond: cond, Body: body, Line: line, Column: col}, nil

	default:
		return nil, fmt.Errorf("comando inesperado '%s' na linha %d:%d",
			p.current.Value, p.current.Line, p.current.Column)
	}
}

// Precedência de expressões:
// parseExpression -> comparações (=, <>, <, <=, >, >=)
// parseAdditive   -> soma e subtração (+, -)
// parseMultiplicative -> multiplicação e divisão (*, div, /)
// parseUnary      -> unário (+, -)
// parsePrimary    -> número, string, variável, parênteses

func (p *Parser) parseExpression() (Expr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenEqual || p.current.Type == TokenNotEqual ||
		p.current.Type == TokenLess || p.current.Type == TokenLessEq ||
		p.current.Type == TokenGreater || p.current.Type == TokenGreaterEq {
		op := p.current.Type
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right, Line: line, Column: col}
	}

	return left, nil
}

func (p *Parser) parseAdditive() (Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenPlus || p.current.Type == TokenMinus {
		op := p.current.Type
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right, Line: line, Column: col}
	}

	return left, nil
}

func (p *Parser) parseMultiplicative() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenMul || p.current.Type == TokenDiv {
		op := p.current.Type
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right, Line: line, Column: col}
	}

	return left, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if p.current.Type == TokenPlus || p.current.Type == TokenMinus {
		op := p.current.Type
		line := p.current.Line
		col := p.current.Column
		_ = p.advance()

		sub, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Expr: sub, Line: line, Column: col}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	tok := p.current

	switch tok.Type {
	case TokenNumber:
		_ = p.advance()
		return &NumberLiteral{Value: tok.Number, Line: tok.Line, Column: tok.Column}, nil

	case TokenString:
		_ = p.advance()
		return &StringLiteral{Value: tok.Value, Line: tok.Line, Column: tok.Column}, nil

	case TokenIdent:
		_ = p.advance()
		if p.match(TokenLParen) {
			// Chamada de function como expressão: Nome(args)
			var args []Expr
			for p.current.Type != TokenRParen && p.current.Type != TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.match(TokenComma) {
					continue
				}
				break
			}
			if _, err := p.expect(TokenRParen); err != nil {
				return nil, err
			}
			return &CallExpr{Name: tok.Value, Args: args, Line: tok.Line, Column: tok.Column}, nil
		}
		return &VarExpr{Name: tok.Value, Line: tok.Line, Column: tok.Column}, nil

	case TokenLParen:
		_ = p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, fmt.Errorf("expressão inesperada '%s' na linha %d:%d",
			strings.TrimSpace(tok.Value), tok.Line, tok.Column)
	}
}
