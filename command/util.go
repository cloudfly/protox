package main

import (
	"fmt"
	"io"
	"strings"
)

func genWrite(w io.Writer, format string, args ...any) {
	w.Write([]byte(fmt.Sprintf(format, args...)))
}

func genWriteln(w io.Writer, format string, args ...any) {
	w.Write([]byte(fmt.Sprintf(format+"\n", args...)))
}
func genWritelnIndent(w io.Writer, indent int, format string, args ...any) {
	w.Write([]byte(strings.Repeat("\t", indent) + fmt.Sprintf(format+"\n", args...)))
}

func genWritelnln(w io.Writer, format string, args ...any) {
	w.Write([]byte(fmt.Sprintf(format+"\n\n", args...)))
}

func genWriteLines(w io.Writer, lines ...string) {
	for _, line := range lines {
		w.Write([]byte(line + "\n"))
	}
}

func genWriteLinesIndent(w io.Writer, indent int, lines ...string) {
	for _, line := range lines {
		w.Write([]byte(strings.Repeat("\t", indent) + line + "\n"))
	}
}

func genImport(w io.Writer, imports ...string) {
	genWriteln(w, "import (")
	defer genWritelnln(w, ")")

	for _, it := range imports {
		if it == "" {
			genWriteln(w, "")
		} else {
			genWritelnIndent(w, 1, "\"%s\"", it)
		}
	}
}
