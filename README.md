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
    release层 : 由image进程基于代码自动构建


## 3 业务逻辑

    service
    ├── pipeline     -- 构建
    ├── publish      -- 发布
    └── rollback     -- 回滚

* 新服务上线默认为blue
* 蓝绿两组分别创建对应的deployment
* 不再监听endpoint事件, 等待deployment发布完成, 自动将另一组缩成0, 读取endpoint信息并记录


## 4 依赖准备

### 4.1 节点标签

```
# 业务node
kubectl label node xxxx aggregate=default

# 定时node
kubectl label node xxxx batch=cronjob

# 查看所有node label
kubectl get nodes --show-labels
```

### 4.2 ServiceAccount Token

```
kubectl create serviceaccount nautilus-api -n kube-system

kubectl create clusterrolebinding nautilus-api --clusterrole=cluster-admin --group=kube-system:nautilus-api
```

### 4.3 imagePullSecrets

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
curl -d 'service=ivr' http://127.0.0.1:8888/v1/service
```


## 6 创建configmap

```
curl -d 'namespace=default&service=ivr&pair={"LOG_PATH": "/home/tong/www/log/ivr", "LOG_FILE": "application.log"}' http://127.0.0.1:8888/v1/configmap
```


## 7 发布流程

1) 创建job

```
curl -H 'content-type: application/json' -d '{"name": "ivr test", "summary": "test", "service": "ivr",  "module_list": [{"name": "ivr", "branch": "yy"}], "creator": "yangjinlong", "rd": "yangjinlong", "qa": "yangjinlong", "pm": "yangjinlong"}' http://127.0.0.1:8888/v1/pipeline
```

2) 打tag

```
curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/tag
```

3) 构建镜像

```
curl -d 'pipeline_id=4&service=ivr' http://127.0.0.1:8888/v1/image
```

4) 发布沙盒

```
curl -d "pipeline_id=4&phase=sandbox&username=yangjinlong" http://127.0.0.1:8888/v1/deploy | jq .
```

5) 发布全量

```
curl -d "pipeline_id=4&phase=online&username=yangjinlong" http://127.0.0.1:8888/v1/deploy | jq .
```

6) 部署完成

```
curl -d "pipeline_id=4" http://127.0.0.1:8888/v1/finish | jq .
```

7) 回滚

```
curl -d "pipeline_id=4&username=yangjinlong" http://127.0.0.1:8888/v1/rollback | jq .
```

