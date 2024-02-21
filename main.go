package main

import (
    "crypto/tls"
    "log"
    "net/http"
)

func main() {
    staticFiles := http.FileServer(http.Dir("lynas_dev_static"))
    http.Handle("lynas.dev/", staticFiles)

    http.HandleFunc("polytopia.lynas.dev/", polytopiaHandler)

    log.Println("Listening ...")
     server := &http.Server{
        Addr: ":8443",
        TLSConfig: &tls.Config{
            GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
                var certPath, keyPath string
                if info.ServerName == "polytopia.lynas.dev" {
                    certPath = "certs/polytopia_lynas_dev.crt"
                    keyPath = "certs/polytopia_lynas_dev.key"
                } else {
                    certPath = "certs/lynas_dev.crt"
                    keyPath = "certs/lynas_dev.key"
                }
                cert, err := tls.LoadX509KeyPair(certPath, keyPath)
                if err != nil {
                    return nil, err
                }
                return &cert, nil
            },
        },
    }
    log.Fatal(server.ListenAndServeTLS("", ""))
}
