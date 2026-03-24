package v1

import "time"

// SysOperLogListQuery 操作日志列表查询参数
type SysOperLogListQuery struct {
	*BasePageQuery
	Title         string `json:"title"`          // 模块标题
	BusinessType  string `json:"business_type"`  // 业务类型
	OperName      string `json:"oper_name"`      // 操作人员
	Status        string `json:"status"`         // 操作状态（0正常 1异常）
	BeginTime     string `json:"begin_time"`     // 开始时间
	EndTime       string `json:"end_time"`       // 结束时间
}

// SysOperLogDetailResponse 操作日志详情响应
type SysOperLogDetailResponse struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	BusinessType  int       `json:"business_type"`
	Method        string    `json:"method"`
	RequestMethod string    `json:"request_method"`
	OperatorType  int       `json:"operator_type"`
	OperName      string    `json:"oper_name"`
	DeptName      string    `json:"dept_name"`
	OperURL       string    `json:"oper_url"`
	OperIP        string    `json:"oper_ip"`
	OperLocation  string    `json:"oper_location"`
	OperParam     string    `json:"oper_param"`
	JSONResult    string    `json:"json_result"`
	Status        int       `json:"status"`
	ErrorMsg      string    `json:"error_msg"`
	OperTime      time.Time `json:"oper_time"`
	CostTime      int64     `json:"cost_time"`
}
