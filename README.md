nautilus
-------------
Jinlong Yang

# nautilus

## 1 why blue-green？

    1 采用滚动更新, maxSurge设置为25%.

        如果pod比较多, 就会进行多次的滚动更新, 发布缓慢同时也会产生新老pod共存, 共同接流量情况, 页面访问出现404, 这是业务不期望的情况.

    2 蓝绿部署的优势

        另一组部署完成后, 流量一刀切.

    3 蓝绿部署的缺点

        瞬间对集群造成一些压力, 但是还好, 发布完成后, 会对旧版本的deployment的pod数量缩成0, 使其不占资源.


## 2 镜像分层

    1) base层    : centos6.7 centos7.5
    2) run层     : python(conda环境)、go(go mod环境)
    3) service层 : 具体的服务环境
    4) release层 : 由imaged进程基于代码自动构建


## 3 节点标签

    kubectl label node x.x.x.x aggregate=default


## 4 业务逻辑

    bll
    ├── image        -- 构建镜像
    ├── listen       -- 监听事件
    ├── pipeline     -- 构建相关流程
    ├── publish      -- 发布
    └── rollback     -- 回滚

    * 新服务上线默认为blue
    * 蓝绿两组分别创建对应的deployment
    * 不再监听endpoint事件, 等待deployment发布完成, 自动将另一组缩成0, 读取endpoint信息并记录


## 5 创建服务

    curl -d 'service=ivr' http://127.0.0.1:8888/v1/service


## 6 发布流程

    1) 创建job

        curl -H 'content-type: application/json' -d '{"name": "ivr test", "summary": "test", "service": "ivr",  "module_list": [{"name": "ivr", "branch": "yy"}], "creator": "yangjinlong", "rd": "yangjinlong", "qa": "yangjinlong", "pm": "yangjinlong"}' http://127.0.0.1:8888/v1/pipeline

    2) 打tag

        curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/tag

    3) 构建镜像

        curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/image

    4) 发布沙盒

        curl -d "pipeline_id=4&phase=sandbox&username=yangjinlong" http://127.0.0.1:8888/v1/deploy | jq .

    5) 发布全量

        curl -d "pipeline_id=4&phase=online&username=yangjinlong" http://127.0.0.1:8888/v1/deploy | jq .

    6) 部署完成

        curl -d "pipeline_id=4" http://127.0.0.1:8888/v1/finish | jq .

