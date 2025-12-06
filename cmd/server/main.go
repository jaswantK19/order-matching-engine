package main

import (
	"fmt"
	"net/http"

	"github.com/jaswantK19/order-matching-engine/internal/api"
)

func main() {
	mux := api.SetupServer()
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}
