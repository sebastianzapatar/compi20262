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
	if len(os.Args) > 1 {
		// Modo ejecución de archivo
		filename := os.Args[1]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("¡Miau! No pude abrir el archivo %s: %s\n", filename, err)
			os.Exit(1)
		}
		defer file.Close()
		
		fmt.Printf("🐾 Ejecutando %s...\n", filename)
		repl.StartFile(file, os.Stdout)
		return
	}

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
	fmt.Printf("3. Modo Evaluador (¡Ejecuta el código en vivo!)\n")
	fmt.Printf("Opción [1, 2 o 3]: ")
	
	reader := bufio.NewReader(os.Stdin)
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	mode := repl.LEXER_MODE
	if option == "3" {
		mode = repl.EVALUATOR_MODE
		fmt.Printf("\n¡Iniciando en Modo Evaluador! Escribe código y mira el resultado.\n")
	} else if option == "2" {
		mode = repl.PARSER_MODE
		fmt.Printf("\n¡Iniciando en Modo Parser! Escribe código para ver el árbol.\n")
	} else {
		fmt.Printf("\n¡Iniciando en Modo Lexer! Escribe código para ver los tokens.\n")
	}

	repl.Start(os.Stdin, os.Stdout, mode)
}
