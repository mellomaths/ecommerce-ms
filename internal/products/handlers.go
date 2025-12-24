package products

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts(r.Context())
	if err != nil {
		log.Println(err)
		responses.NewJsonErrorResponse(w, http.StatusInternalServerError, "server_error", "unexpected error when listing products")
		return
	}
	responses.NewJsonResponse(w, http.StatusOK, products)
}

func (h *handler) FindProductById(w http.ResponseWriter, r *http.Request) {
	productId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		responses.NewJsonErrorResponse(w, http.StatusBadRequest, "validation_error", "invalid product id")
		return
	}
	product, err := h.service.FindProductById(r.Context(), productId)
	if product.ID == 0 {
		log.Println("product not found")
		responses.NewJsonErrorResponse(w, http.StatusNotFound, "not found", "product not found")
		return
	}
	responses.NewJsonResponse(w, http.StatusOK, product)
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productParams CreateProductParams
	if err := requests.DecodeJsonBody(r, &productParams); err != nil {
		log.Println(err)
		responses.NewJsonErrorResponse(w, http.StatusBadRequest, "validation_error", "invalid product")
		return
	}
	p, err := h.service.CreateProduct(r.Context(), productParams)
	if err != nil {
		log.Println(err)
		responses.NewJsonErrorResponse(w, http.StatusInternalServerError, "server_error", "unexpected error when creating a new product")
		return
	}
	responses.NewJsonResponse(w, http.StatusOK, p)
}
