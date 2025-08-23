package delivery

import (
	"net/http"

	u "remarks/internal/usecase"
)

type WebHandler struct {
	uc u.UsecaseInterface
}

func NewWebHandler(uc u.UsecaseInterface) *WebHandler {
	return &WebHandler{
		uc: uc,
	}
}

func ReturnErrorJSON(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	//json.NewEncoder(w).Encode(&model.Error{Error: err.Error()})
	return
}

func (api *WebHandler) LoadExcelRegistry(w http.ResponseWriter, r *http.Request) {
	//json.NewEncoder(w).Encode(&model.Response{})
}
