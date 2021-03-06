package main

import "fmt"

type Converter struct {
	TypeIn, TypeOut string
	Function        string
}

func main() {
	converters := []Converter{
		{TypeIn: "io.Closer", TypeOut: "io.Closer"},
		{TypeIn: "io.Reader", TypeOut: "Reader", Function: "wrapReader"},
		{TypeIn: "io.ReaderAt", TypeOut: "ReaderAt", Function: "wrapReaderAt"},
		{TypeIn: "io.ReaderFrom", TypeOut: "io.ReaderFrom"},
		{TypeIn: "io.Seeker", TypeOut: "io.Seeker"},
		{TypeIn: "io.Writer", TypeOut: "Writer", Function: "wrapWriter"},
		{TypeIn: "io.WriterAt", TypeOut: "WriterAt", Function: "wrapWriterAt"},
		{TypeIn: "io.WriterTo", TypeOut: "io.WriterTo"},
	}

	fmt.Print(`// Code generated by generate-wrap; DO NOT EDIT.
package contextaware

import (
	"io"
)

func wrapIO(i interface{}) interface{} {
`)
	fmt.Printf("\ttype (\n")
	for i, c := range converters {
		fmt.Printf("\t\tt%02di = %s\n", i, c.TypeIn)
		if c.TypeIn != c.TypeOut {
			fmt.Printf("\t\tt%02do = %s\n", i, c.TypeOut)
		}
	}
	fmt.Printf("\t)\n\n")

	fmt.Printf("\tvar (\n")
	for i, c := range converters {
		if c.TypeIn != c.TypeOut {
			fmt.Printf("\t\tf%02d = %s\n", i, c.Function)
		}
	}
	fmt.Printf("\t)\n\n")
	fmt.Printf("\tvar f uint64\n")

	for i := range converters {
		fmt.Printf("\to%02d, b%02d := i.(t%02di)\n", i, i, i)
		fmt.Printf("\tif b%02d { f |= 0x%04x }\n", i, pow2(i))
	}
	fmt.Printf("\n")

	fmt.Printf("\tswitch f {\n")
	for i := 0; i < pow2(len(converters)); i++ {
		fmt.Printf("\tcase 0x%04x: return struct{", i)
		addSep := false
		for j, c := range converters {
			if pow2(j)&i > 0 {
				if addSep {
					fmt.Printf(";")
				}
				if c.TypeIn == c.TypeOut {
					fmt.Printf("t%02di", j)
				} else {
					fmt.Printf("t%02do", j)
				}
				addSep = true
			}
		}
		fmt.Printf("}{")
		addSep = false
		for j, c := range converters {
			if pow2(j)&i > 0 {
				if addSep {
					fmt.Printf(",")
				}
				if c.TypeIn == c.TypeOut {
					fmt.Printf("o%02d", j)
				} else {
					fmt.Printf("f%02d(o%02d)", j, j)
				}
				addSep = true
			}
		}
		fmt.Printf("}\n")
	}
	fmt.Printf("\t}\n\n")

	fmt.Printf("\tpanic(\"unreachable\")\n")
	fmt.Println("}")
}

func pow2(i int) int {
	if i == 0 {
		return 1
	}
	return 2 << (i - 1)
}
