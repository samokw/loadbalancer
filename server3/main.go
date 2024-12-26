package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	homeHandler := func(w http.ResponseWriter, r *http.Request) {
		l := log.New(os.Stdout, "[Server3] ", log.Ldate|log.Ltime)
		l.Printf("running...")
		io.WriteString(w, "hi\n")
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("starting server 3 ..... listening on server :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
