package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("template/demo.html")

	t.Execute(w, nil)
}

func main() {
	fmt.Println("启动 Demo 应用")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", index)
	http.ListenAndServe(":8310", nil)
}
