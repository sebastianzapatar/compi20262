// =============================================================================
// PAQUETE object (Sistema de Objetos)
// =============================================================================
//
// Este paquete define los VALORES REALES que existen durante la ejecución.
// Cuando el evaluador procesa `5 + 3`, el resultado es un objeto Integer
// con Value = 8. Cada tipo de dato de CatCompiler tiene su propio Object.
//
// =============================================================================
package object

import (
	"compigo/ast"
	"fmt"
	"strings"
)

// ObjectType identifica qué tipo de objeto es (entero, booleano, etc.)
type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
)

// Object es la interfaz que todo valor en CatCompiler debe cumplir.
// Type() devuelve qué tipo es, Inspect() devuelve su representación como texto.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// =============================================================================
// TIPOS DE OBJETOS
// =============================================================================

// Integer representa un número entero (ej. 42)
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// Boolean representa true o false
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// Null representa la ausencia de valor
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue envuelve un valor que está siendo retornado por un return.
// Necesitamos envolverlo para que el evaluador sepa que debe "parar" y
// propagar el valor hacia arriba sin seguir ejecutando más sentencias.
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error representa un error en tiempo de ejecución (ej. "división por cero")
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Function representa una función definida por el usuario.
// Guarda los parámetros, el cuerpo (bloque de código) y el entorno
// en el que fue creada (esto es clave para los closures).
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("function(%s) { ... }", strings.Join(params, ", "))
}

// =============================================================================
// EL ENTORNO (ENVIRONMENT) - LA MEMORIA DEL PROGRAMA
// =============================================================================
//
// El Environment es un diccionario que asocia nombres de variables con sus
// valores. Tiene un puntero "outer" a un entorno padre para soportar scopes
// (ámbitos). Cuando una función se ejecuta, crea un nuevo entorno hijo;
// si no encuentra una variable en su scope local, busca en el padre.
//
// =============================================================================

type Environment struct {
	store map[string]Object
	outer *Environment // Entorno padre (nil si es el global)
}

// NewEnvironment crea un entorno nuevo (el global)
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment crea un entorno hijo que hereda del padre.
// Se usa cuando se ejecuta una función: el cuerpo de la función tiene
// su propio scope, pero puede acceder a las variables del scope padre.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get busca una variable por nombre. Si no está en el scope local,
// busca en el padre (y así recursivamente hasta el global).
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set guarda una variable en el scope actual.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
