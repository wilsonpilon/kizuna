package wirth80

// Node é a interface base para todos os nós da árvore sintática
type Node interface {
	Pos() (int, int)
}

// Expr representa uma expressão que produz um valor
type Expr interface {
	Node
	exprNode()
}

// Stmt representa um comando executável
type Stmt interface {
	Node
	stmtNode()
}

// ProgramNode representa o programa Pascal completo
type ProgramNode struct {
	Name   string
	Vars   []*VarDecl
	Block  *BlockStmt
	Line   int
	Column int
}

func (p *ProgramNode) Pos() (int, int) { return p.Line, p.Column }

// VarDecl representa a declaração de uma ou mais variáveis de um tipo
type VarDecl struct {
	Names  []string
	Type   string // "Integer", "Char", "Boolean", "String"
	Line   int
	Column int
}

func (v *VarDecl) Pos() (int, int) { return v.Line, v.Column }

// BlockStmt representa um bloco de comandos (begin ... end)
type BlockStmt struct {
	Statements []Stmt
	Line       int
	Column     int
}

func (b *BlockStmt) Pos() (int, int) { return b.Line, b.Column }
func (b *BlockStmt) stmtNode()        {}

// AssignStmt representa uma atribuição (var := expr)
type AssignStmt struct {
	VarName string
	Expr    Expr
	Line    int
	Column  int
}

func (a *AssignStmt) Pos() (int, int) { return a.Line, a.Column }
func (a *AssignStmt) stmtNode()        {}

// WriteStmt representa uma chamada Write(...) ou WriteLn(...)
type WriteStmt struct {
	Args    []Expr
	NewLine bool
	Line    int
	Column  int
}

func (w *WriteStmt) Pos() (int, int) { return w.Line, w.Column }
func (w *WriteStmt) stmtNode()        {}

// IfStmt representa uma condicional if cond then stmt [else stmt]
type IfStmt struct {
	Cond   Expr
	Then   Stmt
	Else   Stmt
	Line   int
	Column int
}

func (i *IfStmt) Pos() (int, int) { return i.Line, i.Column }
func (i *IfStmt) stmtNode()        {}

// WhileStmt representa um laço while cond do stmt
type WhileStmt struct {
	Cond   Expr
	Body   Stmt
	Line   int
	Column int
}

func (w *WhileStmt) Pos() (int, int) { return w.Line, w.Column }
func (w *WhileStmt) stmtNode()        {}

// BinaryExpr representa uma operação binária (+, -, *, div, =, <>, <, <=, >, >=)
type BinaryExpr struct {
	Left   Expr
	Op     TokenType
	Right  Expr
	Line   int
	Column int
}

func (b *BinaryExpr) Pos() (int, int) { return b.Line, b.Column }
func (b *BinaryExpr) exprNode()        {}

// UnaryExpr representa uma operação unária (+ ou -)
type UnaryExpr struct {
	Op     TokenType
	Expr   Expr
	Line   int
	Column int
}

func (u *UnaryExpr) Pos() (int, int) { return u.Line, u.Column }
func (u *UnaryExpr) exprNode()        {}

// NumberLiteral representa uma constante inteira
type NumberLiteral struct {
	Value  int64
	Line   int
	Column int
}

func (n *NumberLiteral) Pos() (int, int) { return n.Line, n.Column }
func (n *NumberLiteral) exprNode()        {}

// StringLiteral representa um literal de texto ('...')
type StringLiteral struct {
	Value  string
	Line   int
	Column int
}

func (s *StringLiteral) Pos() (int, int) { return s.Line, s.Column }
func (s *StringLiteral) exprNode()        {}

// VarExpr representa o acesso a uma variável por nome
type VarExpr struct {
	Name   string
	Line   int
	Column int
}

func (v *VarExpr) Pos() (int, int) { return v.Line, v.Column }
func (v *VarExpr) exprNode()        {}
