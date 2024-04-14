package delete

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	resp "github.com/JustForWorld/banner-shift/internal/http-server/handlers"
	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Request struct {
	ID int64 `json:"id"`
}

type Response struct {
	resp.Response
}

type BannerRemove interface {
	DeleteBanner(ctx context.Context, bannerID int64) error
}

func New(log *slog.Logger, bannerRemove BannerRemove) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		_, claims, _ := jwtauth.FromContext(r.Context())
		fmt.Println(claims)
		if claims["role"] != "admin" {
			render.Status(r, 403)
			render.JSON(w, r, resp.Error("Пользователь не имеет доступа"))
			return
		}

		var (
			req Request
			err error
		)
		req.ID, err = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Error("request path parameter id is not integer")
			fmt.Println(chi.URLParam(r, "id"))

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		log.Info("path parameter is valid", slog.Any("request", req))

		err = bannerRemove.DeleteBanner(r.Context(), req.ID)
		if errors.Is(err, storage.ErrBannerNotExists) {
			log.Info("banner not found",
				slog.Any("banner_id", req.ID),
			)

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

		fmt.Fprintln(w, http.StatusNoContent)
	}
}
