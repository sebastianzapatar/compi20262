// =============================================================================
// PAQUETE ast (Abstract Syntax Tree - Árbol de Sintaxis Abstracta)
// =============================================================================
//
// ¿QUÉ ES ESTE ARCHIVO?
// ----------------------
// Este archivo define la ESTRUCTURA de nuestro lenguaje de programación.
// Piensa en él como el "plano arquitectónico" de todo lo que podemos escribir
// en CatCompiler.
//
// ¿QUÉ ES UN AST?
// -----------------
// Cuando escribimos código como `let x = 5 + 3;`, el lexer lo convierte en
// tokens sueltos: [LET, IDENT("x"), ASSIGN, INT(5), PLUS, INT(3), SEMICOLON]
//
// Pero eso es solo una lista plana. El AST toma esa lista y la convierte en
// un ÁRBOL con jerarquía y significado:
//
//            LetStatement
//            /          \
//      Name: "x"    Value: InfixExpression
//                        /     |      \
//                    Left:5  Op:"+"  Right:3
//
// Cada "cajita" del árbol es un NODO. Este archivo define todos los tipos
// de nodos posibles en nuestro lenguaje.
//
// ¿POR QUÉ NECESITAMOS UN ÁRBOL?
// --------------------------------
// Porque un árbol nos permite representar la PRECEDENCIA (jerarquía) de las
// operaciones. En `1 + 2 * 3`, el árbol pone `2 * 3` como un sub-árbol
// que se resuelve primero, y luego `1 + resultado`. Sin un árbol, no
// podríamos saber qué operación va primero.
//
// =============================================================================
package ast

import (
	"bytes"
	"compigo/token"
	"strings"
)

// =============================================================================
// INTERFACES PRINCIPALES
// =============================================================================
//
// En Go, una interfaz define un "contrato": cualquier estructura que implemente
// los métodos listados en la interfaz, automáticamente "cumple" esa interfaz.
//
// Definimos 3 interfaces que clasifican todo en nuestro árbol:
//   1. Node       → La base de todo (todos son nodos)
//   2. Statement  → Instrucciones que NO producen un valor
//   3. Expression → Código que SÍ produce un valor
//
// =============================================================================

// Node es la interfaz raíz. TODO en el AST es un nodo.
// Cada nodo debe poder:
//   - TokenLiteral(): Decirnos qué token lo originó (para depuración)
//   - String():       Imprimirse como texto legible (para ver el árbol)
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement (Sentencia) = Instrucciones que HACEN algo pero NO devuelven un valor.
//
// Ejemplos de sentencias:
//   - `let x = 5;`    → Crea una variable (hace algo), pero no devuelve un valor
//   - `return 10;`    → Retorna un valor, pero la instrucción en sí no produce uno
//
// ¿Por qué tiene un método vacío `statementNode()`?
// Es un truco de Go para que el compilador nos obligue a distinguir entre
// Statements y Expressions. Si accidentalmente usamos un Statement donde
// se espera una Expression, Go nos dará un error de compilación.
type Statement interface {
	Node
	statementNode() // Método "marcador" vacío. Solo existe para diferenciar de Expression.
}

// Expression (Expresión) = Código que SÍ produce/devuelve un valor.
//
// Ejemplos de expresiones:
//   - `5`             → Produce el valor 5
//   - `5 + 3`         → Produce el valor 8
//   - `add(2, 3)`     → Produce el valor que retorne la función add
//   - `x`             → Produce el valor que tenga la variable x en memoria
//
// La distinción es importante: en `let edad = 5 + 3;`, la parte `5 + 3` es
// una Expression (produce 8), y todo junto `let edad = 8;` es un Statement.
type Expression interface {
	Node
	expressionNode() // Método "marcador" vacío. Solo existe para diferenciar de Statement.
}

// =============================================================================
// EL NODO RAÍZ: Program
// =============================================================================

// Program es el nodo más importante: la RAÍZ del árbol.
//
// Todo programa (todo lo que escribimos en CatCompiler) es simplemente
// una lista de sentencias ejecutadas de arriba hacia abajo.
//
// Ejemplo: Si escribimos:
//   let x = 5;
//   let y = 10;
//   return x + y;
//
// El AST será:
//   Program {
//       Statements: [
//           LetStatement("x", 5),
//           LetStatement("y", 10),
//           ReturnStatement(InfixExpression(x + y))
//       ]
//   }
type Program struct {
	Statements []Statement // Lista ordenada de todas las sentencias del programa
}

// TokenLiteral devuelve el literal del primer token del programa (para depuración)
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String convierte TODO el programa de vuelta a texto, concatenando cada sentencia.
// Esto nos permite "imprimir" el AST y verificar que se construyó correctamente.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// =============================================================================
// SENTENCIAS (STATEMENTS)
// =============================================================================
// Las sentencias son las "instrucciones" de nuestro lenguaje.
// No producen un valor por sí mismas, sino que realizan una acción.
// =============================================================================

