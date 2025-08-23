package main

import (
	"fmt"
	"net/http"
	"remarks/internal/httputils"
)

func main() {
	http.HandleFunc("/health", httputils.HealthHandler)
	http.HandleFunc("/load_excel_reestr", httputils.HealthHandler)
	fmt.Println("remarks service running on :8080")
	httputils.ParseExcel("../../internal/httputils/РЕЕСТР_ягодное - нулевой.xlsx")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
