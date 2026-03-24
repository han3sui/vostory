package v1

type SysLogininforQueryParams struct {
	*BasePageQuery
	LoginName string
	IPAddr    string
	Status    string
	StartTime string
	EndTime   string
}
