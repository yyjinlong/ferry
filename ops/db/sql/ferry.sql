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
    deploy_path varchar(100),                        -- 服务部署路径
    multi_phase bool default true,                   -- 服务是否是多阶段部署(分级发布)
    rd varchar(50) not null,                         -- 该服务对应的rd
    op varchar(50) not null,                         -- 该服务对应的op

    replicas int default 0,                          -- 服务的副本数(在线的)
    container json,                                  -- 服务的容器配置信息
    volume json,                                     -- 服务的数据卷配置信息

    online_group varchar(20) default 'none',         -- 当前在线组(blue、green), 默认是none(表示服务未上线)
    lock varchar(100) not null default '',           -- 服务锁
    reserve_time int default 60,                     -- 服务停止时预留多长时间再关闭

    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);

--
-- 模块表
-- 比如: 有个test服务, 分前端vue项目、后端python项目. 并且前后端不分开部署.
--
create table if not exists module (
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
    module_id int not null,                                                    -- 变更模块
    deploy_branch varchar(20) default 'master',                                -- 变更模块对应的部署分支 master上线、分支上线
    code_tag varchar(50),                                                      -- 基于代码模块打的tag
    create_at timestamp not null default now()
);

--
-- 流程镜像
-- 注: 每次发布都全量记录该服务下的所有模块的镜像信息, 如果一个模块没有变更则使用历史镜像.
--
create table if not exists pipeline_image (
    id serial primary key,
    pipeline_id int not null,                                                  -- 对应的流水线
    module_id int not null,                                                    -- 模块
    image_url varchar(200),                                                    -- 基于模块构建的镜像地址
    image_tag varchar(50),                                                     -- 基于模块构建的镜像tag
    create_at timestamp not null default now()
);


--
-- 流程阶段
--
create table if not exists pipeline_phase (
    id serial primary key,
    pipeline_id int not null,                                                     -- 对应的流水线
    name varchar(20) check(name in ('image', 'sandbox', 'online')),               -- 部署阶段: 镜像构建、沙盒、全流量
    status int not null check(status in (0, 1, 2, 3, 4)) default 0,               -- 0 待执行 1 执行中 3 执行成功  4 执行失败
    log text,                                                                     -- 阶段日志
    deployment text,                                                              -- 生成的deployment json串
    create_at timestamp not null default now(),
    update_at timestamp not null default now()
);


-- 插入命名空间
insert into namespace (name, creator) values('default', 'yangjinlong');


commit;
