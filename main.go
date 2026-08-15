package main

import (
	"compigo/repl"
	"fmt"
	"os"
	"os/user"
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
	fmt.Printf("Escribe algún comando y deja que la magia ocurra...\n")
	repl.Start(os.Stdin, os.Stdout)
}
