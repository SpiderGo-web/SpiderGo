package engine

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type FindLine struct {
	lineNumber  int
	lineContent string
}

type GoCmd struct {
	cmdName  string
	cmdValue string
}

var randClassIdStrings []string

func BuildApp() {
	buildHtml("./src/App.spider", "")
}

func buildHtml(fileName string, tag string) {
	var tmplt *template.Template
	var out bytes.Buffer
	var htmlForFile []string
	var cleanHtml string
	var goCmds []GoCmd
	var updateWeb []string
	var updateStyle []string
	var ifBlockIndex []int
	var ifLogic string
	var ifBlock []string
	f, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	randClassString := randString(16)
	randClassIdStrings = append(randClassIdStrings, randClassString)
	goTagOpen, err := findLineInFile("<go>", f)
	if err != nil {
		log.Fatal(err)
	}
	goTagClose, err := findLineInFile("</go>", f)
	if err != nil {
		log.Fatal(err)
	}
	loadGoCode, err := readFileByLines(goTagOpen[0].lineNumber, goTagClose[0].lineNumber, f)
	if err != nil {
		log.Fatal(err)
	}
	removeGoTag := loadGoCode[1 : len(loadGoCode)-1]
	goCmds = goCompiler(removeGoTag)
	styleTagOpen, err := findLineInFile("<style>", f)
	if err != nil {
		log.Fatal(err)
	}
	styleTagClose, err := findLineInFile("</style>", f)
	if err != nil {
		log.Fatal(err)
	}
	loadstyleCode, err := readFileByLines(styleTagOpen[0].lineNumber, styleTagClose[0].lineNumber, f)
	if err != nil {
		log.Fatal(err)
	}
	removeStyleTag := loadstyleCode[1 : len(loadstyleCode)-1]
	for i, l := range removeStyleTag {
		if len(updateStyle) < 1 {
			updateStyle = removeStyleTag
		}
		if strings.Contains(l, " {") {
			class := strings.Split(l, " {")
			class[0] = class[0] + "-" + randClassString
			updateStyle[i] = strings.Join(class, " {")
		} else {
			updateStyle[i] = l
		}
	}
	createSpiderCss(strings.Join(updateStyle, "\n"))
	webTagOpen, err := findLineInFile("<web>", f)
	if err != nil {
		log.Fatal(err)
	}
	webTagClose, err := findLineInFile("</web>", f)
	if err != nil {
		log.Fatal(err)
	}
	loadWebCode, err := readFileByLines(webTagOpen[0].lineNumber, webTagClose[0].lineNumber, f)
	if err != nil {
		log.Fatal(err)
	}
	removeWebTag := loadWebCode[1 : len(loadWebCode)-1]
	for i, l := range removeWebTag {
		randAdded := false
		if len(updateWeb) < 1 {
			updateWeb = removeWebTag
		}
		if strings.Contains(l, "class=") {
			class := strings.Split(l, "\"")
			class[1] = class[1] + "-" + randClassString
			updateWeb[i] = strings.Join(class, "\"")
			randAdded = true
		}
		if strings.Contains(l, "id=") {
			id := strings.Split(l, "\"")
			id[1] = id[1] + "-" + randClassString
			updateWeb[i] = strings.Join(id, "\"")
			randAdded = true
		}
		if strings.Contains(l, "{#if") {
			ifLogic = strings.Split(strings.Split(l, "{#if ")[1], "}")[0]
			endIf, err := findLineInFile("{/if}", []byte(strings.Join(removeWebTag, "\n")))
			if err != nil {
				log.Fatal(err)
			}
			ifBlock, err = readFileByLines(i+2, endIf[0].lineNumber-1, []byte(strings.Join(removeWebTag, "\n")))
			if err != nil {
				log.Fatal(err)
			}
			for j := i + 1; j < endIf[0].lineNumber-1; j++ {
				ifBlockIndex = append(ifBlockIndex, j)
			}
		}
		for _, t := range goCmds {
			if slices.Contains(ifBlockIndex, i) {
				if ifLogic == t.cmdName && t.cmdValue == "true" {
					updateWeb[i] = strings.Join(ifBlock, "\n")
				} else {
					updateWeb[i] = ""
				}
			} else if strings.Contains(l, "{#if ") || strings.Contains(l, "{/if}") {
				updateWeb[i] = ""
			} else if strings.Contains(l, t.cmdName) && strings.Contains(l, "{{") {
				start := strings.Split(l, "{{")[0]
				end := strings.Split(l, "}}")[1]
				updateWeb[i] = start + strings.Trim(t.cmdValue, "\"") + end
			} else if !randAdded && !strings.Contains(l, "{{") && !strings.Contains(l, "{#if ") && !strings.Contains(l, "{/if}") {
				updateWeb[i] = l
			}
		}
	}
	var test []string
	for _, l := range updateWeb {
		if l != "" {
			test = append(test, l)
		}
	}
	htmlForFile = append(htmlForFile, test...)
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
		file, err := os.ReadFile("./public/spider.html")
		if err != nil {
			log.Fatal(err)
		}
		lines := strings.Split(string(file), "\n")
		tagLine, err := findLineInFile(fmt.Sprintf("<%v />", tag), file)
		indentionCount := len(strings.Split(tagLine[0].lineContent, "<")[0])
		importHtml := strings.Join(htmlForFile, ("\n" + strings.Repeat(" ", indentionCount))) //fix indention
		if err != nil && err.Error() == "import not found" {
			return
		} else if err != nil {
			log.Fatal(err)
		}
		lines[tagLine[0].lineNumber-1] = fmt.Sprintf((strings.Repeat(" ", indentionCount) + "%v"), importHtml) //fix indention
		cleanHtml = strings.Join(lines, "\n")
	}
	createSpiderHtml(cleanHtml)
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
		importFile := strings.Split(importLine[0], "from")
		trimmedImportFile := strings.Trim(importFile[1], " \"")
		importTag := strings.Split(importFile[0], " ")
		buildHtml(trimmedImportFile, importTag[1])
	}
}

