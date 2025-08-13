package route

import (
	"csm-api/handler"
	"csm-api/service"
	"csm-api/store"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func ExcelRoute(safeDB *sqlx.DB, r *store.Repository) chi.Router {
	router := chi.NewRouter()

	excelHandler := &handler.HandlerExcel{
		Service: &service.ServiceExcel{
			SafeDB:      safeDB,
			SafeTDB:     safeDB,
			Store:       r,
			WorkerStore: r,
			FileStore:   r,
		},
		FileService: &service.ServiceUploadFile{
			DB:    safeDB,
			TDB:   safeDB,
			Store: r,
		},
		DB: safeDB,
	}

	router.Post("/import", excelHandler.ImportExcel)                                      // excel import
	router.Get("/export", excelHandler.UploadExportExcel)                                 // upload excel export
	router.Get("/daily-worker/form/export", excelHandler.DailyWorkerFormExport)           // 현장근로자 양식 다운로드
	router.Get("/total-worker/form/export", excelHandler.TotalWorkerFormExport)           // 현장근로자 양식 다운로드
	router.Post("/daily-worker/record/export", excelHandler.DailyWorkerRecordExcelExport) // 근로자 근태기록 export
	router.Get("/download", excelHandler.DownloadFormExcel)                               // 양식 다운로드
	return router
}
