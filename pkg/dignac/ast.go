package dignac

// Node representa um nó genérico na Árvore Sintática Abstrata (AST)
type Node interface {
	node()
}

// Expr representa uma expressão que produz um valor
type Expr interface {
	Node
	expr()
}

// Stmt representa uma declaração ou instrução executável
type Stmt interface {
	Node
	stmt()
}

// ModuleNode representa o módulo principal do BASIC Dignified
type ModuleNode struct {
	Name       string
	Bank       int
	Publics    []string
	Externs    []string
	Globals    []*DimDeclNode
	Procedures []*ProcedureNode
}

func (m *ModuleNode) node() {}

// ParamDecl representa a declaração de um parâmetro de sub-rotina
type ParamDecl struct {
	Name string
	Type string // "INTEGER", "STRING", "SINGLE", "DOUBLE" ou "BOOLEAN"
}

// VarDecl representa uma variável dentro de uma lista LOCAL/DIM, com seu
// próprio tipo -- resolvido pelo sufixo do nome (%/$/!/#) a menos que a
// declaração inteira tenha um "AS <Tipo>" explícito no final, que então
// vale pra todo mundo da lista (ex: "LOCAL a, b AS BOOLEAN").
type VarDecl struct {
	Name string
	Type string
}

// LocalDeclNode representa declaração de variáveis locais (LOCAL a%, b$, c!)
type LocalDeclNode struct {
	Decls []VarDecl
}

func (l *LocalDeclNode) node() {}
func (l *LocalDeclNode) stmt() {}

// DimDeclNode representa declaração de variáveis globais (DIM a%, b$, c!)
type DimDeclNode struct {
	Decls []VarDecl
}

func (d *DimDeclNode) node() {}
func (d *DimDeclNode) stmt() {}

// ProcedureNode representa um procedimento ou função (PROCEDURE / SUB / FUNCTION)
type ProcedureNode struct {
	Name       string
	IsFunction bool
	ReturnType string
	Params     []ParamDecl
	Locals     []*LocalDeclNode
	Body       []Stmt
	IsPublic   bool
}

func (p *ProcedureNode) node() {}
func (p *ProcedureNode) stmt() {}

// --- Declarações Executáveis (Statements) ---

// AssignStmt representa uma atribuição: var = expr
type AssignStmt struct {
	VarName string
	Value   Expr
}

func (s *AssignStmt) node() {}
func (s *AssignStmt) stmt() {}

// ForStmt representa um laço FOR var = start TO end [STEP step] ... NEXT [var]
type ForStmt struct {
	VarName string
	Start   Expr
	End     Expr
	Step    Expr // nil se não especificado (padrão 1)
	Body    []Stmt
}

func (s *ForStmt) node() {}
func (s *ForStmt) stmt() {}

// IfStmt representa uma condicional IF cond THEN ... [ELSE ...] [END IF]
type IfStmt struct {
	Condition Expr
	ThenBody  []Stmt
	ElseBody  []Stmt
}

func (s *IfStmt) node() {}
func (s *IfStmt) stmt() {}

// WhileStmt representa um laço WHILE cond ... WEND
type WhileStmt struct {
	Condition Expr
	Body      []Stmt
}

func (s *WhileStmt) node() {}
func (s *WhileStmt) stmt() {}

// DoLoopStmt representa um laço DO [WHILE cond] ... LOOP
type DoLoopStmt struct {
	Condition Expr
	IsWhile   bool
	Body      []Stmt
}

func (s *DoLoopStmt) node() {}
func (s *DoLoopStmt) stmt() {}

// CallStmt representa a chamada de um procedimento: ProcName(arg1, arg2)
type CallStmt struct {
	Name string
	Args []Expr
}

func (s *CallStmt) node() {}
func (s *CallStmt) stmt() {}

// LineStmt representa a primitiva LINE (x1,y1)-(x2,y2)[, color][, B | BF]
type LineStmt struct {
	X1      Expr
	Y1      Expr
	X2      Expr
	Y2      Expr
	Color   Expr // pode ser nil se omitido
	Box     bool // B
	BoxFill bool // BF
}

func (s *LineStmt) node() {}
func (s *LineStmt) stmt() {}

// PsetStmt representa a primitiva PSET (x, y)[, color]
type PsetStmt struct {
	X     Expr
	Y     Expr
	Color Expr // pode ser nil se omitido
}

func (s *PsetStmt) node() {}
func (s *PsetStmt) stmt() {}

