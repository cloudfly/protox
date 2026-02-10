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

func genImport(w io.Writer, imports ...string) {
	if len(imports) == 0 {
		return
	}
	genWriteln(w, "import (")
	defer genWritelnln(w, ")")

	for _, it := range imports {
		if it == "" {
			genWriteln(w, "")
		} else {
			if strings.ContainsRune(it, '"') {
				genWritelnIndent(w, 1, "%s", it)
			} else {
				genWritelnIndent(w, 1, "\"%s\"", it)
			}
		}
	}
}
