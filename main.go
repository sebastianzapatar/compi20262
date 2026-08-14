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
	fmt.Printf("**    -- BIENVENIDO A COMPIGO REPL --     **\n")
	fmt.Printf("**                                        **\n")
	fmt.Printf("**      (  •_•)                           **\n")
	fmt.Printf("**      / >[TOKENS]                       **\n")
	fmt.Printf("**      ¡Hora de compilar sin llorar!     **\n")
	fmt.Printf("**                                        **\n")
	fmt.Printf("********************************************\n\n")
	fmt.Printf("¡Hola %s! Estás en el lenguaje de programación CompiGo.\n",
		user.Username)
	fmt.Printf("Escribe algún comando y deja que la magia ocurra...\n")
	repl.Start(os.Stdin, os.Stdout)
}
