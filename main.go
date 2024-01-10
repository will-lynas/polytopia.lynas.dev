package main

import (
    "log"
    "net/http"
)

func main() {
     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/construction.jpg", http.StatusFound)
    })

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/construction.jpg", fs)

    log.Println("Listening ...")
    err := http.ListenAndServeTLS(":8443", "certs/lynas_dev.crt", "certs/lynas_dev.key", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
