package service

import (
	"context"
	"csm-api/entity"
	"time"
)

type MenuService interface {
	GetMenu(ctx context.Context, roles []string) (entity.MenuRes, error)
}

type GetUserValidService interface {
	GetUserValid(ctx context.Context, userId string, userPwd string, isAdmin bool) (entity.User, error)
	GetCompanyUserValid(ctx context.Context, userId string, userPwd string, isAdmin bool) (entity.User, error)
}

type SiteService interface {
	GetSiteList(ctx context.Context, targetDate time.Time, isRole bool) (*entity.Sites, error)
	GetSiteNmList(ctx context.Context, page entity.Page, search entity.Site, nonSite int) (*entity.Sites, error)
	GetSiteNmCount(ctx context.Context, search entity.Site, nonSite int) (int, error)
	GetSiteStatsList(ctx context.Context, targetDate time.Time) (*entity.Sites, error)
	ModifySite(ctx context.Context, site entity.Site) error
	AddSite(ctx context.Context, jno int64, user entity.User) error
	ModifySiteIsNonUse(ctx context.Context, site entity.ReqSite) error
	ModifySiteIsUse(ctx context.Context, site entity.ReqSite) error
	ModifySiteJobNonUse(ctx context.Context, site entity.ReqSite) error
	ModifySiteJobUse(ctx context.Context, site entity.ReqSite) error
	SettingWorkRate(ctx context.Context, targetDate time.Time) (int64, error)
	ModifyWorkRate(ctx context.Context, workRate entity.SiteWorkRate) error
	GetSiteWorkRateByDate(ctx context.Context, jno int64, month string) (entity.SiteWorkRate, error)
	GetSiteWorkRateListByMonth(ctx context.Context, jno int64, month string) (entity.SiteWorkRates, error)
	AddWorkRate(ctx context.Context, workRate entity.SiteWorkRate) error
}

type SitePosService interface {
	GetSitePosList(ctx context.Context) ([]entity.SitePos, error)
	GetSitePosData(ctx context.Context, sno int64) (*entity.SitePos, error)
	ModifySitePos(ctx context.Context, sno int64, sitePos entity.SitePos) error
}

type SiteDateService interface {
	GetSiteDateData(ctx context.Context, sno int64) (*entity.SiteDate, error)
	ModifySiteDate(ctx context.Context, sno int64, siteDate entity.SiteDate) error
}

type ProjectService interface {
	GetProjectList(ctx context.Context, sno int64, targetDate time.Time) (*entity.ProjectInfos, error)
	GetProjectWorkerCountList(ctx context.Context, targetDate time.Time) (*entity.ProjectInfos, error)
	GetProjectNmList(ctx context.Context, isRole bool) (*entity.ProjectInfos, error)
	GetUsedProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, retry string, includeJno string, snoString string) (*entity.JobInfos, error)
	GetUsedProjectCount(ctx context.Context, search entity.JobInfo, retry string, includeJno string, snoString string) (int, error)
	GetAllProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, isAll int, retry string) (*entity.JobInfos, error)
	GetAllProjectCount(ctx context.Context, search entity.JobInfo, isAll int, retry string) (int, error)
	GetStaffProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, uno int64, retry string) (*entity.JobInfos, error)
	GetStaffProjectCount(ctx context.Context, search entity.JobInfo, uno int64, retry string) (int, error)
	GetProjectNmUnoList(ctx context.Context, uno int64, role string) (*entity.ProjectInfos, error)
	GetNonUsedProjectList(ctx context.Context, page entity.Page, search entity.NonUsedProject, retry string) (*entity.NonUsedProjects, error)
	GetNonUsedProjectCount(ctx context.Context, search entity.NonUsedProject, retry string) (int, error)
	GetNonUsedProjectListByType(ctx context.Context, page entity.Page, search entity.NonUsedProject, retry string, typeString string) (*entity.NonUsedProjects, error)
	GetNonUsedProjectCountByType(ctx context.Context, search entity.NonUsedProject, retry string, typeString string) (int, error)
	GetProjectBySite(ctx context.Context, sno int64) (entity.ProjectInfos, error)
	AddProject(ctx context.Context, project entity.ReqProject) error
	ModifyDefaultProject(ctx context.Context, project entity.ReqProject) error
	ModifyUseProject(ctx context.Context, project entity.ReqProject) error
	RemoveProject(ctx context.Context, sno int64, jno int64) error
}

