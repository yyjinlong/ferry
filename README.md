nautilus
-------------
Jinlong Yang

# nautilus

## 1 why blue-green？

### 1.1 采用滚动更新, maxSurge设置为25%.

    如果pod比较多, 就会进行多次的滚动更新, 发布缓慢同时也会产生新老pod共存, 共同接流量情况, 页面访问出现404, 这是业务不期望的情况.

### 1.2 蓝绿部署的优势

    另一组部署完成后, 流量一刀切.

### 1.3 蓝绿部署的缺点

    瞬间对集群造成一些压力, 但是还好, 发布完成后, 会对旧版本的deployment的pod数量缩成0, 使其不占资源.


## 2 镜像分层

    base层    : 操作系统层: centos6.7 centos7.5
    run层     : 运行时环境: python(conda环境)、java(tomcat环境)
    service层 : 具体的服务
    code层    : 基于代码构建镜像(一个服务对应多个代码模块)

## 2.1 为什么不用release层, 而用code层

    (1) 镜像太大的问题: 如果直接继承service层, 将代码拷贝到service层, 构建release镜像, 镜像就会很大.
    (2) 服务对应多个代码模块, 如果只变更一个模块, 其他模块也需要跟着再进行编译, 无法复用.


## 3 业务逻辑

    ├── pipeline  -- 构建
    └── publish   -- 发布

    * 新服务上线默认为blue
    * 蓝绿两组分别创建对应的deployment、service
    * 部署确认完成时, 需要将另一组缩成0
    * endpoint事件中, 所有pod都ready后, 再获取对应的ip信息
    * 流量接入nginx, 在发布时会存在蓝绿pod同时在线, 所以endpoint事件中需要控制当前接入流量的pod


## 4 依赖准备

### 4.1 节点标签

```
# 业务node
kubectl label node xxxx aggregate=default

# 定时node
kubectl label node xxxx aggregate=cronjob

# 查看所有node label
kubectl get nodes --show-labels
```

### 4.2 imagePullSecrets

```
# 登录私有仓库
docker login 10.12.28.4:80

# 支持http方式拉取
vim /usr/lib/systemd/system/docker.service
...
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock --insecure-registry 10.12.28.4:80

# 创建imagePullSecrets
kubectl create secret generic harborkey --from-file=.dockerconfigjson=/root/.docker/config.json --type=kubernetes.io/dockerconfigjson
```

## 5 创建服务

```
curl -d 'service=ivr' http://127.0.0.1:8888/v1/deploy/service
```


## 6 创建configmap

```
curl -d 'namespace=default&service=ivr&pair={"LOG_PATH": "/home/tong/www/log/ivr", "LOG_FILE": "application.log"}' http://127.0.0.1:8888/v1/deploy/configmap
```


## 7 发布流程

1) 创建发布任务

```
curl -H 'content-type: application/json' -d '{"name": "ivr test", "summary": "test", "service": "ivr",  "module_list": [{"name": "ivr", "branch": "yy"}], "creator": "yangjinlong", "rd": "yangjinlong", "qa": "yangjinlong", "pm": "yangjinlong"}' http://127.0.0.1:8888/v1/pipeline/create
```

2) 打tag

```
curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/deploy/tag
```

3) 构建镜像

```
curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/deploy/image
```

4) 发布沙盒

```
curl -d "pipeline_id=4&phase=sandbox&username=yangjinlong" http://127.0.0.1:8888/v1/deploy/do | jq .
```

5) 发布全量

```
curl -d "pipeline_id=4&phase=online&username=yangjinlong" http://127.0.0.1:8888/v1/deploy/do | jq .
```

6) 部署完成

```
curl -d "pipeline_id=4" http://127.0.0.1:8888/v1/deploy/finish | jq .
```

7) 回滚

```
curl -d "pipeline_id=4&username=yangjinlong" http://127.0.0.1:8888/v1/rollback/ro | jq .
```

8) 发布cronjob

```
curl -d 'namespace=default&service=ivr&command=sleep 60&schedule=*/10 * * * *' http://127.0.0.1:8888/v1/cronjob/create | jq .
```

9) 删除cronjob

```
curl -d 'namespace=default&service=ivr&job_id=7' http://127.0.0.1:8888/v1/cronjob/delete
```
