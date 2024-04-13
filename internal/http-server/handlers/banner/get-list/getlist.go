package getlist

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	response "github.com/JustForWorld/banner-shift/internal/http-server/handlers"
	"github.com/JustForWorld/banner-shift/internal/storage/postgresql"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	FeatureID int64 `json:"feature_id"`
	TagID     int64 `json:"tag_id"`
	Limit     int64 `json:"limit"`
	Offset    int64 `json:"offset"`
}

type Response struct {
	response.Response
	Banners []*postgresql.Banner
}

type BannerGetterList interface {
	GetBannerList(featureID, tagID, limit, offset int64) ([]*postgresql.Banner, error)
}

func New(log *slog.Logger, bannerGetterList BannerGetterList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.getList.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var (
			req Request
			err error
		)

		featureIDStr := r.URL.Query().Get("feature_id")
		if featureIDStr != "" {
			req.FeatureID, err = strconv.ParseInt(featureIDStr, 10, 64)
			if err != nil {
				log.Error("request query parameter feature_id is not integer")
				fmt.Printf("%T %v\n", req.FeatureID, req.FeatureID)

				render.Status(r, 400)
				render.JSON(w, r, response.Error("Некорректные данные"))
				return
			}
		}

		tagIDStr := r.URL.Query().Get("tag_id")
		if tagIDStr != "" {
			req.TagID, err = strconv.ParseInt(tagIDStr, 10, 64)
			if err != nil {
				log.Error("request query parameter tag_id is not integer")
				fmt.Printf("%T %v\n", req.TagID, req.TagID)

				render.Status(r, 400)
				render.JSON(w, r, response.Error("Некорректные данные"))
				return
			}
		}
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			req.Limit, err = strconv.ParseInt(limitStr, 10, 64)
			fmt.Printf("%T %v\n", req.Limit, req.Limit)
			if err != nil {
				log.Error("request query parameter limit is not integer")
				fmt.Printf("%T %v\n", req.Limit, req.Limit)

				render.Status(r, 400)
				render.JSON(w, r, response.Error("Некорректные данные"))
				return
			}
		}
		offsetStr := r.URL.Query().Get("offset")
		if offsetStr != "" {
			req.Offset, err = strconv.ParseInt(offsetStr, 10, 64)
			if err != nil {
				log.Error("request query parameter offset is not integer")
				fmt.Printf("%T %v\n", req.Offset, req.Offset)

				render.Status(r, 400)
				render.JSON(w, r, response.Error("Некорректные данные"))
				return
			}
		}

		var resp Response
		resp.Banners, err = bannerGetterList.GetBannerList(req.FeatureID, req.TagID, req.Limit, req.Offset)
		if err != nil {
			// TODO!!! server error
			fmt.Println(err)
			log.Error("server error")

			render.Status(r, 500)
			render.JSON(w, r, response.Error("Некорректные данные"))
			return
		}

		log.Info("banners found", slog.Any("banners", resp.Banners))
		render.Status(r, 200)
		render.JSON(w, r, resp.Banners)
	}
}
