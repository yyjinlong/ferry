start transaction;

--
-- 命名空间
--
create table if not exists namespace (
    id serial primary key,
    name varchar(32) not null unique,                -- 命名空间的名字
    creator varchar(50) not null,                    -- 命名空间的创建人
    create_at timestamp not null default now()
);

--
-- 服务表
--
create table if not exists service (
    id serial primary key,
    namespace_id int not null,                       -- 服务所在命名空间

    name varchar(32) not null unique,                -- 服务名
    image_addr varchar(500) not null,                -- 服务镜像地址
    quota_cpu varchar(20) not null,                  -- 服务容器正常使用的cpu配额
    quota_max_cpu varchar(20) not null,              -- 服务容器最大使用的cpu配额
    quota_mem varchar(20) not null,                  -- 服务容器正常使用的mem配额
    quota_max_mem  varchar(20) not null,             -- 服务容器最大使用的mem配额
    replicas int default 0,                          -- 服务的副本数(在线的)
    volume json,                                     -- 服务的数据卷配置信息
    reserve_time int default 60,                     -- 服务停止时预留多长时间再关闭
    port int,                                        -- 服务端口
    container_port int,                              -- 容器端口
    online_group varchar(20) default '',             -- 当前在线组(blue、green), 默认是空
    deploy_group varchar(20) default 'blue',         -- 当前发布组(blue、green), 默认为blue
    multi_phase bool default true,                   -- 服务是否是多阶段部署(分级发布)
    lock varchar(100) not null default '',           -- 服务锁
    rd varchar(50) not null,                         -- 该服务对应的rd
    op varchar(50) not null,                         -- 该服务对应的op

    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);

--
-- 代码模块表
--
-- 一个服务可能对应一个或多个代码模块, 比如ivr服务只有一个git仓库地址, 也就是只依赖一个代码模块.
--
create table if not exists code_module (
    id serial primary key,
    service_id int not null,
    name varchar(50) not null,                                                             -- 模块名
    language varchar(20) not null,                                                         -- 模块对应的语言
    repos_name varchar(10) not null check(repos_name in ('GIT', 'SVN')) default 'GIT',     -- 模块使用的仓库名
    repos_addr varchar(200),                                                               -- 模块所在的仓库地址
    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);

--
-- 流水线
--
create table if not exists pipeline (
    id serial primary key,
    service_id int not null,                                                   -- 上线的服务
    name varchar(100) not null,                                                -- 上线名称
    summary text not null,                                                     -- 上线概要
    creator varchar(50) not null,                                              -- 创建人
    rd varchar(500) not null,                                                  -- 项目rd
    qa varchar(200),                                                           -- 项目qa
    pm varchar(200) not null,                                                  -- 项目pm
    status int not null check(status in (0, 1, 2, 3, 4, 5, 6, 7)) default 0,   -- 0 待上线 1 上线中 2 上线成功 3 上线失败 4 回滚中 5 回滚成功 6 回滚失败 7 流程终止
    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);

--
-- 流程变更
--
create table if not exists pipeline_update (
    id serial primary key,
    pipeline_id int not null,                                                  -- 对应的流水线
    code_module_id int not null,                                               -- 变更的代码模块
    deploy_branch varchar(20) default 'master',                                -- 变更的代码模块对应的部署分支 master上线、分支上线
    code_tag varchar(50),                                                      -- 基于代码模块打的tag
    create_at timestamp not null default now()
);

--
-- 服务镜像
--
create table if not exists pipeline_image (
    id serial primary key,
    pipeline_id int not null,                                                  -- 对应的流水线
    image_url varchar(200),                                                    -- 基于所有代码模块构建的服务镜像地址
    image_tag varchar(50),                                                     -- 基于所有代码模块构建的服务镜像tag
    create_at timestamp not null default now()
);

--
-- 流程阶段
--
create table if not exists pipeline_phase (
    id serial primary key,
    pipeline_id int not null,                                                     -- 对应的流水线
    name varchar(20) check(name in ('image', 'sandbox', 'online')),               -- 部署阶段: 镜像构建、沙盒、全流量
    kind varchar(20) check(kind in ('deploy', 'rollback')),
    status int not null check(status in (0, 1, 2, 3, 4)) default 0,               -- 0 待执行 1 执行中 3 执行成功  4 执行失败
    log text,                                                                     -- 阶段日志
    deployment text,                                                              -- 生成的deployment json串
    resource_version varchar(32),                                                 -- 该阶段deployment的resourceVersion
    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);


-- 插入命名空间
insert into namespace (name, creator) values('default', 'yangjinlong');

-- 插入测试服务
insert into service(namespace_id, name, image_addr, quota_cpu, quota_max_cpu, quota_mem, quota_max_mem, replicas, volume, rd, op) values(1, 'ivr', '10.12.28.4:80/service/ivr', '1000', '1024', '1024', '2000', 2, '[{"newvolume_type": "hostPath", "hostpath_type": "DirectoryOrCreate", "newvolume_name": "logs", "hostpath": "/home/logs/ivr"}]', 'yangjinlong', 'yangjinlong');

-- 插入测试代码模块
insert into code_module(service_id, name, language, repos_name, repos_addr) values(1, 'ivr', 'python', 'GIT', 'http://127.0.0.1:4567/devops/ivr');

-- 测试pipline
insert into pipeline(service_id, name, summary, creator, rd, qa, pm) values(1, 'ONLINE TEST', 'TEST', 'yangjinlong', 'yangjinlong', 'yangjinlong', 'yangjinlong');
insert into pipeline_update(pipeline_id, code_module_id, deploy_branch, code_tag) values(1, 1, 'master', 'release_ivr_20210817_153246');

commit;
