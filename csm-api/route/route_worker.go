package route

import (
	"csm-api/config"
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func WorkerRoute(safeDB *sqlx.DB, cfg *config.Config, r *store.Repository) chi.Router {
	router := chi.NewRouter()

	workerHandler := handler.HandlerWorker{
		Service: &service.ServiceWorker{
			SafeDB:  safeDB,
			SafeTDB: safeDB,
			Store:   r,
			Config:  cfg,
		},
	}

	router.Get("/total", workerHandler.TotalList)                                   // 전체근로자 조회
	router.Get("/total/absent", workerHandler.AbsentList)                           // 미출근 근로자 검색(현장근로자 추가시 사용)
	router.Get("/total/depart", workerHandler.DepartList)                           // 프로젝트 회사명 조회
	router.Post("/total", workerHandler.Add)                                        // 추가
	router.Put("/total", workerHandler.Modify)                                      // 수정
	router.Post("/total/delete", workerHandler.Remove)                              // 삭제
	router.Get("/site-base", workerHandler.SiteBaseList)                            // 현장근로자 조회
	router.Get("/site-base/co", workerHandler.SiteBaseListByCompany)                // 현장근로자 조회 - 협력업체
	router.Post("/site-base", workerHandler.Merge)                                  // 현장근로자 추가&수정
	router.Post("/site-base/deadline", workerHandler.ModifyDeadline)                // 현장근로자 마감처리
	router.Post("/site-base/project", workerHandler.ModifyProject)                  // 현장근로자 프로젝트 이동
	router.Post("/site-base/delete", workerHandler.SiteBaseRemove)                  // 현장근로자 삭제
	router.Post("/site-base/deadline-cancel", workerHandler.SiteBaseDeadlineCancel) // 마감 취소
	router.Get("/site-base/record", workerHandler.DailyWorkersByJnoAndDate)         // 프로젝트, 기간내 모든 현장근로자 근태정보 조회
	router.Post("/site-base/work-hours", workerHandler.ModifyWorkHours)             // 현장근로자 일괄 공수 변경
	router.Get("/site-base/history", workerHandler.GetDailyWorkerHistory)           // 변경 이력 조회
	router.Get("/site-base/reason", workerHandler.GetDailyWorkerHistoryReason)      // 변경 이력 사유 조회

	return router
}
