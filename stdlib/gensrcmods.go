//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
)

var lokumModFileRE = regexp.MustCompile(`^srcmod_(\w+).lokum$`)

func main() {
	modules := make(map[string]string)
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		m := lokumModFileRE.FindStringSubmatch(file.Name())
		if m != nil {
			modName := m[1]

			src, err := ioutil.ReadFile(file.Name())
			if err != nil {
				log.Fatalf("dosya '%s' okuma hatası: %s",
					file.Name(), err.Error())
			}

			modules[modName] = string(src)
		}
	}

	var out bytes.Buffer
	out.WriteString(`// gensrcmods.go ile oluşturuldu, değiştirmeyin.

package stdlib

var SourceModules = map[string]string{` + "\n")
	for modName, modSrc := range modules {
		out.WriteString("\t\"" + modName + "\": " +
			strconv.Quote(modSrc) + ",\n")
	}
	out.WriteString("}\n")

	const target = "source_modules.go"
	if err := ioutil.WriteFile(target, out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
