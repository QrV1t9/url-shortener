package deleter

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "go-urlshortner/internal/lib/api/response"
	"go-urlshortner/internal/lib/logger/sl"
	"go-urlshortner/internal/storage"
	"log/slog"
	"net/http"
)

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.deleter.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("not found"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Info("failed to delete url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("deleted url", slog.String("alias", alias))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}