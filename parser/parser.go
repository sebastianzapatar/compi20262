// =============================================================================
// PAQUETE parser (Analizador Sintáctico)
// =============================================================================
//
// ¿QUÉ ES ESTE ARCHIVO?
// ----------------------
// El Parser es el "cerebro gramatical" de nuestro intérprete. Toma la lista
// de tokens que generó el Lexer y construye el Árbol de Sintaxis Abstracta (AST).
//
// ¿QUÉ ALGORITMO USA?
// --------------------
// Usamos el algoritmo de "Pratt Parsing" (Top-Down Operator Precedence).
// Fue inventado por Vaughan Pratt en 1973.
//
// La idea es simple pero poderosa:
//   1. Cada token tiene una "prioridad" (precedencia). Por ejemplo, * tiene
//      más prioridad que +.
//   2. Cuando encontramos un token, buscamos en un DICCIONARIO qué función
//      sabe parsearlo.
//   3. La función de parseo puede llamarse a sí misma recursivamente,
//      y la precedencia decide cuándo "parar" de agrupar.
//
// EJEMPLO VISUAL:
// ---------------
// Input: "1 + 2 * 3"
// Tokens: [INT(1), PLUS, INT(2), ASTERISK, INT(3)]
//
// El parser hace esto paso a paso:
//   1. Ve INT(1) → crea IntegerLiteral(1) como leftExp
//   2. Ve PLUS (+) con prioridad SUM(4) → crea InfixExpression con Left=1
//   3. Ahora necesita el Right. Llama recursivamente a parseExpression(SUM)
//   4. Ve INT(2) → crea IntegerLiteral(2) como leftExp
//   5. Ve ASTERISK (*) con prioridad PRODUCT(5). Como 5 > 4 (SUM), sigue agrupando
//   6. Crea InfixExpression con Left=2, Op=*, Right=3
//   7. Ese InfixExpression(2*3) se convierte en el Right del paso 2
//
// Resultado: InfixExpression(1, "+", InfixExpression(2, "*", 3))
// Impreso:   (1 + (2 * 3))
//
// =============================================================================
package parser

import (
	"compigo/ast"
	"compigo/lexer"
	"compigo/token"
	"fmt"
	"strconv"
)

// =============================================================================
// TABLA DE PRECEDENCIAS (JERARQUÍA MATEMÁTICA)
// =============================================================================
//
// La precedencia define el ORDEN en que se agrupan las operaciones.
// Un número más alto = se agrupa primero.
//
// Ejemplo práctico:
//   En la expresión "1 + 2 * 3 == 7"
//   - Primero se agrupa 2 * 3     (PRODUCT = 5)
//   - Luego se agrupa 1 + 6       (SUM = 4)
//   - Finalmente se agrupa 7 == 7 (EQUALS = 2)
//   - Resultado: ((1 + (2 * 3)) == 7)
//
// La constante LOWEST (1) es el "piso". Cuando llamamos a parseExpression(LOWEST)
// estamos diciendo "agrupa TODO lo que puedas".
//
// =============================================================================

const (
	_ int = iota
	LOWEST      // 1 - El piso. Se usa para empezar a parsear sin restricciones.
	EQUALS      // 2 - == y !=
	LESSGREATER // 3 - > y <
	SUM         // 4 - + y -
	PRODUCT     // 5 - * y /
	PREFIX      // 6 - -X o !X (operadores unarios)
	EXPONENT    // 7 - ** (potencia, máxima prioridad matemática)
	CALL        // 8 - miFuncion(X) (llamadas a funciones, la más alta)
)

// precedences es el diccionario que asocia cada token-operador con su nivel
// de precedencia. Si un token no está aquí, se asume LOWEST.
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,      // ==  → prioridad 2
	token.NOT_EQ:   EQUALS,      // !=  → prioridad 2
	token.LT:       LESSGREATER, // <   → prioridad 3
	token.GT:       LESSGREATER, // >   → prioridad 3
	token.LTE:      LESSGREATER, // <=  → prioridad 3
	token.GTE:      LESSGREATER, // >=  → prioridad 3
	token.PLUS:     SUM,         // +   → prioridad 4
	token.MINUS:    SUM,         // -   → prioridad 4
	token.SLASH:    PRODUCT,     // /   → prioridad 5
	token.ASTERISK: PRODUCT,     // *   → prioridad 5
	token.POW:      EXPONENT,    // **  → prioridad 7
	token.LPAREN:   CALL,        // (   → prioridad 8 (para llamadas a función)
}

