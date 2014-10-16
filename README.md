# 应用消息服务 [![Build Status](https://drone.io/github.com/EPICPaaS/appmsgsrv/status.png)](https://drone.io/github.com/EPICPaaS/appmsgsrv/latest)

在 [gopush-cluster](https://github.com/Terry-Mao/gopush-cluster) 基础上做了扩展，加入了用户、群、组织机构等应用元素。

## 概要

* 在 gopush 中进行实现（web 进程）
* 用户登录后返回
  * id：作为 gopush 的 key
  * token：用于身份验证
* 用户/应用/群/组织机构都是使用 id 作为 key
* 用户未登录时使用默认的 key
* 用户发送群组消息时转为对该群组内所有用户的 id 批量发送
* 对外提供对应用数据管理的 RESTful 接口，例如用户 CRUD

### 数据库表

* 用户（user、user_user）
* 组织机构（tenant、org、org_user）
* 应用（application）
* 群（qun、qun_user）
* 客户端版本（client_version）

### 会话

一个会话对应一个推送连接，连接断开会话终止。

可以针对会话级别进行推送控制，请参考[会话推送](https://github.com/EPICPaaS/appmsgsrv/issues/1)。

### Name

发送/监听时使用 `Name` 作为 gopush-cluster 的 key：`数据库记录 id` + `_` + `会话 id` + `@后缀`

#### @后缀

* 组织机构单位：@tenant
* 组织机构部门：@org
* 群：@qun
* 用户：@user
* 应用：@app

开发中可以使用定义在 `app/config.go` 中的常量。

## 开发

* 熟悉 gopush-cluster 的[架构](https://camo.githubusercontent.com/3c2f6df17ff0bace9f88e657819160f0bcb14a8c/687474703a2f2f7261772e6769746875622e636f6d2f54657272792d4d616f2f676f707573682d636c75737465722f6d61737465722f77696b692f6172636869746563747572652f6172636869746563747572652e6a7067)
* 项目只能在 Linux 下开发（通过虚拟机方式方便一些）。

### web 模块

迁出项目后需要修改 web.conf，用于在当前目录下运行的默认配置文件。
* base.app.bind：应用接口监听端口
* zookeeper.addr：gopush-cluster 集群节点通知n
* user daniel：当前操作系统登录用户 
* IP 绑定

启动命令 `./web -v=1 -logtostderr=true`

### Postman

Postman 是一个 Chrome 扩展，用于开发时调试 HTTP 接口。

* 在线[安装](https://chrome.google.com/webstore/detail/postman-rest-client/fdmmgilgnpjigdojojpjoooidkmcomcm)（**需翻墙**） / 离线[安装](https://github.com/a85/POSTMan-Chrome-Extension)
* 导入[测试用例](https://www.getpostman.com/collections/cba11454feb866c965c3)
* 修改测试用例中的 IP，开始测试

### GitBook

* [《使用 GitBook 写文档》](http://88250.b3log.org/write-doc-via-gitbook)
* **开发的同时我们也需要同步完善文档**
* https://github.com/EPICPaaS/youxin-dev-guide
* [《有信开发指南》](http://88250.gitbooks.io/youxin-dev-guide)

## 部署

* static、view 需要放到 bin 目录下
