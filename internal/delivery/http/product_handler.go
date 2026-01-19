package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/handokobeni/agent-openhands/internal/usecase"
	"github.com/handokobeni/agent-openhands/pkg/response"
)

type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/products")
	path = strings.TrimSuffix(path, "/")

	switch r.Method {
	case http.MethodGet:
		if path == "" {
			h.GetAll(w, r)
		} else {
			h.GetByID(w, r, path)
		}
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodPut:
		h.Update(w, r, path)
	case http.MethodDelete:
		h.Delete(w, r, path)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.productUseCase.Create(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrProductNameRequired, usecase.ErrInvalidPrice, usecase.ErrInvalidStock:
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.Success(w, http.StatusCreated, "product created successfully", product)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request, path string) {
	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.productUseCase.GetByID(r.Context(), id)
	if err != nil {
		if err == usecase.ErrProductNotFound {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.Success(w, http.StatusOK, "product retrieved successfully", product)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	products, err := h.productUseCase.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.Success(w, http.StatusOK, "products retrieved successfully", products)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request, path string) {
	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req usecase.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.productUseCase.Update(r.Context(), id, &req)
	if err != nil {
		switch err {
		case usecase.ErrProductNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		case usecase.ErrProductNameRequired, usecase.ErrInvalidPrice, usecase.ErrInvalidStock:
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.Success(w, http.StatusOK, "product updated successfully", product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request, path string) {
	idStr := strings.TrimPrefix(path, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	err = h.productUseCase.Delete(r.Context(), id)
	if err != nil {
		if err == usecase.ErrProductNotFound {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.Success(w, http.StatusOK, "product deleted successfully", nil)
}