func createSpiderHtml(html string) {
	err := os.Remove("./public/spider.html")
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile("./public/spider.html", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(html)
	if err != nil {
		log.Fatal(err)
	}
}

func createSpiderCss(css string) {
	var currentCss []string
	oldCssFile, err := os.ReadFile("./public/spider.css")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove("./public/spider.css")
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile("./public/spider.css", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	currentCssFile := strings.FieldsFunc(string(oldCssFile), splitCssFile)
	for i, sel := range currentCssFile {
		if strings.Contains(sel, "-") && strings.Contains(strings.Join(randClassIdStrings, "\n"), strings.Trim(strings.Split(sel, "-")[1], " ")) {
			currentCss = append(currentCss, currentCssFile[i]+"{"+currentCssFile[i+1]+"\n    }")
		}
	}
	joinedCss := css + "\n" + strings.Join(currentCss, "\n")
	_, err = file.WriteString(joinedCss)
	if err != nil {
		log.Fatal(err)
	}
}

func splitCssFile(r rune) bool {
	return string(r) == "{" || string(r) == "}"
}

func readFileByLines(start int, finish int, file []byte) ([]string, error) {
	var lines []string
	fileLines := strings.Split(string(file), "\n")
	for i := start - 1; i < finish; i++ {
		lines = append(lines, fileLines[i])
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

func goCompiler(goCode []string) []GoCmd {
	var commands []GoCmd
	for _, line := range goCode {
		// get vars
		if strings.Contains(line, ":=") {
			l := strings.Split(line, ":=")
			commands = append(commands, GoCmd{strings.Trim(l[0], " "), strings.TrimLeft(l[1], " ")})
		}
	}
	return commands
}

func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz1234567890"
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
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
