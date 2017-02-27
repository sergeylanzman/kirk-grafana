# Pili 定制 Grafana

## Clone 这个 Repo

- 必须放在 ``src/github.com/grafana/grafana`` 这个路径中. 

## 准备

> 请确认安装了正确的版本

- node 6.9
- npm 3.10
- go 1.8
- yarn 0.20
- grunt 0.45

## 安装

- ``pili_scripts/setup``
- ``npm install``

## 构造

- ``pili_scripts/build_server``
- ``npm run build``

## 启动服务

- ``bin/grafana-server``

## 构造 Plugin

- ``cd data/plugin/xxx``
- ``grunt``

## 注意事项

- 任何针对 Grafana 的修改都用重新构造，并重启服务器。
- 仅修改 Grafana Plugin，仅需要重新构造 Plugin，并刷新浏览器即可。