type ProjectSettingService interface {
	GetProjectSetting(ctx context.Context, jno int64) (*entity.ProjectSettings, error)
	GetManHourList(ctx context.Context, jno int64) (*entity.ManHours, error)
	MergeManHours(ctx context.Context, manHours *entity.ManHours) error
	MergeProjectSetting(ctx context.Context, project entity.ProjectSetting) error
	CheckProjectSetting(ctx context.Context) (count int, err error)
	DeleteManHour(ctx context.Context, mhno int64, manhour entity.ManHour) error
	AddManHour(ctx context.Context, manhour entity.ManHour) error
}

type OrganizationService interface {
	GetOrganizationClientList(ctx context.Context, jno int64) (*entity.OrganizationPartitions, error)
	GetOrganizationHtencList(ctx context.Context, jno int64) (*entity.OrganizationPartitions, error)
}

type ProjectDailyService interface {
	GetDailyJobList(ctx context.Context, jno int64, targetDate string) (entity.ProjectDailys, error)
	AddDailyJob(ctx context.Context, project entity.ProjectDailys) error
	ModifyDailyJob(ctx context.Context, project entity.ProjectDaily) error
	RemoveDailyJob(ctx context.Context, idx int64) error
}

type UserService interface {
	GetUserInfoPeList(ctx context.Context, unoList []int) (*entity.UserPeInfos, error)
	GetUserRole(ctx context.Context, jno int64, uno int64) (string, error)
	GetAuthorizationList(ctx context.Context, api string) (*entity.RoleList, error)
}

type CodeService interface {
	GetCodeList(ctx context.Context, pCode string) (*entity.Codes, error)
	GetCodeTree(ctx context.Context, pCode string) (*entity.CodeTrees, error)
	MergeCode(ctx context.Context, code entity.Code) error
	RemoveCode(ctx context.Context, idx int64) error
	ModifySortNo(ctx context.Context, codeSorts entity.CodeSorts) error
	DuplicateCheckCode(ctx context.Context, code string) (bool, error)
}

type NoticeService interface {
	GetNoticeList(ctx context.Context, page entity.Page, isRole bool, search entity.Notice) (*entity.Notices, error)
	GetNoticeListCount(ctx context.Context, isRole bool, search entity.Notice) (int, error)
	AddNotice(ctx context.Context, notice entity.Notice) error
	ModifyNotice(ctx context.Context, notice entity.Notice) error
	RemoveNotice(ctx context.Context, idx int64) error
}

type DeviceService interface {
	GetDeviceList(ctx context.Context, page entity.Page, search entity.Device, retry string) (*entity.Devices, error)
	GetDeviceListCount(ctx context.Context, search entity.Device, retry string) (int, error)
	AddDevice(ctx context.Context, device entity.Device) error
	ModifyDevice(ctx context.Context, device entity.Device) error
	RemoveDevice(ctx context.Context, dno int64) error
	GetCheckRegisteredDevices(ctx context.Context) ([]string, error)
}

type WorkerService interface {
	GetWorkerTotalList(ctx context.Context, page entity.Page, search entity.Worker, retry string) (*entity.Workers, error)
	GetWorkerTotalCount(ctx context.Context, search entity.Worker, retry string) (int, error)
	GetAbsentWorkerList(ctx context.Context, page entity.Page, search entity.WorkerDaily, retry string) (*entity.Workers, error)
	GetAbsentWorkerCount(ctx context.Context, search entity.WorkerDaily, retry string) (int, error)
	GetWorkerDepartList(ctx context.Context, jno int64) ([]string, error)
	AddWorker(ctx context.Context, worker entity.Worker) error
	ModifyWorker(ctx context.Context, worker entity.Worker) error
	RemoveWorker(ctx context.Context, worker entity.Worker) error
	GetWorkerSiteBaseList(ctx context.Context, page entity.Page, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error)
	GetWorkerSiteBaseCount(ctx context.Context, search entity.WorkerDaily, retry string) (int, error)
	MergeSiteBaseWorker(ctx context.Context, workers entity.WorkerDailys) error
	ModifyWorkerDeadline(ctx context.Context, workers entity.WorkerDailys) error
	ModifyWorkerProject(ctx context.Context, workers entity.WorkerDailys) error
	ModifyWorkerDeadlineInit(ctx context.Context) error
	ModifyWorkerOverTime(ctx context.Context) (int, error)
	RemoveSiteBaseWorkers(ctx context.Context, workers entity.WorkerDailys) error
	ModifyDeadlineCancel(ctx context.Context, workers entity.WorkerDailys) error
	GetDailyWorkersByJnoAndDate(ctx context.Context, param entity.RecordDailyWorkerReq) ([]entity.RecordDailyWorkerRes, error)
	ModifyWorkHours(ctx context.Context, workers entity.WorkerDailys) error
	MergeRecdWorker(ctx context.Context) error
	MergeRecdDailyWorker(ctx context.Context) error
	GetHistoryDailyWorkers(ctx context.Context, startDate string, endDate string, sno int64, retry string, userKeys []string) (entity.WorkerDailys, error)
	GetHistoryDailyWorkerReason(ctx context.Context, cno int64) (string, error)
}