// =============================================================================
// TIPOS DE FUNCIONES DE PARSEO (PRATT PARSING)
// =============================================================================
//
// El truco de Pratt Parsing es usar DOS tipos de funciones:
//
// 1. prefixParseFn: Se invoca cuando el token aparece AL INICIO de una expresión.
//    Ejemplo: El "-" en "-5" o el "!" en "!true" o un número "42".
//    No recibe parámetros porque no hay nada a su izquierda.
//
// 2. infixParseFn: Se invoca cuando el token aparece EN MEDIO de una expresión.
//    Ejemplo: El "+" en "5 + 3" o el "*" en "2 * 4".
//    Recibe como parámetro la expresión de la IZQUIERDA (lo que ya se parseó).
//
// =============================================================================

type (
	// prefixParseFn: función que sabe parsear un token cuando aparece al inicio.
	// Ejemplo: parseIntegerLiteral() para el token INT
	prefixParseFn func() ast.Expression

	// infixParseFn: función que sabe parsear un token cuando aparece en medio.
	// Recibe la expresión izquierda como argumento.
	// Ejemplo: parseInfixExpression(left) para el token PLUS
	infixParseFn func(ast.Expression) ast.Expression
)

// =============================================================================
// LA ESTRUCTURA DEL PARSER
// =============================================================================

// Parser contiene todo el estado necesario para analizar los tokens.
//
// Campos:
//   - l:              Puntero al Lexer del que leemos los tokens
//   - errors:         Lista de errores sintácticos encontrados
//   - curToken:       El token que estamos evaluando AHORA
//   - peekToken:      El SIGUIENTE token (nos permite "espiar" hacia adelante)
//   - prefixParseFns: Diccionario {tipo_de_token → función_prefix}
//   - infixParseFns:  Diccionario {tipo_de_token → función_infix}
type Parser struct {
	l      *lexer.Lexer // El lexer que nos da los tokens
	errors []string     // Errores acumulados durante el parseo

	curToken  token.Token // El token actual que estamos analizando
	peekToken token.Token // El siguiente token (para mirar hacia adelante)

	// Diccionarios de Pratt Parsing: asocian cada tipo de token con
	// la función que sabe parsearlo
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// =============================================================================
// CONSTRUCTOR DEL PARSER
// =============================================================================

// New crea e inicializa un nuevo Parser.
//
// Aquí es donde REGISTRAMOS qué función de parseo corresponde a cada token.
// Esto es el corazón del Pratt Parsing: un diccionario que dice
// "cuando veas un INT, usa parseIntegerLiteral" o
// "cuando veas un PLUS en medio, usa parseInfixExpression".
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// =========================================================================
	// REGISTRO DE FUNCIONES PREFIX
	// =========================================================================
	// Estas funciones se usan cuando el token aparece AL INICIO de una expresión.
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)            // Variables: x, edad, nombre
	p.registerPrefix(token.INT, p.parseIntegerLiteral)          // Números: 5, 42, 100
	p.registerPrefix(token.BANG, p.parsePrefixExpression)       // Negación lógica: !true
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)      // Negación numérica: -5
	p.registerPrefix(token.TRUE, p.parseBoolean)                // Literal: true
	p.registerPrefix(token.FALSE, p.parseBoolean)               // Literal: false
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)    // Paréntesis: (5 + 3)
	p.registerPrefix(token.IF, p.parseIfExpression)             // Condicional: if (...)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)    // Función: function(x) { ... }
	p.registerPrefix(token.WHILE, p.parseWhileExpression)       // Ciclo: while (...) { ... }
	p.registerPrefix(token.FOR, p.parseForExpression)           // Ciclo: for (...) { ... }

	// =========================================================================
	// REGISTRO DE FUNCIONES INFIX
	// =========================================================================
	// Estas funciones se usan cuando el token aparece EN MEDIO de dos expresiones.
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)     // Suma: 5 + 3
	p.registerInfix(token.MINUS, p.parseInfixExpression)    // Resta: 5 - 3
	p.registerInfix(token.SLASH, p.parseInfixExpression)    // División: 10 / 2
	p.registerInfix(token.ASTERISK, p.parseInfixExpression) // Multiplicación: 5 * 3
	p.registerInfix(token.EQ, p.parseInfixExpression)       // Igualdad: x == y
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)   // Desigualdad: x != y
	p.registerInfix(token.LT, p.parseInfixExpression)       // Menor que: x < y
	p.registerInfix(token.GT, p.parseInfixExpression)       // Mayor que: x > y
	p.registerInfix(token.LTE, p.parseInfixExpression)      // Menor o igual: x <= y
	p.registerInfix(token.GTE, p.parseInfixExpression)      // Mayor o igual: x >= y
	p.registerInfix(token.POW, p.parseInfixExpression)      // Potencia: x ** y
	p.registerInfix(token.LPAREN, p.parseCallExpression)    // Llamada: add(x, y)

	// Leemos dos tokens para inicializar tanto curToken como peekToken.
	// Necesitamos tener siempre DOS tokens cargados para poder "espiar" el siguiente.
	p.nextToken()
	p.nextToken()

	return p
}

