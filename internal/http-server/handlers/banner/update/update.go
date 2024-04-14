package update

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	BannerID int64 `json:"id"`
	Banner
}

type Response struct {
	resp.Response
}

type Banner struct {
	TagIDs    []int       `json:"tag_ids"`
	FeatureID int64       `json:"feature_id"`
	Content   interface{} `json:"content"`
	IsActive  interface{} `json:"is_active"`
}

type BannerUpdater interface {
	UpdateBanner(ctx context.Context, bannerID int64, featureID int64, tagIDs []int, content interface{}, isActive interface{}) error
}

func New(log *slog.Logger, bannerUpdater BannerUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.update.New"

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
		req.BannerID, err = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Error("request path parameter id is not integer")
			fmt.Println(chi.URLParam(r, "id"))

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		log.Info("path parameter is valid", slog.Any("request", req))

		err = render.DecodeJSON(r.Body, &req.Banner)
		// checking for an empty request body
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", err)

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		err = bannerUpdater.UpdateBanner(r.Context(), req.BannerID, req.Banner.FeatureID, req.Banner.TagIDs, req.Banner.Content, req.Banner.IsActive)
		if errors.Is(err, storage.ErrBannerInvalidData) {
			log.Info("banner with invalid data",
				slog.Any("feature_id", req.Banner.FeatureID),
				slog.Any("tag_ids", req.Banner.TagIDs),
				slog.Any("content", req.Banner.Content),
				slog.Any("is_active", req.Banner.IsActive),
			)

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		if errors.Is(err, storage.ErrBannerNotExists) {
			log.Info("banner not exists",
				slog.Any("feature_id", req.Banner.FeatureID),
				slog.Any("tag_ids", req.Banner.TagIDs),
				slog.Any("content", req.Banner.Content),
				slog.Any("is_active", req.Banner.IsActive),
			)

			render.Status(r, 404)
			render.JSON(w, r, resp.Error("Баннер не найден"))
			return
		}
		if err != nil {
			fmt.Println(err)
			log.Error("failed to update banner", err)

			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Внутренняя ошибка сервера"))
			return
		}

		log.Info("banner update", slog.Int64("id", req.BannerID))

		fmt.Fprintln(w, http.StatusOK)
	}
}
