package get

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	resp "github.com/JustForWorld/banner-shift/internal/http-server/handlers"
	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type Request struct {
	TagID     int64 `json:"tag_id"`
	FeatureID int64 `json:"feature_id"`
}

type Response struct {
	resp.Response
	Content json.RawMessage `json:"content"`
}

type BannerGetter interface {
	GetBanner(ctx context.Context, tagID, featureID int64) (json.RawMessage, error)
}

type BannerGetterCache interface {
	GetBanner(ctx context.Context, key string) ([]byte, error)
}

func New(log *slog.Logger, bannerGetter BannerGetter, bannerGetterCache BannerGetterCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		_, claims, _ := jwtauth.FromContext(r.Context())
		log.Debug("user auth", claims["username"], claims["tag"])
		if claims["role"] == "user" {
			TagID, err := strconv.ParseInt(r.URL.Query().Get("tag_id"), 10, 64)
			if err != nil {
				log.Error("request query parameter tag_id is not integer")

				render.Status(r, 400)
				render.JSON(w, r, resp.Error("Некорректные данные"))
				return
			}
			log.Info("request query parameter is valid")

			userTag, ok := claims["tag"].(int)
			if ok && userTag != int(TagID) {
				render.Status(r, 404)
				render.JSON(w, r, resp.Error(fmt.Sprintf("Баннер для %v не найден", claims["username"])))
				return
			}
		}

		var (
			req Request
			err error
		)
		req.FeatureID, err = strconv.ParseInt(r.URL.Query().Get("feature_id"), 10, 64)
		if err != nil {
			log.Error("request query parameter feature_id is not integer")

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}

		req.TagID, err = strconv.ParseInt(r.URL.Query().Get("tag_id"), 10, 64)
		if err != nil {
			log.Error("request query parameter tag_id is not integer")

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		log.Info("request query parameter is valid", slog.Any("request", req))

		var res Response
		// check if false get from Redis
		LastRevision := r.URL.Query().Get("use_last_revision")
		if (LastRevision == "false" || LastRevision == "") && LastRevision != "true" {
			value, err := bannerGetterCache.GetBanner(r.Context(), fmt.Sprintf("%v:%v", req.TagID, req.FeatureID))
			if errors.Is(err, storage.ErrBannerNotFound) {
				log.Info("banner not found in Redis",
					slog.Any("feature_id", req.FeatureID),
					slog.Any("tag_id", req.TagID),
				)
			}
			if err == nil {
				log.Info("banner found", slog.Any("content", res.Content))

				render.Status(r, 200)
				render.JSON(w, r, string(value))
				return
			}
		}

		res.Content, err = bannerGetter.GetBanner(r.Context(), req.TagID, req.FeatureID)
		if errors.Is(err, storage.ErrBannerInvalidData) {
			log.Info("banner with invalid fata",
				slog.Any("feature_id", req.FeatureID),
				slog.Any("tag_id", req.TagID),
			)

			render.Status(r, 400)
			render.JSON(w, r, resp.Error("Некорректные данные"))
			return
		}
		if errors.Is(err, storage.ErrBannerNotFound) {
			log.Info("banner not found",
				slog.Any("feature_id", req.FeatureID),
				slog.Any("tag_id", req.TagID),
			)

			render.Status(r, 404)
			render.JSON(w, r, resp.Error(fmt.Sprintf("Баннер для %v не найден", claims["username"])))
			return
		}
		if err != nil {
			log.Error("failed to create banner", err)

			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Внутренняя ошибка сервера"))
			return
		}

		log.Info("banner found", slog.Any("content", res.Content))

		render.Status(r, 200)
		render.JSON(w, r, res.Content)
	}
}