type WorkHourService interface {
	ModifyWorkHour(ctx context.Context, user entity.Base) error
	ModifyWorkHourByJno(ctx context.Context, jno int64, user entity.Base, uuids []string) error
}

type CompanyService interface {
	GetJobInfo(ctx context.Context, jno int64) (*entity.JobInfo, error)
	GetSiteManagerList(ctx context.Context, jno int64) (*entity.Managers, error)
	GetSafeManagerList(ctx context.Context, jno int64) (*entity.Managers, error)
	GetSupervisorList(ctx context.Context, jno int64) (*entity.Supervisors, error)
	GetWorkInfoList(ctx context.Context) (*entity.WorkInfos, error)
	GetCompanyInfoList(ctx context.Context, jno int64) (*entity.CompanyInfoResList, error)
}

type WeatherApiService interface {
	GetWeatherSrtNcst(date string, time string, nx int, ny int) (entity.WeatherSrtEntityRes, error)
	GetWeatherWrnMsg() (entity.WeatherWrnMsgList, error)
	SaveWeather(ctx context.Context) error
	GetWeatherList(ctx context.Context, sno int64, targetDate time.Time) (*entity.Weathers, error)
}
type AddressSearchAPIService interface {
	GetAPILatitudeLongtitude(roadAddress string) (*entity.Point, error)
	GetAPISiteMapPoint(roadAddress string) (*entity.MapPoint, error)
}

type RestDateApiService interface {
	GetRestDelDates(year string, month string) (entity.RestDates, error)
}

type EquipService interface {
	GetEquipList(ctx context.Context) (entity.EquipTemps, error)
	MergeEquipCnt(ctx context.Context, equips entity.EquipTemps) error
}

type ScheduleService interface {
	GetRestScheduleList(ctx context.Context, jno int64, year string, month string) (entity.RestSchedules, error)
	AddRestSchedule(ctx context.Context, schedule entity.RestSchedules) error
	ModifyRestSchedule(ctx context.Context, schedule entity.RestSchedule) error
	RemoveRestSchedule(ctx context.Context, cno int64) error
}

type ExcelService interface {
	ImportTbm(ctx context.Context, path string, tbm entity.Tbm, file entity.UploadFile) error
	ImportDeduction(ctx context.Context, path string, deduction entity.Deduction, file entity.UploadFile) error
	ImportAddDailyWorker(ctx context.Context, path string, worker entity.WorkerDaily) (entity.WorkerDailys, error)
	ImportAddWorker(ctx context.Context, path string, worker entity.Worker) (entity.Workers, error)
}

type UploadFileService interface {
	GetUploadFileList(ctx context.Context, file entity.UploadFile) ([]entity.UploadFile, error)
	GetUploadFile(ctx context.Context, file entity.UploadFile) (entity.UploadFile, error)
}

type CompareService interface {
	GetCompareList(ctx context.Context, compare entity.Compare, retry string, order string) ([]entity.Compare, error)
	ModifyWorkerCompareApply(ctx context.Context, workers entity.WorkerDailys) error
}

type UserRoleService interface {
	GetUserRoleListByUno(ctx context.Context, uno int64) ([]entity.UserRoleMap, error)
	GetUserRoleListByCodeAndJno(ctx context.Context, code string, jno int64) ([]entity.UserRoleMap, error)
	AddUserRole(ctx context.Context, userRoles []entity.UserRoleMap) error
	RemoveUserRole(ctx context.Context, userRoles []entity.UserRoleMap) error
	GetUserMenuRoleCheck(ctx context.Context, role string, menuId string) (bool, error)
}