// Errors devuelve todos los errores sintácticos encontrados durante el parseo.
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken avanza al siguiente token. curToken toma el valor de peekToken,
// y peekToken lee un nuevo token del Lexer.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// =============================================================================
// PUNTO DE ENTRADA: ParseProgram
// =============================================================================

// ParseProgram es la función principal. Lee TODOS los tokens hasta EOF
// y construye el nodo raíz Program con sus sentencias.
//
// Algoritmo:
//   1. Crear un Program vacío
//   2. Mientras no hayamos llegado al final (EOF):
//      a. Intentar parsear la línea actual como una sentencia
//      b. Si fue exitoso, agregarla a la lista de sentencias
//      c. Avanzar al siguiente token
//   3. Devolver el Program completo
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement decide qué tipo de sentencia estamos leyendo.
//
// Mira el token actual y decide:
//   - Si es LET     → parseLetStatement()
//   - Si es RETURN  → parseReturnStatement()
//   - Cualquier otra cosa → parseExpressionStatement()
//
// Esto último es clave: si no es un let ni un return, asumimos que es
// una expresión suelta como `5 + 3;` o `add(1, 2);`
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// =============================================================================
// PARSEO DE SENTENCIAS (LET, RETURN, EXPRESSION)
// =============================================================================

// parseLetStatement construye un nodo LetStatement.
//
// Espera la estructura: LET <IDENT> = <EXPRESIÓN> ;
//
// Paso a paso para `let x = 5 + 3;`:
//   1. curToken es LET → guardamos el token
//   2. expectPeek(IDENT) → avanzamos y verificamos que lo siguiente sea un nombre ("x")
//   3. expectPeek(ASSIGN) → avanzamos y verificamos que venga un "="
//   4. nextToken() → avanzamos al inicio de la expresión
//   5. parseExpression(LOWEST) → parseamos "5 + 3" como una InfixExpression
//   6. Si hay ";", lo consumimos
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken} // Guardamos el token LET

	// Paso 2: Lo siguiente DEBE ser un identificador (el nombre de la variable)
	if !p.expectPeek(token.IDENT) {
		return nil // Error: "let 5 = ..." no tiene sentido
	}

	// Creamos el nodo Identifier con el nombre de la variable
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Paso 3: Después del nombre DEBE venir un signo "="
	if !p.expectPeek(token.ASSIGN) {
		return nil // Error: "let x 5" falta el "="
	}

	// Paso 4: Avanzamos al primer token de la expresión (el valor)
	p.nextToken()

	// Paso 5: Parseamos toda la expresión del lado derecho
	stmt.Value = p.parseExpression(LOWEST)

	// Paso 6: Si hay punto y coma, lo consumimos (es opcional en el REPL)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement construye un nodo ReturnStatement.
