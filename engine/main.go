package engine

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func BuildApp() {
	buildHtml("./src/App.spider")
}

func buildHtml(fileName string) {
	var fileContent []byte
	var out bytes.Buffer
	f, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	importsLine, _ := findLineInFile("import", f)
	for _, line := range importsLine {
		importLine, err := readFileByLines(line, line, f)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(importLine)
		importFile := strings.Split(importLine[0], "from")
		trimmedImportFile := strings.Trim(importFile[1], " \"")
		buildHtml(trimmedImportFile)
	}
	if err != nil {
		log.Fatal(err)
	}
	tmplt, err := template.New("./engine/html.tmpl").ParseFiles("./engine/html.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	data := struct {
		Body []byte
	}{
		Body: f,
	}
	err = tmplt.Execute(&out, data)
	if err != nil {
		log.Fatal(err)
	}
	fileContent = out.Bytes()
	if err := os.WriteFile("./public/App.html", fileContent, 0666); err != nil {
		log.Fatal(err)
	}
}

func readFileByLines(start int, finish int, file []byte) ([]string, error) {
	length := (finish + 1) - start
	lines := make([]string, length)
	fileLines := strings.Split(string(file), "\n")
	for i := start - 1; i < length; i++ {
		lines[i] = fileLines[i]
	}
	return lines, nil
}

func findLineInFile(s string, f []byte) ([]int, error) {
	var lineFound []int
	temp := strings.Split(string(f), "\n")

	for i, line := range temp {
		if strings.Contains(line, s) {
			lineFound = append(lineFound, i+1)
		}
	}
	if len(lineFound) < 1 {
		return lineFound, fmt.Errorf("%v not found", s)
	}

	return lineFound, nil
}

func findSpiderFiles(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}
