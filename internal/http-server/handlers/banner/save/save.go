package save

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	resp "github.com/JustForWorld/banner-shift/internal/http-server/handlers"
	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/JustForWorld/banner-shift/internal/storage/redis"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Request struct {
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int64           `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
}

type Response struct {
	resp.Response
	BannerID int64 `json:"banner_id"`
}

type BannerSaver interface {
	CreateBanner(ctx context.Context, redis *redis.Storage, featureID int64, tagIDs []int, content []byte, isActive bool) (int64, error)
}

func New(log *slog.Logger, bannerSaver BannerSaver, redis *redis.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.save.New"

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

		var req Request
		err := render.DecodeJSON(r.Body, &req)
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

		var res Response
		res.BannerID, err = bannerSaver.CreateBanner(r.Context(), redis, req.FeatureID, req.TagIDs, req.Content, req.IsActive)
		if errors.Is(err, storage.ErrBannerInvalidData) {
			log.Warn("banner with invalid data",
				slog.Any("feature_id", req.FeatureID),
				slog.Any("tag_ids", req.TagIDs),
				slog.Any("content", req.Content),
				slog.Any("is_active", req.IsActive),
				slog.String("error", err.Error()),
			)

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		if errors.Is(err, storage.ErrBannerExists) {
			log.Info("banner exists",
				slog.Any("feature_id", req.FeatureID),
				slog.Any("tag_ids", req.TagIDs),
				slog.Any("content", req.Content),
				slog.Any("is_active", req.IsActive),
			)

			render.Status(r, 409)
			render.JSON(w, r, resp.Error("Баннер уже существует"))
			return
		}

		if err != nil {
			fmt.Println(err)
			log.Error("failed to create banner", err)

			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Внутренняя ошибка сервера"))
			return
		}

		log.Info("banner created", slog.Int64("id", res.BannerID))

		render.Status(r, 201)
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, fmt.Sprintf(`{"banner_id": %v}`, res.BannerID))
	}
}
