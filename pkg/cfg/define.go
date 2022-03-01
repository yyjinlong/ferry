package cfg

// 创建pipeline

const (
	DB_CREATE_PIPELINE_ERROR = "存储上线流程信息错误: %s"
)

const (
	ONLINE_NAME = "上线说明"
	ONLINE_DESC = "上线简介"
	MODULE_NAME = "模块名"
	BRANCH_NAME = "分支名"
)

const (
	PL_SEGMENT_IS_EMPTY     = "字段: %s 内容为空!"
	PL_QUERY_MODULE_ERROR   = "查询模块: %s 信息失败!"
	PL_EXEC_GIT_CHECK_ERROR = "执行git分支检查失败: %s"
	PL_RESULT_HANDLER_ERROR = "结果处理失败: %s"
	PL_GIT_CHECK_FAILED     = "git分支检查失败!"
)
