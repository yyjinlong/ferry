// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

const (
	DB_QUERY_SERVICE_ERROR  = "查询服务: %s 错误: %s"
	DB_PIPELINE_NOT_FOUND   = "pipeline: %d 查询不存在"
	DB_PIPELINE_QUERY_ERROR = "查询pipeline: %d 失败: %s"
	DB_SERVICE_QUERY_ERROR  = "查询service信息失败: %s"
)

const (
	SVC_WAIT_ALL_SERVICE_ERROR = "等待所有service创建完成失败: %s"
)

const (
	TAG_OPERATE_FORBIDDEN    = "服务被上线单(%s)占用, 不能发布!"
	TAG_WRITE_LOCK_ERROR     = "服务占锁: %s 失败: %s"
	TAG_EXECUTE_SCRIPT_ERROR = "打tag失败: %s"
	TAG_UPDATE_DB_ERROR      = "更新tag信息失败: %s"
)

const (
	IMG_QUERY_PIPELINE_ERROR     = "镜像查询pipelien信息失败: %s"
	IMG_BUILD_FINISHED           = "镜像已操作完, 不能重复操作!"
	IMG_QUERY_UPDATE_ERROR       = "查询镜像变更信息错误: %s"
	IMG_BUILD_PARAM_ENCODE_ERROR = "镜像构建参数json encode失败: %s"
)

const (
	FSH_UPDATE_ONLINE_GROUP_ERROR = "设置当前在线组、部署组失败: %s"
)
