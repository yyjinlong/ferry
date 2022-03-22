package config

// 上线单
const (
	ONLINE_NAME = "上线说明"
	ONLINE_DESC = "上线简介"
	MODULE_NAME = "模块名"
	BRANCH_NAME = "分支名"
)

// 创建pipeline
const (
	PL_SEGMENT_IS_EMPTY      = "字段: %s 内容为空!"
	PL_QUERY_MODULE_ERROR    = "查询模块: %s 信息失败!"
	PL_EXEC_GIT_CHECK_ERROR  = "执行git分支检查失败: %s"
	PL_RESULT_HANDLER_ERROR  = "结果处理失败: %s"
	PL_GIT_CHECK_FAILED      = "git分支检查失败!"
	PL_CREATE_PIPELINE_ERROR = "存储上线流程信息错误: %s"
)

// 打tag
const (
	TAG_OPERATE_FORBIDDEN  = "服务被上线单(%s)占用, 不能发布!"
	TAG_QUERY_UPDATE_ERROR = "查询变更模块信息失败: %s"
	TAG_CREATE_PIPE_ERROR  = "打tag创建输出管道失败: %s"
	TAG_START_EXEC_ERROR   = "打tag执行命令失败: %s"
	TAG_WAIT_FINISH_ERROR  = "等待命令执行完成失败: %s"
	TAG_BUILD_FAILED       = "打tag失败!"
	TAG_UPDATE_DB_ERROR    = "更新tag信息失败: %s"
)

// 数据库查询
const (
	DB_QUERY_SERVICE_ERROR   = "查询服务: %s 错误: %s"
	DB_PIPELINE_NOT_FOUND    = "pipeline: %d 查询不存在"
	DB_PIPELINE_QUERY_ERROR  = "查询pipeline: %d 失败: %s"
	DB_PIPELINE_UPDATE_ERROR = "查询pipeline变更信息失败: %s"
	DB_SERVICE_QUERY_ERROR   = "查询service信息失败: %s"
	DB_QUERY_NAMESPACE_ERROR = "查询命名空间信息失败: %s"
	DB_QUERY_PHASES_ERROR    = "查询pipeline对应阶段错误: %s"
	DB_UPDATE_PIPELINE_ERROR = "更新pipeline状态失败: %s"
	DB_WRITE_LOCK_ERROR      = "服务占锁: %v 失败: %s"
)

const (
	SVC_BUILD_SERVICE_YAML_ERROR = "创建service yaml失败: %s"
	SVC_K8S_SERVICE_EXEC_FAILED  = "K8S创建service失败: %s"
	SVC_WAIT_ALL_SERVICE_ERROR   = "等待所有service创建完成失败: %s"
)

// 构建镜像
const (
	IMG_QUERY_PIPELINE_ERROR     = "镜像查询pipelien信息失败: %s"
	IMG_BUILD_FINISHED           = "镜像已操作完, 不能重复操作!"
	IMG_QUERY_UPDATE_ERROR       = "查询镜像变更信息错误: %s"
	IMG_BUILD_PARAM_ENCODE_ERROR = "镜像构建参数json encode失败: %s"
	IMG_SEND_BUILD_TO_MQ_FAILED  = "发送镜像构建信息到MQ失败: %s"
	IMG_QUERY_IS_BUILD_ERROR     = "查询镜像是否构建失败: %s"
	IMG_QUERY_IMAGE_IS_BUILED    = "查询镜像信息已构建!"
	IMG_CREATE_IMAGE_INFO_ERROR  = "写镜像信息到数据库失败: %s"
)

const (
	PUB_DEPLOY_FINISHED               = "服务已部署完成, 不能重复操作!"
	PUB_BUILD_DEPLOYMENT_YAML_ERROR   = "创建deployment yaml失败: %s"
	PUB_K8S_DEPLOYMENT_EXEC_FAILED    = "K8S创建deployment失败: %s"
	PUB_RECORD_DEPLOYMENT_TO_DB_ERROR = "写deployment信息到数据库失败: %s"
)

// 确认完成
const (
	FSH_UPDATE_ONLINE_GROUP_ERROR = "设置当前在线组、部署组失败: %s"
)

// 回滚
const (
	ROL_CANNOT_EXECUTE     = "不能执行回滚"
	ROL_PROCESS_NO_EXECUTE = "发布中的deployment不能回滚!"
	ROL_RECORD_PHASE_ERROR = "记录回滚阶段: %s 错误: %s"
)
