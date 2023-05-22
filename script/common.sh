#!/bin/bash

err=1

function echowarn() {
    echo -e "\033[33m$1\033[0m"
}

function echoerror() {
    echo -e "\033[31m$1\033[0m"
}

# 加粗
function echobold() {
    echo -e "\033[1m$1\033[0m"
}

# 输出: ok
# 退出: 0
function request() {
    for i in $(seq 1 3); do
        res=$(curl -s "$1")
        if [ $res == "ok" ];then
            break
        else
            if [ $i -ge 3 ];then
               echo "$2: $res"
               exit $err
            fi
        fi
    done
    echo $res
}

function report_tag() {
    taskid=$1
    module=$2
    tag=$3
    errmsg="上报tag信息失败"

    url="$BASE_URL/v1/deploy/tag?taskid=$taskid&module=$module&tag=$tag"
    request $url $errmsg
}

function report_img() {
    taskid=$1
    module=$2
    image_url=$3
    image_tag=$4
    errmsg="上报镜像信息失败"

    url="$BASE_URL/v1/deploy/image/update?taskid=$taskid&module=$module&image_url=$image_url&image_tag=$image_tag"
    request $url $errmsg
}
