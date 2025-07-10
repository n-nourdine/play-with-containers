package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/n-nourdine/play-with-containers/billing-app/database"
	"github.com/n-nourdine/play-with-containers/billing-app/util"
)

type Handler struct {
	C *database.OrderStore
}

func NewHandler(l *log.Logger) (*Handler, error) {
	c, err := database.NewConn()
	if err != nil {
		return nil, err
	}
	return &Handler{C: c}, nil
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Billing service is healthy"))
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	order := database.Order{}
	if err := util.FromJSON(&order, r.Body); err != nil {
		log.Println(err)
		http.Error(w, "invalide order", http.StatusBadRequest)
		return
	}

	order.ID = util.NewUUID()
	order.UserID = util.NewUUID()

	if err := h.C.CreateOrder(ctx, order); err != nil {
		log.Println(err)
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Délai d'attente dépassé ", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orders, err := h.C.GetAllOrders(ctx)
	if err != nil {
		if err != fmt.Errorf("sql no row") {
			util.ToJSON("No orders found", w)
			return
		}
		log.Printf("Erreur lors de la récupération des commandes: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	if err := util.ToJSON(orders, w); err != nil {
		log.Printf("Erreur lors de la sérialisation JSON: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
}
