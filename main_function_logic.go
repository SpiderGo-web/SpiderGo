package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"syscall/js"
)

type News struct {
	Headline string
	Body     string
}

var tmplt *template.Template
var done = make(chan struct{})

func getHtml() string {
	var out bytes.Buffer
	event := News{
		Headline: "makeuseof.com has everything Tech",
		Body:     "Visit MUO for anything technology related",
	}
	tmplt, err := template.New("test.tmpl").ParseFiles("test.tmpl")
	if err != nil {
		panic(err)
	}
	err = tmplt.Execute(&out, event)
	if err != nil {
		panic(err)
	}
	return out.String()
}

func not_main() {
	newHtml, err := os.ReadFile("test.html")
	if err != nil {
		log.Fatal(err)
	}
	s := string(newHtml)
	js.Global().Get("document").Call("getElementById", "root").Set("innerHTML", s)
}
