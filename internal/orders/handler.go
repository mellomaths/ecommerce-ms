package orders

import (
	"log"
	"net/http"

	"github.com/mellomaths/ecommerce-ms/internal/requests"
	"github.com/mellomaths/ecommerce-ms/internal/responses"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var orderParams CreateOrderParams
	if err := requests.DecodeJsonBody(r, &orderParams); err != nil {
		log.Println(err)
		responses.NewJsonErrorResponse(w, http.StatusBadRequest, "validation_error", "invalid order")
		return
	}
	o, err := h.service.PlaceOrder(r.Context(), orderParams)
	if err != nil {
		log.Println(err)
		if err == ErrProductNotFound {
			responses.NewJsonErrorResponse(w, http.StatusNotFound, "validation_error", err.Error())
			return
		}
		if err == ErrProductNoStock {
			responses.NewJsonErrorResponse(w, http.StatusExpectationFailed, "validation_error", err.Error())
			return
		}
		if err == ErrInvalidOrder {
			responses.NewJsonErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		responses.NewJsonErrorResponse(w, http.StatusInternalServerError, "server_error", "unexpected error when placing a new order")
		return
	}
	responses.NewJsonResponse(w, http.StatusCreated, o)
}
