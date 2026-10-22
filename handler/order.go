package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"mcrsrv_go/model"
	"mcrsrv_go/repository/order"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	decimal         = 10
	base            = 10
	bitSize         = 64
	size            = 50
	completedStatus = "completed"
	shippedStatus   = "shipped"
)

var now = time.Now().UTC()

var response struct {
	Items []model.Order `json:"items"`
	Next  uint64        `json:"next,omitempty"`
}

var body struct {
	CustomerId uuid.UUID        `json:"customer_id"`
	LineItems  []model.LineItem `json:"line_items"`
}

var statusBody struct {
	Status string `json:"status"`
}

type Order struct {
	Repo *order.RedisRepo
}

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	order := model.Order{
		OrderId:   rand.Uint64(),
		CreatorId: body.CustomerId,
		LineItems: body.LineItems,
		CreatedAt: &now,
	}

	err := h.Repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("failed to insert order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("failed to marshall:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(res)

}

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.Repo.FindAll(r.Context(), model.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("failed to find all:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshall:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)

}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := h.Repo.FindByID(r.Context(), orderId)
	if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(order); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *Order) Update(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&statusBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := h.Repo.FindByID(r.Context(), orderId)
	if errors.Is(err, model.ErrNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch statusBody.Status {
	case shippedStatus:
		if order.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		order.ShippedAt = &now
	case completedStatus:
		if order.CompletedAt != nil || order.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		order.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.Update(r.Context(), order)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(order); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *Order) Delete(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.Delete(r.Context(), orderId)
	if errors.Is(err, model.ErrNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
