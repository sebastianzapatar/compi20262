package repl

import (
	"bufio"
	"compigo/lexer"
	"compigo/token"
	"fmt"
	"io"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		
		cleanedLine := strings.TrimSpace(strings.ToLower(line))
		if cleanedLine == "exit" || cleanedLine == "salir" {
			fmt.Fprintf(out, "¡Nos vemos! Saliendo de CompiGo...\n")
			return
		}

		l := lexer.New(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
