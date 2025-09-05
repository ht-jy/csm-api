package store

import (
	"context"
	"csm-api/entity"
	"database/sql"
	"github.com/guregu/null"
	"time"
)

type MenuStore interface {
	GetParentMenu(ctx context.Context, db Queryer, roles []string) ([]entity.Menu, error)
	GetChildMenu(ctx context.Context, db Queryer, roles []string) ([]entity.Menu, error)
}

type GetUserValidStore interface {
	GetUserInfo(ctx context.Context, db Queryer, userId string) (entity.User, error)
	GetUserValid(ctx context.Context, db Queryer, userId string, userPwd string) (entity.User, error)
	GetCompanyUserValid(ctx context.Context, db Queryer, userId string, userPwd string) (entity.CompanyInfo, error)
	GetCompanyUser(ctx context.Context, db Queryer, userId string) (entity.CompanyInfo, error)
}

type SiteStore interface {
	GetSiteList(ctx context.Context, db Queryer, targetDate time.Time, role int, uno int64, status string) (*entity.Sites, error)
	GetSiteNmList(ctx context.Context, db Queryer, page entity.PageSql, search entity.Site, nonSite int) (*entity.Sites, error)
	GetSiteNmCount(ctx context.Context, db Queryer, search entity.Site, nonSite int) (int, error)
	GetSiteStatsList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.Sites, error)
	ModifySite(ctx context.Context, tx Execer, site entity.Site) error
	DeleteSite(ctx context.Context, tx Execer, sno int64, user entity.User) error
	AddSite(ctx context.Context, db Queryer, tx Execer, jno int64, user entity.User) error
	ModifySiteIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	ModifySiteIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	SettingWorkRate(ctx context.Context, tx Execer, targetDate time.Time) (int64, error)
	ModifyWorkRate(ctx context.Context, tx Execer, workRate entity.SiteWorkRate) error
	GetSiteWorkRateByDate(ctx context.Context, db Queryer, jno int64, month string) (entity.SiteWorkRate, error)
	GetSiteWorkRateListByMonth(ctx context.Context, db Queryer, jno int64, searchDate string) (entity.SiteWorkRates, error)
	AddWorkRate(ctx context.Context, tx Execer, workRate entity.SiteWorkRate) error
}

type SitePosStore interface {
	GetSitePosList(ctx context.Context, db Queryer) ([]entity.SitePos, error)
	GetSitePosData(ctx context.Context, db Queryer, sno int64) (*entity.SitePos, error)
	ModifySitePosData(ctx context.Context, tx Execer, sno int64, sitePosSql entity.SitePos) error
	ModifySitePosIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	ModifySitePosIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error
}

type SiteDateStore interface {
	GetSiteDateData(ctx context.Context, db Queryer, sno int64) (*entity.SiteDate, error)
	ModifySiteDate(ctx context.Context, tx Execer, sno int64, siteDateSql entity.SiteDate) error
	ModifySiteDateIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	ModifySiteDateIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error
}