// PrintStmt representa o comando PRINT expr1, expr2; ... ou, com FileNum
// definido, PRINT #n, expr1, expr2; ...
type PrintStmt struct {
	Args              []Expr
	TrailingSemicolon bool
	FileNum           *int // nil = console; caso contrário, número de arquivo (PRINT #n,)
}

func (s *PrintStmt) node() {}
func (s *PrintStmt) stmt() {}

// PutSpriteStmt representa PUT SPRITE index, (x, y), color, pattern
type PutSpriteStmt struct {
	Index   Expr
	X       Expr
	Y       Expr
	Color   Expr
	Pattern Expr
}

func (s *PutSpriteStmt) node() {}
func (s *PutSpriteStmt) stmt() {}

// SpritePatternStmt representa SPRITE PATTERN pattern#, b0, b1, ..., bN
type SpritePatternStmt struct {
	Pattern Expr
	Bytes   []Expr
}

func (s *SpritePatternStmt) node() {}
func (s *SpritePatternStmt) stmt() {}

// SpriteOffStmt representa SPRITE OFF (oculta todos os sprites)
type SpriteOffStmt struct{}

func (s *SpriteOffStmt) node() {}
func (s *SpriteOffStmt) stmt() {}

// PlayStmt representa PLAY "mml" -- a string MML é traduzida em tempo de
// compilação para uma tabela de eventos de PSG_PlaySequence
type PlayStmt struct {
	MML string
}

func (s *PlayStmt) node() {}
func (s *PlayStmt) stmt() {}

// OpenStmt representa OPEN caminho$ FOR modo AS #n
type OpenStmt struct {
	Path    Expr
	Mode    string // "INPUT", "OUTPUT" ou "APPEND"
	FileNum int
}

func (s *OpenStmt) node() {}
func (s *OpenStmt) stmt() {}

// CloseStmt representa CLOSE #n
type CloseStmt struct {
	FileNum int
}

func (s *CloseStmt) node() {}
func (s *CloseStmt) stmt() {}

// ClsStmt representa o comando CLS
type ClsStmt struct{}

func (s *ClsStmt) node() {}
func (s *ClsStmt) stmt() {}

// BeepStmt representa o comando BEEP
type BeepStmt struct{}

func (s *BeepStmt) node() {}
func (s *BeepStmt) stmt() {}

// ScreenStmt representa o comando SCREEN mode
type ScreenStmt struct {
	Mode Expr
}

func (s *ScreenStmt) node() {}
func (s *ScreenStmt) stmt() {}

// ColorStmt representa o comando COLOR fg, bg, bd
type ColorStmt struct {
	Foreground Expr
	Background Expr
	Border     Expr
}

func (s *ColorStmt) node() {}
func (s *ColorStmt) stmt() {}

// ReturnStmt representa o comando RETURN [expr] ou retorno de função
type ReturnStmt struct {
	Value Expr
}

func (s *ReturnStmt) node() {}
func (s *ReturnStmt) stmt() {}

// --- Expressões ---

// NumberExpr representa um número inteiro literal
type NumberExpr struct {
	Value int
}

func (e *NumberExpr) node() {}
func (e *NumberExpr) expr() {}

// FloatExpr representa um literal de ponto flutuante SINGLE (sufixo ! ou
// notação "e") ou DOUBLE (sufixo # ou notação "d"), representado
// internamente em IEEE 754 (binary32/binary64) -- ver a nota de design em
// pkg/dignac/codegen.go sobre por que não é o MBF que o MSX-BASIC real usa.
type FloatExpr struct {
	Value    float64
	IsDouble bool
}

func (e *FloatExpr) node() {}
func (e *FloatExpr) expr() {}

// StringExpr representa uma string literal
type StringExpr struct {
	Value string
}

func (e *StringExpr) node() {}
func (e *StringExpr) expr() {}

// VarExpr representa uma referência a variável ou parâmetro
type VarExpr struct {
	Name string
}

func (e *VarExpr) node() {}
func (e *VarExpr) expr() {}

// BinaryExpr representa uma operação binária (+, -, *, /, MOD, AND, =, etc.)
type BinaryExpr struct {
	Left  Expr
	Op    TokenType
	Right Expr
}

func (e *BinaryExpr) node() {}
func (e *BinaryExpr) expr() {}

// UnaryExpr representa uma operação unária (-, NOT)
type UnaryExpr struct {
	Op   TokenType
	Expr Expr
}

func (e *UnaryExpr) node() {}
func (e *UnaryExpr) expr() {}

// CallExpr representa a chamada de uma função como expressão: Func(a, b)
type CallExpr struct {
	Name string
	Args []Expr
}

func (e *CallExpr) node() {}
func (e *CallExpr) expr() {}
