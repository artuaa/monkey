package repl

import (
	"bufio"
	"fmt"
	"interpreter/lexer"
	"interpreter/parser"
	"io"
)

const (
	PROMPT = ">> "
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()

		if !scanned {
			return
		}

		line := scanner.Text()

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String()+"\n")

	}

}

func printParseErrors(out io.Writer, s []string) {
	for _, msg := range s {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