//
// Espera la estructura: RETURN <EXPRESIÓN> ;
//
// Ejemplo para `return 42;`:
//   1. curToken es RETURN → guardamos el token
//   2. Avanzamos al inicio de la expresión
//   3. Parseamos "42" como IntegerLiteral
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken() // Avanzamos al inicio de la expresión

	stmt.ReturnValue = p.parseExpression(LOWEST) // Parseamos la expresión

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement envuelve una expresión suelta en un Statement.
//
// Ejemplo: Si escribimos `5 + 3;` en el REPL, no es un let ni un return.
// Es simplemente una expresión. La envolvemos en ExpressionStatement para
// poder agregarla a la lista de sentencias del programa.
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// Parseamos la expresión empezando con la prioridad más baja (LOWEST)
	// para que agrupe TODO lo que pueda
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseBlockStatement parsea todo lo que está dentro de { ... }
//
// Se usa dentro de if, else, while, for y funciones.
// Ejemplo: { let x = 5; return x; }
//
// Lee sentencias una por una hasta encontrar la llave de cierre "}" o el EOF.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // Saltamos la llave de apertura "{"

	// Leemos sentencias hasta encontrar "}" o llegar al final del archivo
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// =============================================================================
// ★★★ PRATT PARSER: EL NÚCLEO MÁGICO ★★★
// =============================================================================
//
// Esta es LA función más importante de todo el parser.
// Es recursiva y es la que decide cómo AGRUPAR las operaciones.
//
// REGLA DE ORO:
//   "Sigue agrupando tokens hacia la derecha mientras el siguiente token
//    tenga MAYOR precedencia que tu precedencia actual."
//
// EJEMPLO DETALLADO para "1 + 2 * 3":
//
//   Llamada 1: parseExpression(LOWEST=1)
//     → curToken=1, llama a parseIntegerLiteral() → leftExp = IntegerLiteral(1)
//     → peekToken=+, peekPrecedence()=SUM(4). Como 1 < 4, ENTRA al loop.
//     → Avanza. Llama a parseInfixExpression(left=1)
//       → Guarda Left=1, Op="+", precedence=SUM(4)
//       → Llama recursivamente: parseExpression(SUM=4) para obtener Right
//
//   Llamada 2: parseExpression(SUM=4)
//     → curToken=2, llama a parseIntegerLiteral() → leftExp = IntegerLiteral(2)
//     → peekToken=*, peekPrecedence()=PRODUCT(5). Como 4 < 5, ENTRA al loop.
//     → Avanza. Llama a parseInfixExpression(left=2)
//       → Guarda Left=2, Op="*", precedence=PRODUCT(5)
//       → Llama recursivamente: parseExpression(PRODUCT=5) para obtener Right
//
//   Llamada 3: parseExpression(PRODUCT=5)
//     → curToken=3, llama a parseIntegerLiteral() → leftExp = IntegerLiteral(3)
//     → peekToken=EOF, peekPrecedence()=LOWEST(1). Como 5 > 1, NO entra al loop.
//     → Retorna IntegerLiteral(3)
//
//   De vuelta en Llamada 2: Right = IntegerLiteral(3)
//     → Retorna InfixExpression(2, "*", 3)
//
//   De vuelta en Llamada 1: Right = InfixExpression(2, "*", 3)
//     → Retorna InfixExpression(1, "+", InfixExpression(2, "*", 3))
//
//   ¡Resultado final: (1 + (2 * 3))!
//
// =============================================================================

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// Paso 1: Buscar la función PREFIX para el token actual
	// (¿cómo se parsea este token cuando aparece al inicio?)
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// Si no existe función prefix para este token, es un error.
		// Ejemplo: si el token es ";" o "}", no tiene sentido como inicio de expresión.
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// Paso 2: Ejecutar la función prefix para obtener la expresión izquierda
	leftExp := prefix()

	// Paso 3: El loop mágico de Pratt
	// Mientras NO hayamos llegado a un ";" Y la precedencia del siguiente token
	// sea MAYOR que nuestra precedencia actual, seguimos agrupando.
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// Buscar la función INFIX para el siguiente token
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp // No hay función infix, devolvemos lo que tenemos
		}

		p.nextToken() // Avanzamos al operador

		// Ejecutamos la función infix, pasándole lo que teníamos a la izquierda.
		// Esta función llamará recursivamente a parseExpression para obtener
		// lo que viene a la derecha.
		leftExp = infix(leftExp)
	}

	return leftExp
}

// =============================================================================
// FUNCIONES PREFIX (se usan al INICIO de una expresión)
// =============================================================================

