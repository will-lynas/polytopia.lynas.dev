package main

import (
    "log"
    "net/http"
)

func main() {
    staticFiles := http.FileServer(http.Dir("static"))
    http.Handle("/", staticFiles)

    http.HandleFunc("/polytopia", polytopiaHandler)
    http.HandleFunc("/polytopia.html", polytopiaHandler)

    log.Println("Listening ...")
    err := http.ListenAndServeTLS(":8443",
    "certs/lynas_dev.crt",
    "certs/lynas_dev.key", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
