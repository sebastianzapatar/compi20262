package main

import (
	"bufio"
	"compigo/repl"
	"fmt"
	"os"
	"os/user"
	"strings"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("********************************************\n")
	fmt.Printf("**                                        **\n")
	fmt.Printf("**  -- BIENVENIDO A CATCOMPILER REPL --   **\n")
	fmt.Printf("**                                        **\n")
	fmt.Printf("**      /\\_/\\     ¡Miau! Soy Limón,       **\n")
	fmt.Printf("**     ( o.o )    tu asistente felino.    **\n")
	fmt.Printf("**      > ^ <     ¡Dame esos tokens!      **\n")
	fmt.Printf("**                                        **\n")
	fmt.Printf("********************************************\n\n")
	fmt.Printf("¡Hola %s! Estás en CatCompiler.\n",
		user.Username)
	
	fmt.Printf("\nSelecciona el modo de interacción:\n")
	fmt.Printf("1. Modo Lexer (Muestra los Tokens generados)\n")
	fmt.Printf("2. Modo Parser (Construye y muestra el Árbol AST)\n")
	fmt.Printf("Opción [1 o 2]: ")
	
	reader := bufio.NewReader(os.Stdin)
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	mode := repl.LEXER_MODE
	if option == "2" {
		mode = repl.PARSER_MODE
		fmt.Printf("\n¡Iniciando en Modo Parser! Escribe código para ver el árbol.\n")
	} else {
		fmt.Printf("\n¡Iniciando en Modo Lexer! Escribe código para ver los tokens.\n")
	}

	repl.Start(os.Stdin, os.Stdout, mode)
}