// parseIdentifier crea un nodo Identifier para nombres de variables.
// Es la función prefix más simple: solo devuelve el identificador actual.
// Ejemplo: "miVariable" → Identifier{Value: "miVariable"}
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral convierte un token INT en un nodo IntegerLiteral.
// Convierte el string "42" en el número entero 42 usando strconv.ParseInt.
// Si la conversión falla (ej. número demasiado grande), registra un error.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("No se pudo parsear %q como entero", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parsePrefixExpression maneja operadores unarios como -5 o !true.
//
// Paso a paso para "-5":
//   1. curToken es "-" → guardamos el operador
//   2. Avanzamos al siguiente token (el "5")
//   3. Parseamos "5" con precedencia PREFIX (alta) para que no agrupe más allá
//
// Resultado: PrefixExpression{Operator: "-", Right: IntegerLiteral(5)}
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal, // "-" o "!"
	}

	p.nextToken() // Avanzamos al operando (lo que viene después del operador)
	expression.Right = p.parseExpression(PREFIX) // Parseamos con prioridad PREFIX (alta)
	return expression
}

// parseBoolean crea un nodo Boolean.
// true/false son constantes del lenguaje que se convierten directamente.
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression maneja los paréntesis: (5 + 3)
//
// Los paréntesis no crean un nodo en el AST. Solo controlan la precedencia:
// al parsear lo que hay dentro con LOWEST, forzamos a que se agrupe primero.
//
// Paso a paso para "(5 + 3)":
//   1. curToken es "(" → lo saltamos
//   2. Parseamos "5 + 3" con precedencia LOWEST (se agrupa todo)
//   3. Verificamos que venga el ")" de cierre
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // Saltamos el "("

	exp := p.parseExpression(LOWEST) // Parseamos TODO lo de adentro

	if !p.expectPeek(token.RPAREN) { // Verificamos que haya un ")"
		return nil
	}
	return exp
}

// =============================================================================
// IF, WHILE, FOR Y FUNCIONES
// =============================================================================

// parseIfExpression construye un nodo IfExpression.
//
// Sintaxis esperada: if (<condición>) { <bloque> } else { <bloque> }
//
// Paso a paso para "if (x > 5) { return true; } else { return false; }":
//   1. curToken es IF
//   2. expectPeek("(") → verificamos el paréntesis de apertura
//   3. Parseamos la condición "x > 5" como una InfixExpression
//   4. expectPeek(")") → verificamos el cierre del paréntesis
//   5. expectPeek("{") → verificamos la llave de apertura
//   6. parseBlockStatement() → parseamos "return true;" como bloque
//   7. Si hay un "else", repetimos para el bloque alternativo
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil // Error: falta el "(" después de if
	}

	p.nextToken() // Saltamos el "(" para empezar a parsear la condición
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil // Error: falta el ")" después de la condición
	}

	if !p.expectPeek(token.LBRACE) {
		return nil // Error: falta la "{" del bloque
	}

	expression.Consequence = p.parseBlockStatement() // Parseamos el bloque if { ... }

	// Si después del bloque hay un "else", parseamos el bloque alternativo
	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // Avanzamos al "else"

		if !p.expectPeek(token.LBRACE) {
			return nil // Error: falta la "{" después de else
		}

		expression.Alternative = p.parseBlockStatement() // Parseamos el bloque else { ... }
	}

	return expression
}

// parseWhileExpression construye un nodo WhileExpression.
//
// Sintaxis esperada: while (<condición>) { <bloque> }
// Funciona exactamente igual que if, pero sin rama "else".
func (p *Parser) parseWhileExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockStatement()
	return expression
}

// parseForExpression construye un nodo ForExpression.
//
// Sintaxis esperada: for (<condición>) { <bloque> }
// Por ahora funciona igual que while, con una sola condición.
func (p *Parser) parseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockStatement()
	return expression
}

// parseFunctionLiteral construye un nodo FunctionLiteral.
//
// Sintaxis esperada: function(<params>) { <bloque> }
// Ejemplo: function(x, y) { return x + y; }
//
// Paso a paso:
//   1. curToken es FUNCTION
//   2. expectPeek("(") → verificamos el paréntesis
//   3. parseFunctionParameters() → leemos los nombres de los parámetros
//   4. expectPeek("{") → verificamos la llave
//   5. parseBlockStatement() → parseamos el cuerpo de la función
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters() // Lee la lista de parámetros

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement() // Lee el cuerpo de la función
	return lit
}

