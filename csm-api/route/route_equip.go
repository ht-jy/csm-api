package route

import (
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func EquipRoute(safeDB *sqlx.DB, r *store.Repository) chi.Router {
	router := chi.NewRouter()

	equipHandler := &handler.HandlerEquip{
		Service: &service.ServiceEquip{
			SafeDB:  safeDB,
			SafeTDB: safeDB,
			Store:   r,
		},
	}

	router.Get("/", equipHandler.List)       // 장비 조회
	router.Get("/all", equipHandler.AllList) // 장비 전체 조회
	router.Post("/", equipHandler.Merge)     // 장비 추가/수정

	return router
}
