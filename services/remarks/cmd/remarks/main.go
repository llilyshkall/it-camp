package main

import (
	"fmt"
	"net/http"
	"remarks/internal/httputils"
)

func main() {
	http.HandleFunc("/health", httputils.HealthHandler)
	fmt.Println("remarks service running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