// parseFunctionParameters lee la lista de parámetros de una función.
//
// Ejemplo: para "function(x, y, z)", lee [x, y, z] como una lista de Identifiers.
//
// Algoritmo:
//   1. Si el siguiente token es ")", no hay parámetros → devuelve lista vacía
//   2. Lee el primer parámetro
//   3. Mientras haya comas, lee más parámetros
//   4. Verifica que termine con ")"
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Caso especial: function() sin parámetros
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken() // Avanzamos al primer parámetro

	// Leemos el primer parámetro
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// Mientras haya comas, leemos más parámetros
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Saltamos la coma
		p.nextToken() // Avanzamos al siguiente parámetro
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// Verificamos el cierre con ")"
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// =============================================================================
// FUNCIONES INFIX (se usan EN MEDIO de dos expresiones)
// =============================================================================

// parseInfixExpression construye un nodo InfixExpression.
// Esta función se llama cuando encontramos un operador como +, -, *, /, ==, etc.
//
// Recibe la expresión de la IZQUIERDA (lo que ya se parseó antes del operador).
//
// Paso a paso para "5 + 3" (cuando ya tenemos Left=5 y curToken="+"):
//   1. Guardamos el operador "+" y Left=5
//   2. Obtenemos la precedencia actual del "+" (SUM=4)
//   3. Avanzamos al siguiente token ("3")
//   4. Llamamos recursivamente a parseExpression(SUM) para obtener Right
//   5. Resultado: InfixExpression{Left: 5, Op: "+", Right: 3}
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left, // Lo que ya se parseó a la izquierda
	}

	precedence := p.curPrecedence() // Obtenemos la prioridad del operador actual
	p.nextToken()                   // Avanzamos al inicio de la expresión derecha
	expression.Right = p.parseExpression(precedence) // Parseamos el lado derecho

	return expression
}

// parseCallExpression construye un nodo CallExpression (llamada a función).
//
// Se activa cuando el parser encuentra un "(" después de una expresión.
// Ejemplo: `add(2, 3)` → El parser ya parseó "add" como Identifier,
// y cuando ve el "(", llama a esta función con function=Identifier("add").
//
// Recibe la expresión de la izquierda (el nombre o literal de la función).
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

// parseCallArguments lee la lista de argumentos de una llamada a función.
//
// Ejemplo: para "add(2, 3 + 1, x)", lee [2, (3+1), x] como una lista de Expressions.
//
// Funciona de forma similar a parseFunctionParameters, pero en vez de leer
// solo nombres (Identifiers), lee expresiones completas.
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// Caso especial: add() sin argumentos
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken() // Avanzamos al primer argumento
	args = append(args, p.parseExpression(LOWEST)) // Parseamos el primer argumento

	// Mientras haya comas, parseamos más argumentos
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Saltamos la coma
		p.nextToken() // Avanzamos al siguiente argumento
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil // Error: falta el ")" de cierre
	}

	return args
}

// =============================================================================
// FUNCIONES AUXILIARES (HELPERS)
// =============================================================================

// curTokenIs verifica si el token actual es del tipo esperado.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs verifica si el SIGUIENTE token es del tipo esperado.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek es el "guardián de la gramática".
//
// Verifica que el siguiente token sea el esperado:
//   - SI lo es: avanza y retorna true ✓
//   - NO lo es: registra un error y retorna false ✗
//
// Ejemplo: después de "let", esperamos un IDENT.
// Si viene un número, expectPeek registra el error:
// "Se esperaba IDENT, se obtuvo INT"
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// peekError registra un error cuando el token siguiente no es el esperado.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Error sintáctico: Se esperaba el token %s, se obtuvo %s",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// registerPrefix registra una función de parseo prefix para un tipo de token.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registra una función de parseo infix para un tipo de token.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// noPrefixParseFnError se llama cuando no hay función prefix registrada
// para un token. Esto significa que el token no puede iniciar una expresión.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("No se encontró función de parseo (prefix) para el token %s", t)
	p.errors = append(p.errors, msg)
}

// peekPrecedence devuelve la precedencia del SIGUIENTE token.
// Si el token no tiene precedencia registrada, devuelve LOWEST.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence devuelve la precedencia del token ACTUAL.
// Si el token no tiene precedencia registrada, devuelve LOWEST.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
