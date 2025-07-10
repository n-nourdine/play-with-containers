package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/n-nourdine/play-with-containers/inventory-app/database"
	"github.com/n-nourdine/play-with-containers/inventory-app/util"
)

type Handler struct {
	L *log.Logger
	C *database.MovieStream
}

func NewHandler(l *log.Logger) (*Handler, error) {
	c, err := database.NewConn()
	if err != nil {
		return nil, err
	}
	return &Handler{L: l, C: c}, nil
}

func (h *Handler) GetMovies(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if title := r.URL.Query().Get("title"); title != "" {
		movies, err := h.C.ListeByTitle(ctx, title)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				http.Error(rw, "Délai d'attente dépassé lors de la récupération des films", http.StatusGatewayTimeout)
				h.L.Panic(err)
				return
			}
			h.L.Panic(err)
			http.Error(rw, "Aucun Film trouvé", http.StatusNotFound)
			return
		}

		if err = util.Tojson(movies, rw); err != nil {
			h.L.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		h.L.Printf("Film trouvés: %v\n", movies)
		return
	}

	if strings.HasSuffix(r.URL.String(), "movies") {
		movies, err := h.C.Liste(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				h.L.Println("Délai d'attente dépassé")
				http.Error(rw, "Délai d'attente dépassé lors de la récupération des films", http.StatusGatewayTimeout)
				return
			}
			h.L.Println(err)
			http.Error(rw, "Aucun Film trouvé", http.StatusNotFound)
			return
		}

		if err = util.Tojson(movies, rw); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			h.L.Println(err)
			return
		}
		h.L.Printf("Film trouvés: %v\n", movies)
		return
	}
}

func (h *Handler) GetMovie(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id := r.PathValue("id")
	if id == "" {
		h.L.Panic("ID invalide")
		http.Error(rw, "ID invalide", http.StatusBadRequest)
		return
	}

	movies, err := h.C.GetById(ctx, id)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			h.L.Println("Délai d'attente dépassé")
			http.Error(rw, "Délai d'attente dépassé lors de la récupération du film", http.StatusGatewayTimeout)
			return
		}
		h.L.Println(err)
		http.Error(rw, "Aucun Film trouvé", http.StatusNotFound)
		return
	}

	if err = util.Tojson(movies, rw); err != nil {
		h.L.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	h.L.Printf("Film trouvé: %v\n", movies)

}

func (h *Handler) AddMovie(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	movie := database.Movies{}

	if err := util.FromJson(&movie, r.Body); err != nil {
		h.L.Println(err)
		http.Error(rw, "invalide movie", http.StatusBadRequest)
		return
	}

	if movie.Title == "" {
		h.L.Println("require title")
		http.Error(rw, "require title", http.StatusBadRequest)
		return
	}

	movie.ID = util.NewUUID()
	if err := h.C.Add(ctx, movie); err != nil {
		h.L.Println(err)
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(rw, "Délai d'attente dépassé lors de la création du film", http.StatusGatewayTimeout)
			return
		}
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := util.Tojson(movie, rw); err != nil {
		h.L.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	h.L.Printf("Movie added: %v\n", movie)

}

func (h *Handler) UpdateMovie(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id := r.PathValue("id")
	if id == "" {
		h.L.Println("invalide ID")
		http.Error(rw, "ID invalide", http.StatusBadRequest)
		return
	}
	var movie database.Movies
	if err := util.FromJson(&movie, r.Body); err != nil {
		h.L.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	movie.ID = id
	if movie.Title == "" || movie.Description == "" {
		h.L.Println("Champs invalid")
		http.Error(rw, "invalide fields", http.StatusBadRequest)
		return
	}
	if err := h.C.Update(ctx, movie); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(rw, "Délai d'attente dépassé lors de la mise à jour du film", http.StatusGatewayTimeout)
			return
		}
		h.L.Println(err)
		http.Error(rw, "impossible de faire la mise à jour", http.StatusInternalServerError)
		return
	}

	if err := util.Tojson(movie, rw); err != nil {
		h.L.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	h.L.Printf("Movie updated: %v]\n", movie)
}
func (h *Handler) DeleteMovie(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	id := r.PathValue("id")

	if id == "" {
		h.L.Println("ID manquant")
		http.Error(rw, "id manquant", http.StatusBadRequest)
		return
	}

	if err := h.C.Delete(ctx, id); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(rw, "Délai d'attente dépassé lors de la suppression du film", http.StatusGatewayTimeout)
			return
		}
		http.Error(rw, "Film non trouvé", http.StatusNotFound)
		return
	}
}
func (h *Handler) DeleteMovies(rw http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	confirmHeader := r.Header.Get("Confirm-Delete")
	if confirmHeader != "yes" {
		http.Error(rw, "Cette opération requiert un en-tête 'Confirm-Delete: yes'", http.StatusBadRequest)
		return
	}

	if err := h.C.DeleteAll(ctx); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(rw, "Délai d'attente dépassé lors de la suppression du film", http.StatusGatewayTimeout)
			return
		}
		http.Error(rw, "Film non trouvé", http.StatusNotFound)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}
