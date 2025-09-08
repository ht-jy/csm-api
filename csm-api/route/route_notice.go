package route

import (
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func NoticeRoute(safeDB *sqlx.DB, r *store.Repository) chi.Router {
	router := chi.NewRouter()

	noticeHandler := &handler.NoticeHandler{
		Service: &service.ServiceNotice{
			SafeDB:  safeDB,
			SafeTDB: safeDB,
			Store:   r,
		},
	}

	router.Get("/", noticeHandler.List)           // 조회
	router.Post("/", noticeHandler.Add)           // 추가
	router.Put("/", noticeHandler.Modify)         // 수정
	router.Delete("/{idx}", noticeHandler.Remove) // 삭제

	return router
}