type ProjectStore interface {
	GetProjectList(ctx context.Context, db Queryer, sno int64, targetDate time.Time) (*entity.ProjectInfos, error)
	GetProjectWorkerCountList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.ProjectInfos, error)
	GetProjectSafeWorkerCountList(ctx context.Context, db Queryer, targetDate time.Time) (*entity.ProjectSafeCounts, error)
	GetProjectNmList(ctx context.Context, db Queryer, role int, uno int64) (*entity.ProjectInfos, error)
	GetUsedProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, search entity.JobInfo, retry string, includeJno string, snoString string) (*entity.JobInfos, error)
	GetUsedProjectCount(ctx context.Context, db Queryer, search entity.JobInfo, retry string, includeJno string, snoString string) (int, error)
	GetAllProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, search entity.JobInfo, isAll int, retry string) (*entity.JobInfos, error)
	GetAllProjectCount(ctx context.Context, db Queryer, search entity.JobInfo, retry string) (int, error)
	GetStaffProjectList(ctx context.Context, db Queryer, pageSql entity.PageSql, searchSql entity.JobInfo, uno sql.NullInt64, retry string) (*entity.JobInfos, error)
	GetStaffProjectCount(ctx context.Context, db Queryer, searchSql entity.JobInfo, uno sql.NullInt64, retry string) (int, error)
	GetProjectNmUnoList(ctx context.Context, db Queryer, uno sql.NullInt64, role int) (*entity.ProjectInfos, error)
	GetNonUsedProjectList(ctx context.Context, db Queryer, page entity.PageSql, search entity.NonUsedProject, retry string) (*entity.NonUsedProjects, error)
	GetNonUsedProjectCount(ctx context.Context, db Queryer, search entity.NonUsedProject, retry string) (int, error)
	GetNonUsedProjectListByType(ctx context.Context, db Queryer, page entity.PageSql, search entity.NonUsedProject, retry string, typeString string) (*entity.NonUsedProjects, error)
	GetNonUsedProjectCountByType(ctx context.Context, db Queryer, search entity.NonUsedProject, retry string, typeString string) (int, error)
	GetProjectBySite(ctx context.Context, db Queryer, sno int64) (entity.ProjectInfos, error)
	AddProject(ctx context.Context, tx Execer, project entity.ReqProject) error
	ModifyDefaultProject(ctx context.Context, tx Execer, project entity.ReqProject) error
	ModifyUseProject(ctx context.Context, tx Execer, project entity.ReqProject) error
	RemoveProject(ctx context.Context, tx Execer, sno int64, jno int64) error
	ModifyProjectIsNonUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	ModifyProjectIsUse(ctx context.Context, tx Execer, site entity.ReqSite) error
	ModifyProject(ctx context.Context, tx Execer, project entity.ReqProject) error
}
type ProjectSettingStore interface {
	GetManHourList(ctx context.Context, db Queryer, jno int64) (*entity.ManHours, error)
	MergeManHour(ctx context.Context, tx Execer, manHour entity.ManHour) (int64, error)
	AddManHour(ctx context.Context, tx Execer, manHour entity.ManHour) error
	MergeProjectSetting(ctx context.Context, tx Execer, project entity.ProjectSetting) (int64, error)
	GetCheckProjectSetting(ctx context.Context, db Queryer) (*entity.ProjectSettings, error)
	GetCheckProjectManHours(ctx context.Context, db Queryer) (*entity.ProjectSettings, error)
	GetProjectSetting(ctx context.Context, db Queryer, jno int64) (*entity.ProjectSettings, error)
	DeleteManHour(ctx context.Context, tx Execer, mhno int64) error
	ProjectSettingLog(ctx context.Context, tx Execer, setting entity.ProjectSetting) error
	ManHourLog(ctx context.Context, tx Execer, manhour entity.ManHour) error
}
type OrganizationStore interface {
	GetFuncNameList(ctx context.Context, db Queryer) (*entity.FuncNameSqls, error)
	GetOrganizationClientList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.OrganizationSqls, error)
	GetOrganizationHtencList(ctx context.Context, db Queryer, jno sql.NullInt64, funcNo sql.NullInt64) (*entity.OrganizationSqls, error)
}

type ProjectDailyStore interface {
	GetProjectDailyContentList(ctx context.Context, db Queryer, jno int64, targetDate time.Time) (*entity.ProjectDailys, error)
	GetDailyJobList(ctx context.Context, db Queryer, jno int64, targetDate string) (entity.ProjectDailys, error)
	AddDailyJob(ctx context.Context, tx Execer, project entity.ProjectDailys) error
	ModifyDailyJob(ctx context.Context, tx Execer, project entity.ProjectDaily) error
	RemoveDailyJob(ctx context.Context, tx Execer, idx int64) error
}

type UserStore interface {
	GetUserInfoPeList(ctx context.Context, db Queryer, unoList []int) (*entity.UserPeInfos, error)
	GetSiteRole(ctx context.Context, db Queryer, jno int64, uno int64) (string, error)
	GetOperationalRole(ctx context.Context, db Queryer, jno int64, uno int64) (string, error)
	GetAuthorizationList(ctx context.Context, db Queryer, api string) (*entity.RoleList, error)
	GetSupervisorRole(ctx context.Context, db Queryer, uno int64) (string, error)
}

type CodeStore interface {
	GetCodeList(ctx context.Context, db Queryer, pCode string) (*entity.Codes, error)
	GetCodeTree(ctx context.Context, db Queryer, pCode string) (*entity.Codes, error)
	MergeCode(ctx context.Context, tx Execer, code entity.Code) error
	RemoveCode(ctx context.Context, tx Execer, idx int64) error
	ModifySortNo(ctx context.Context, tx Execer, codeSort entity.CodeSort) error
	DuplicateCheckCode(ctx context.Context, db Queryer, code string) (int, error)
}

