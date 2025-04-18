package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/validate", handleValidate)
	fmt.Println("🚀 Webhook server running on port 8443...")
	err := http.ListenAndServeTLS(":8443", "/tls/tls.crt", "/tls/tls.key", nil)
	if err != nil {
		panic(err)
	}
}
