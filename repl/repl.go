package repl

import (
	"bufio"
	"compigo/evaluator"
	"compigo/lexer"
	"compigo/object"
	"compigo/parser"
	"compigo/token"
	"fmt"
	"io"
	"strings"
)

const PROMPT = ">> "

const (
	LEXER_MODE = iota
	PARSER_MODE
	EVALUATOR_MODE
)

func Start(in io.Reader, out io.Writer, mode int) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		
		cleanedLine := strings.TrimSpace(strings.ToLower(line))
		if cleanedLine == "exit" || cleanedLine == "salir" {
			fmt.Fprintf(out, "¡Miau! Nos vemos. Saliendo de CatCompiler...\n")
			return
		}

		l := lexer.New(line)

		if mode == LEXER_MODE {
			// Modo 1: Solo imprimir Tokens
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Fprintf(out, "%+v\n", tok)
			}
		} else if mode == PARSER_MODE {
			// Modo 2: Construir e imprimir el AST
			p := parser.New(l)
			program := p.ParseProgram()
			
			if len(p.Errors()) != 0 {
				printParserErrors(out, p.Errors())
				continue
			}

			// Imprimir la representación en string del AST
			fmt.Fprintf(out, "AST Generado: \n")
			fmt.Fprintf(out, "%s\n", program.String())
		} else {
			// Modo 3: Evaluador
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				printParserErrors(out, p.Errors())
				continue
			}

			evaluated := evaluator.Eval(program, env)
			if evaluated != nil {
				fmt.Fprintf(out, "%s\n", evaluated.Inspect())
			}
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintf(out, "¡Miau! Ocurrió un error al intentar parsear el código:\n")
	for _, msg := range errors {
		fmt.Fprintf(out, "\t- %s\n", msg)
	}
}

func StartFile(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	
	// Leer todo el archivo en memoria
	var code strings.Builder
	for scanner.Scan() {
		code.WriteString(scanner.Text())
		code.WriteString("\n")
	}

	l := lexer.New(code.String())
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return
	}

	evaluated := evaluator.Eval(program, env)
	if evaluated != nil {
		fmt.Fprintf(out, "%s\n", evaluated.Inspect())
	}
}
