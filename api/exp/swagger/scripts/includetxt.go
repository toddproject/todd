package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Reads all .json files in the current folder
// and encodes them as strings literals in textfiles.go
func main() {
	fs, _ := ioutil.ReadDir("../definitions")
	out, _ := os.Create("swagger.pb.go")
	out.Write([]byte("package swagger \n\nconst (\n"))
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimPrefix(f.Name(), "service.")
			out.Write([]byte(strings.Title(strings.TrimSuffix(name, ".swagger.json") + " = `")))
			f, _ := os.Open(fmt.Sprintf("../definitions/%s", f.Name()))
			io.Copy(out, f)
			out.Write([]byte("`\n"))
		}
	}
	out.Write([]byte(")\n"))
}