// LetStatement representa la declaración de una variable.
//
// Sintaxis: let <nombre> = <expresión>;
// Ejemplo:  let edad = 25;
// Ejemplo:  let resultado = 5 + 3 * 2;
//
// Componentes:
//   - Token: El token "let" que inició esta sentencia
//   - Name:  El identificador (nombre de la variable, ej. "edad")
//   - Value: La expresión cuyo resultado se asignará (ej. "25" o "5 + 3 * 2")
type LetStatement struct {
	Token token.Token  // El token LET
	Name  *Identifier  // El nombre de la variable
	Value Expression   // La expresión que produce el valor a asignar
}

func (ls *LetStatement) statementNode()       {} // Marca: esto es un Statement
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// String reconstruye la sentencia let como texto: "let x = 5;"
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ") // "let "
	out.WriteString(ls.Name.String())         // "x"
	out.WriteString(" = ")                    // " = "
	if ls.Value != nil {
		out.WriteString(ls.Value.String()) // "5" o "(5 + 3)"
	}
	out.WriteString(";") // ";"
	return out.String()
}

// ReturnStatement representa la instrucción de retorno de una función.
//
// Sintaxis: return <expresión>;
// Ejemplo:  return 42;
// Ejemplo:  return x + y;
//
// Componentes:
//   - Token:       El token "return"
//   - ReturnValue: La expresión cuyo resultado se devuelve
type ReturnStatement struct {
	Token       token.Token // El token RETURN
	ReturnValue Expression  // La expresión que se retorna
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// String reconstruye la sentencia return como texto: "return 42;"
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement es un "envoltorio" para las expresiones sueltas.
//
// ¿Por qué existe? Porque en nuestro lenguaje podemos escribir una expresión
// como una línea independiente:
//   5 + 3;
//   add(2, 3);
//
// Estas líneas son válidas, pero el nodo raíz Program solo acepta Statements.
// Entonces envolvemos la Expression dentro de un ExpressionStatement para
// poder meterla en la lista de sentencias del programa.
type ExpressionStatement struct {
	Token      token.Token // El primer token de la expresión
	Expression Expression  // La expresión en sí
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement representa un bloque de código entre llaves { ... }
//
// Se usa dentro de if, else, while, for y funciones.
// Ejemplo: { let x = 5; return x; }
//
// Un bloque es simplemente una mini-lista de sentencias, igual que Program,
// pero delimitada por llaves.
type BlockStatement struct {
	Token      token.Token // El token {
	Statements []Statement // Las sentencias dentro del bloque
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// =============================================================================
// EXPRESIONES SIMPLES (LITERALS)
// =============================================================================
// Los "literales" son los valores más básicos del lenguaje: un número, un
// nombre de variable, un true/false. Son las HOJAS del árbol (no tienen hijos).
// =============================================================================

// Identifier representa el nombre de una variable o función.
//
// Ejemplo: en `let x = 5;`, la "x" es un Identifier.
//          en `return miVariable;`, "miVariable" es un Identifier.
//
// ¿Por qué es una Expression y no un Statement?
// Porque un nombre de variable PRODUCE un valor cuando lo evaluamos.
// Si escribimos `let y = x;`, la "x" se evalúa a lo que valga (ej. 5).
type Identifier struct {
	Token token.Token // El token IDENT
	Value string      // El nombre real de la variable (ej. "x", "edad")
}

func (i *Identifier) expressionNode()      {} // Marca: esto es una Expression
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral representa un número entero escrito directamente en el código.
//
// Ejemplo: 42, 0, 100
//
// Value es int64 (no string) porque durante el parseo convertimos el texto "42"
// al número real 42 para poder hacer operaciones matemáticas después.
type IntegerLiteral struct {
	Token token.Token // El token INT
	Value int64       // El valor numérico real (ej. 42)
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// Boolean representa un valor verdadero o falso: true, false
type Boolean struct {
	Token token.Token // El token TRUE o FALSE
	Value bool        // true o false como valor real de Go
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// =============================================================================
// EXPRESIONES CON OPERADORES
// =============================================================================
// Estas expresiones tienen operadores que combinan valores.
// =============================================================================

// PrefixExpression representa un operador que va ANTES de su operando.
//
// Ejemplos:
//   -5     → Operador: "-", Right: IntegerLiteral(5)
//   !true  → Operador: "!", Right: Boolean(true)
//
// En el árbol se ve así:
//     PrefixExpression
//        /        \
//   Operator:"-"  Right: 5
//
// El método String() lo imprime como "(-5)" con paréntesis para claridad.
type PrefixExpression struct {
	Token    token.Token // El token del prefijo, ej. '!' o '-'
	Operator string      // El operador como string (ej. "-", "!")
	Right    Expression  // La expresión a la que se le aplica el operador
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression representa un operador que va ENTRE dos operandos.
//
// Ejemplos:
//   5 + 3   → Left: 5, Operator: "+", Right: 3
//   x > y   → Left: x, Operator: ">", Right: y
//   a == b  → Left: a, Operator: "==", Right: b
//
// En el árbol se ve así:
//       InfixExpression
//       /      |      \
//   Left:5  Op:"+"  Right:3
//
// Cuando hay precedencia, se anidan:
//   1 + 2 * 3  →  InfixExpression(1, "+", InfixExpression(2, "*", 3))
//
// String() lo imprime con paréntesis: "(1 + (2 * 3))" para que quede
// absolutamente claro el orden de agrupación.
type InfixExpression struct {
	Token    token.Token // El token del operador, ej. '+'
	Left     Expression  // El operando izquierdo
	Operator string      // El operador como string (ej. "+", "*", "==")
	Right    Expression  // El operando derecho
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// =============================================================================
// CONTROL DE FLUJO: IF / ELSE
// =============================================================================

// IfExpression representa una estructura condicional if/else.
//
// Sintaxis: if (<condición>) { <consecuencia> } else { <alternativa> }
// Ejemplo:  if (x > 5) { return true; } else { return false; }
//
// Componentes:
//   - Condition:   La expresión que se evalúa (ej. "x > 5")
//   - Consequence: El bloque de código que se ejecuta SI la condición es verdadera
//   - Alternative: El bloque de código que se ejecuta SI la condición es falsa (opcional)
//
// Nota: Lo tratamos como Expression porque en muchos lenguajes un if puede
// devolver un valor (ej. `let x = if (true) { 5 } else { 10 };`).
type IfExpression struct {
	Token       token.Token     // El token 'if'
	Condition   Expression      // La condición a evaluar
	Consequence *BlockStatement // Bloque del "if" (se ejecuta si la condición es true)
	Alternative *BlockStatement // Bloque del "else" (se ejecuta si la condición es false, puede ser nil)
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// =============================================================================
// FUNCIONES
// =============================================================================

// FunctionLiteral representa la DEFINICIÓN de una función.
//
// Sintaxis: function(<param1>, <param2>) { <cuerpo> }
// Ejemplo:  function(x, y) { return x + y; }
//
// Componentes:
//   - Parameters: Lista de identificadores que recibe la función (ej. [x, y])
//   - Body:       El bloque de código que se ejecuta cuando la función es llamada
//
// En el árbol:
//   FunctionLiteral
//      /          \
//   Params:[x,y]   Body: BlockStatement{ ReturnStatement(x + y) }
type FunctionLiteral struct {
	Token      token.Token    // El token 'function'
	Parameters []*Identifier  // Los nombres de los parámetros
	Body       *BlockStatement // El cuerpo de la función
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression representa la LLAMADA (ejecución) de una función.
//
// Sintaxis: <función>(<arg1>, <arg2>)
// Ejemplo:  add(2, 3)
// Ejemplo:  function(x) { x + 1; }(5)  ← función anónima llamada inmediatamente
//
// Componentes:
//   - Function:  La expresión que produce la función (un Identifier como "add"
//                o un FunctionLiteral completo)
//   - Arguments: Las expresiones que se pasan como argumentos
type CallExpression struct {
	Token     token.Token  // El token '('
	Function  Expression   // El Identifier o FunctionLiteral a ejecutar
	Arguments []Expression // Los argumentos pasados a la función
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// =============================================================================
// CICLOS: WHILE y FOR
// =============================================================================

// WhileExpression representa un bucle while.
//
// Sintaxis: while (<condición>) { <cuerpo> }
// Ejemplo:  while (x < 10) { let x = x + 1; }
//
// El evaluador ejecutará el Body repetidamente mientras Condition sea verdadera.
type WhileExpression struct {
	Token     token.Token     // El token 'while'
	Condition Expression      // La condición que se revisa en cada iteración
	Body      *BlockStatement // El bloque que se repite
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	var out bytes.Buffer
	out.WriteString("while")
	out.WriteString(we.Condition.String())
	out.WriteString(" ")
	out.WriteString(we.Body.String())
	return out.String()
}

// ForExpression representa un bucle for (simplificado).
//
// Sintaxis: for (<condición>) { <cuerpo> }
// Ejemplo:  for (i < 100) { let i = i + 1; }
//
// Por ahora funciona igual que while (con una sola condición).
// Se puede expandir después para soportar for(init; cond; post).
type ForExpression struct {
	Token     token.Token     // El token 'for'
	Condition Expression      // La condición del bucle
	Body      *BlockStatement // El bloque que se repite
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	var out bytes.Buffer
	out.WriteString("for")
	out.WriteString(fe.Condition.String())
	out.WriteString(" ")
	out.WriteString(fe.Body.String())
	return out.String()
}
