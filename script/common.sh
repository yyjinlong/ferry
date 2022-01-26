#!/bin/bash

err=1

function echoinfo() {
    echo -e "\033[35m$1\033[0m"
}

function echowarn() {
    echo -e "\033[33m$1\033[0m"
}

function echoerror() {
    echo -e "\033[31m$1\033[0m"
}

function request() {
    for i in $(seq 1 3); do
        r=$(curl -s "$1" | jq .code)
        if [ $? -eq 0 ]; then
            if [ $r -eq 0 ]; then
                break
            else
                echoerror $2
                exit $err
            fi
        fi

        if [ $i -ge 3 ]; then
            echoerror "获取请求失败!"
            exit $err
        fi
    done
}

function report_tag() {
    taskid=$1
    module=$2
    tag=$3
    errmsg="上报tag信息失败!"

    url="$BASE_URL/v1/tag?taskid=$taskid&module=$module&tag=$tag"
    request $url $errmsg
}
