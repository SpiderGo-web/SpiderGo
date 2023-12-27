package engine

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FindLine struct {
	lineNumber  int
	lineContent string
}

func BuildApp() {
	buildHtml("./src/App.spider", "")
}

func buildHtml(fileName string, tag string) {
	var tmplt *template.Template
	var out bytes.Buffer
	var htmlForFile []string
	var cleanHtml string
	f, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	htmlForFile = append(htmlForFile, strings.Split(string(f), "\n")...)
	if tag == "" {
		tmplt, err = template.ParseFiles("./engine/html.tmpl")
		if err != nil {
			log.Fatal(err)
		}
		data := struct {
			Body []string
		}{
			Body: htmlForFile,
		}
		err = tmplt.Execute(&out, data)
		if err != nil {
			log.Fatal(err)
		}
		cleanHtml = html.UnescapeString(out.String())
	} else {
		importHtml := strings.Join(htmlForFile, "\n                ") //fix indention
		file, err := os.ReadFile("./public/App.html")
		if err != nil {
			log.Fatal(err)
		}
		lines := strings.Split(string(file), "\n")
		tagLine, err := findLineInFile(fmt.Sprintf("<%v />", tag), file)
		if err != nil && err.Error() == "import not found" {
			return
		} else if err != nil {
			log.Fatal(err)
		}
		lines[tagLine[0].lineNumber-1] = fmt.Sprintf("                %v", importHtml) //fix indention
		fmt.Println(lines[tagLine[0].lineNumber-1])
		cleanHtml = strings.Join(lines, "\n")
	}
	createAppHtml(cleanHtml)
	importsLine, err := findLineInFile("import", f)
	if err != nil && err.Error() == "import not found" {
		return
	} else if err != nil {
		log.Fatal(err)
	}
	for _, line := range importsLine {
		importLine, err := readFileByLines(line.lineNumber, line.lineNumber, f)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(line)
		importFile := strings.Split(importLine[0], "from")
		trimmedImportFile := strings.Trim(importFile[1], " \"")
		importTag := strings.Split(importFile[0], " ")
		buildHtml(trimmedImportFile, importTag[1])
	}
}

func createAppHtml(html string) {
	err := os.Remove("./public/App.html")
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile("./public/App.html", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(html)
	if err != nil {
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

func findLineInFile(s string, f []byte) ([]FindLine, error) {
	var lineFound []FindLine
	temp := strings.Split(string(f), "\n")

	for i, line := range temp {
		if strings.Contains(line, s) {
			lineFound = append(lineFound, FindLine{i + 1, line})
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
