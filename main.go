package main

import (
    "log"
    "net/http"
    "time"
    "fmt"
)

func main() {
    http.HandleFunc("/current_time", func(w http.ResponseWriter, r *http.Request) {
        updatedContent := "Updated content at " + time.Now().Format(time.RFC1123)
        fmt.Fprint(w, updatedContent)
    })

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)

    log.Println("Listening ...")
    err := http.ListenAndServeTLS(":8443", "certs/lynas_dev.crt", "certs/lynas_dev.key", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
