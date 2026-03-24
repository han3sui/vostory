package v1

type SysApiListQuery struct {
	*BasePageQuery
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Tag    string `json:"tag"`
	Perms  string `json:"perms"`
}

type SysApiCreateRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Tag    string `json:"tag"`
}

type SysApiUpdateRequest struct {
	ID     uint   `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Tag    string `json:"tag"`
}