type NoticeStore interface {
	GetNoticeList(ctx context.Context, db Queryer, uno null.Int, role int, pageSql entity.PageSql, search entity.Notice) (*entity.Notices, error)
	GetNoticeListCount(ctx context.Context, db Queryer, uno null.Int, role int, search entity.Notice) (int, error)
	AddNotice(ctx context.Context, tx Execer, notice entity.Notice) error
	ModifyNotice(ctx context.Context, tx Execer, notice entity.Notice) error
	RemoveNotice(ctx context.Context, tx Execer, idx int64) error
}

type DeviceStore interface {
	GetDeviceList(ctx context.Context, db Queryer, page entity.PageSql, search entity.Device, retry string) (*entity.Devices, error)
	GetDeviceListCount(ctx context.Context, db Queryer, search entity.Device, retry string) (int, error)
	AddDevice(ctx context.Context, tx Execer, device entity.Device) error
	ModifyDevice(ctx context.Context, tx Execer, device entity.Device) error
	RemoveDevice(ctx context.Context, tx Execer, dno sql.NullInt64) error
	GetDeviceLog(ctx context.Context, db Queryer) (*entity.RecdLogOrigins, error)
	GetCheckRegistered(ctx context.Context, db Queryer, deviceName string) (int, error)
}

type WorkerStore interface {
	GetWorkerTotalList(ctx context.Context, db Queryer, page entity.PageSql, isRole bool, uno string, search entity.Worker, retry string) (*entity.Workers, error)
	GetWorkerTotalCount(ctx context.Context, db Queryer, isRole bool, uno string, search entity.Worker, retry string) (int, error)
	GetAbsentWorkerList(ctx context.Context, db Queryer, page entity.PageSql, search entity.WorkerDaily, retry string) (*entity.Workers, error)
	GetAbsentWorkerCount(ctx context.Context, db Queryer, search entity.WorkerDaily, retry string) (int, error)
	GetWorkerDepartList(ctx context.Context, db Queryer, jno int64) ([]string, error)
	AddWorker(ctx context.Context, tx Execer, worker entity.Worker) (int64, error)
	MergeWorker(ctx context.Context, tx Execer, worker entity.Worker) (int64, error)
	ModifyWorker(ctx context.Context, tx Execer, worker entity.Worker) error
	RemoveWorker(ctx context.Context, tx Execer, worker entity.Worker) error
	GetWorkerSiteBaseList(ctx context.Context, db Queryer, page entity.PageSql, isRole bool, uno string, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error)
	GetWorkerSiteBaseCount(ctx context.Context, db Queryer, isRole bool, uno string, search entity.WorkerDaily, retry string) (int, error)
	GetWorkerSiteBaseListByCompany(ctx context.Context, db Queryer, page entity.PageSql, id string, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error)
	GetWorkerSiteBaseByCompanyCount(ctx context.Context, db Queryer, id string, search entity.WorkerDaily, retry string) (int, error)
	MergeSiteBaseWorker(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	MergeSiteBaseWorkerLog(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyWorkerDeadline(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyWorkerProject(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyWorkerDefaultProject(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyWorkerDeadlineInit(ctx context.Context, tx Execer) error
	GetWorkerOverTime(ctx context.Context, db Queryer) (*entity.WorkerOverTimes, error)
	ModifyWorkerOverTime(ctx context.Context, tx Execer, workerOverTime entity.WorkerOverTime) error
	DeleteWorkerOverTime(ctx context.Context, tx Execer, cno null.Int) error
	RemoveSiteBaseWorkers(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyDeadlineCancel(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	GetDailyWorkerUserKey(ctx context.Context, db Queryer, worker entity.WorkerDaily) (string, error)
	AddDailyWorkers(ctx context.Context, db Queryer, tx Execer, workers entity.WorkerDailys) (entity.WorkerDailys, error)
	GetDailyWorkersByJnoAndDate(ctx context.Context, db Queryer, param entity.RecordDailyWorkerReq) ([]entity.RecordDailyWorkerRes, error)
	ModifyWorkHours(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	GetRecdWorkerList(ctx context.Context, db Queryer) ([]entity.Worker, error)
	GetRecdWorkerUserKey(ctx context.Context, db Queryer, worker entity.Worker) (string, error)
	MergeRecdWorker(ctx context.Context, tx Execer, worker []entity.Worker) error
	GetRecdDailyWorkerList(ctx context.Context, db Queryer) ([]entity.WorkerDaily, error)
	GetRecdDailyWorkerChk(ctx context.Context, db Queryer, userKey string, date null.Time) (bool, error)
	MergeRecdDailyWorker(ctx context.Context, tx Execer, worker []entity.WorkerDaily) error
	GetDailyWorkerBeforeList(ctx context.Context, db Queryer, workers entity.WorkerDailys) (entity.WorkerDailys, error)
	AddHistoryDailyWorkers(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	GetHistoryDailyWorkers(ctx context.Context, db Queryer, startDate string, endDate string, sno int64, retry string, userKeys []string) (entity.WorkerDailys, error)
	GetHistoryDailyWorkerReason(ctx context.Context, db Queryer, cno int64) (string, error)
}

type WorkHourStore interface {
	ModifyWorkHour(ctx context.Context, tx Execer, user entity.Base) error
	ModifyWorkHourByJno(ctx context.Context, tx Execer, jno int64, user entity.Base, uuids []string) error
}

type CompanyStore interface {
	GetJobInfo(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.JobInfo, error)
	GetSiteManagerList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.Managers, error)
	GetSafeManagerList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.Managers, error)
	GetSupervisorList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.Supervisors, error)
	GetConstruction(ctx context.Context, db Queryer, jno int64) (*entity.Supervisors, error)
	GetWorkInfoList(ctx context.Context, db Queryer) (*entity.WorkInfos, error)
	GetCompanyInfoList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.CompanyInfos, error)
	GetCompanyWorkInfoList(ctx context.Context, db Queryer, jno sql.NullInt64) (*entity.WorkInfos, error)
}

type EquipStore interface {
	GetEquipList(ctx context.Context, db Queryer, jno int64, sno int64) (entity.Equips, error)
	GetEquip(ctx context.Context, db Queryer, jno int64, sno int64, recordDate time.Time) (entity.Equips, error)
	MergeEquipCnt(ctx context.Context, tx Execer, equip entity.Equip) error
}

type ScheduleStore interface {
	GetRestScheduleList(ctx context.Context, db Queryer, jno int64, year string, month string) (entity.RestSchedules, error)
	AddRestSchedule(ctx context.Context, tx Execer, schedule entity.RestSchedules) error
	ModifyRestSchedule(ctx context.Context, tx Execer, schedule entity.RestSchedule) error
	RemoveRestSchedule(ctx context.Context, tx Execer, cno int64) error
}

type UploadFileStore interface {
	GetUploadRound(ctx context.Context, db Queryer, file entity.UploadFile) (int, error)
	GetUploadFileList(ctx context.Context, db Queryer, file entity.UploadFile) ([]entity.UploadFile, error)
	GetUploadFile(ctx context.Context, db Queryer, file entity.UploadFile) (entity.UploadFile, error)
	AddUploadFile(ctx context.Context, tx Execer, file entity.UploadFile) error
}

type CompareStore interface {
	GetDailyWorkerList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) (entity.WorkerDailys, error)
	GetTbmList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) ([]entity.Tbm, error)
	GetDeductionList(ctx context.Context, db Queryer, compare entity.Compare, retry string, order string) ([]entity.Deduction, error)
	ModifyWorkerCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyDailyWorkerCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyTbmCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	ModifyDeductionCompareApply(ctx context.Context, tx Execer, workers entity.WorkerDailys) error
	AddCompareLog(ctx context.Context, tx Execer, logs entity.WorkerDailys) error
}

type ExcelStore interface {
	GetTbmOrder(ctx context.Context, db Queryer, tbm entity.Tbm) (string, error)
	AddTbmExcel(ctx context.Context, tx Execer, tbm []entity.Tbm) error
	GetDeductionJobNameByJno(ctx context.Context, db Queryer, jno int64) (string, error)
	GetDeductionOrder(ctx context.Context, db Queryer, tbm entity.Deduction) (string, error)
	AddDeductionExcel(ctx context.Context, tx Execer, tbm []entity.Deduction) error
}

type WeatherStore interface {
	SaveWeather(ctx context.Context, tx Execer, weather entity.Weather) error
	GetWeatherList(ctx context.Context, db Queryer, sno int64, targetDate time.Time) (*entity.Weathers, error)
}

type UserRoleStore interface {
	GetUserRoleListByUno(ctx context.Context, db Queryer, uno int64) ([]entity.UserRoleMap, error)
	GetUserRoleListByCodeAndJno(ctx context.Context, db Queryer, code string, jno int64) ([]entity.UserRoleMap, error)
	AddUserRole(ctx context.Context, tx Execer, userRoles []entity.UserRoleMap) error
	RemoveUserRole(ctx context.Context, tx Execer, userRoles []entity.UserRoleMap) error
	GetUserMenuRoleCheck(ctx context.Context, db Queryer, role string, menuId string) (bool, error)
}
