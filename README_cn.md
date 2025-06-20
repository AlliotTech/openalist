# OpenAList
[English](./README.md) | 中文 
## 注意    
> 这是基于版本 3.45.0 的 [Alist](https://github.com/alist-org/alist) 分支。 

文档站已经部署: http://alist.iots.vip/  

至于各网盘的 Token 获取，强烈建议使用离线方案（项目中的原作者提供的 API 我已经替换为黑洞， 避免安全风险。）


### onedrive 处置方法
原先依赖的是 `api.nn.ci` 这个域名提供的 API 服务。 由于这个服务并不开源，因此为了安全考虑尽量替换掉它。  
做法：  
去 Azure 应用里面将这个应用程序删除。 (寻找回调地址为 https://api.nn.ci/alist/ali_open/token 的即可）

那么现在还想用 onedrive 怎么办？

方法一: rclone 挂载 webdav 给 alist
方法二： 用如下类似： https://github.com/vtzp/alist-onedrive-api 项目 本地生成 refresh_token (用法很简单，下载 index.html 在本地双击打开，然后按照提示创建应用程序，并填入 client_id secret 等进行手动操作即可)  


## 描述  
OpenAList 是原始 Alist 文件列表程序的一个分支版本。  

鉴于好几个 fork 的组织还在 onboard 阶段，并且难辨真假。  

为了自用，我已经 fork 并且修改了这部分的代码。急用可以直接用我的镜像 `alliot/alist:latest`

> 需要注意， 由于修改了静态密码 salt, 所以用这个镜像需要重置密码 
> `docker exec -it alist /bin/sh`
> 然后执行 `./alist admin set my_new_password` ）

构建镜像来自如下仓库 CI, 不放心的可以自行审查:  
https://github.com/AlliotTech/openalist  
https://github.com/AlliotTech/openalist-web  
https://github.com/AlliotTech/openalist-docs  

强烈建议仅将此作为临时方案，因为我只是给自己和几个朋友自用的。



## 功能  
- 原版 Alist 功能  
- UI优化, 驱动优化  
- 部分网盘功能增强(自行探索...)

## 贡献  
欢迎贡献！请随时提交 Pull Request。

## 致谢  
- 原始 [Alist 项目](https://github.com/alist-org/alist)

## 更多  
https://github.com/AlistGo/alist/issues/8649  
https://github.com/AlistGo/alist/issues/8651  
...