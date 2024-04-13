package delete

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	resp "github.com/JustForWorld/banner-shift/internal/http-server/handlers"
	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Request struct {
	ID int64 `json:"id"`
}

type Response struct {
	resp.Response
}

type BannerRemove interface {
	DeleteBanner(bannerID int64) error
}

func New(log *slog.Logger, bannerRemove BannerRemove) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var (
			req Request
			err error
		)
		req.ID, err = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		// TODO: remove
		fmt.Printf("%T, %T %v\n", req.ID, err, req.ID)
		if err != nil {
			log.Error("request path parameter id is not integer")
			fmt.Println(chi.URLParam(r, "id"))

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		log.Info("path parameter is valid", slog.Any("request", req))

		err = bannerRemove.DeleteBanner(req.ID)
		log.Info(err.Error())
		if errors.Is(err, storage.ErrBannerNotExists) {
			log.Info("banner not found", log.With(
				slog.Any("banner_id", req.ID),
			))

			// TODO: check user tag in banner tags
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("Баннер для тэга не найден"))
			return
		}
		if err != nil {
			log.Error("failed to create banner", err)

			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Внутренняя ошибка сервера"))
			return
		}

		log.Info("banner removed", slog.Any("id", req.ID))

		render.Status(r, 204)
		render.JSON(w, r, nil)
	}
}
