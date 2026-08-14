# CatCompiler 🐾

**CatCompiler** es un intérprete escrito en Go donde **Limón** 🐈, nuestro gato protagonista, te acompaña en la aventura de compilar código. Actualmente se encuentra en su primera fase de desarrollo, la cual incluye el análisis léxico (Lexer) y un entorno interactivo de lectura-evaluación-impresión (REPL).

## Características Actuales

El lenguaje soporta una sintaxis básica estilo C/JavaScript. En esta primera fase, el Lexer es capaz de reconocer los siguientes elementos:

* **Palabras reservadas:** `let`, `function`, `if`, `else`, `elif`, `return`, `for`, `while`, `break`, `true`, `false`
* **Tipos de datos literales:** Enteros (`10`, `42`, etc.)
* **Identificadores:** Nombres de variables y funciones (ej. `x`, `suma`, `miVariable`)
* **Operadores:**
  * Asignación: `=`
  * Aritméticos: `+`, `-`, `*`, `/`
  * Comparación: `==`, `!=`, `<`, `>`
  * Lógicos: `!` (Not)
* **Delimitadores:** `,`, `;`, `(`, `)`, `{`, `}`, `[`, `]`

## Estructura del Proyecto

* `/token`: Define la estructura y los tipos de tokens que el Lexer identificará.
* `/lexer`: Contiene la lógica principal del analizador léxico, que toma el código fuente como texto y lo transforma en una secuencia de tokens. Incluye una suite completa de pruebas unitarias.
* `/repl`: Implementa la consola interactiva (Read-Eval-Print Loop) para probar el Lexer.

## Cómo probarlo

Para probar el intérprete, asegúrate de tener Go instalado y ejecuta el siguiente comando en la raíz del proyecto para iniciar el REPL:

```bash
go run main.go
```

Una vez dentro, puedes escribir instrucciones para ver cómo el Lexer las divide en tokens:

```
>> let a = 10;
{Type:LET Literal:let}
{Type:IDENT Literal:a}
{Type:ASSIGN Literal:=}
{Type:INT Literal:10}
{Type:SEMICOLON Literal:;}

>> if (a < 20) { return true; }
{Type:IF Literal:if}
{Type:LPAREN Literal:(}
{Type:IDENT Literal:a}
{Type:LT Literal:<}
{Type:INT Literal:20}
{Type:RPAREN Literal:)}
{Type:LBRACE Literal:{}
{Type:RETURN Literal:return}
{Type:TRUE Literal:true}
{Type:SEMICOLON Literal:;}
{Type:RBRACE Literal:}}
```

## Pruebas Unitarias

Para ejecutar la batería de pruebas del lexer, usa:

```bash
go test ./...
```
