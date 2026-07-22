# OpenAList

[English](./README.md) | 中文

[![Release](https://img.shields.io/github/v/release/AlliotTech/openalist)](https://github.com/AlliotTech/openalist/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/alliot/alist)](https://hub.docker.com/r/alliot/alist)
[![License](https://img.shields.io/github/license/AlliotTech/openalist)](./LICENSE)

## 项目简介

**OpenAList** 是一个独立维护的社区分支，最初分叉自 [Alist](https://github.com/alist-org/alist) v3.45.0，重点关注安全加固、存储兼容性与实用的自托管部署体验。

- [文档站](https://alist.iots.vip/)
- [版本发布](https://github.com/AlliotTech/openalist/releases)
- [Docker Hub](https://hub.docker.com/r/alliot/alist)

## 主要特性

- 支持 60 多种网盘、对象存储、本地存储与虚拟存储驱动
- 提供 WebDAV、FTP、SFTP 和 S3 兼容访问
- 包含针对路径穿越、XSS、SSRF、TLS、重定向和远程下载的安全加固
- 持续维护各类存储驱动及上传、下载兼容性
- 发布多平台二进制文件和多架构 Docker 镜像
- 在独立维护的同时，尽可能保持与 Alist v3 的行为兼容

详细变更请查看[版本发布记录](https://github.com/AlliotTech/openalist/releases)。OpenAList 不包含部分上游驱动，包括 Lark 和 Quqi。

## 快速开始

### Docker

```bash
docker run -d \
  --name alist \
  --restart unless-stopped \
  -p 5244:5244 \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

容器启动后访问 `http://localhost:5244`。全新安装时，OpenAList 会生成随机管理员密码并输出到日志：

```bash
docker logs alist
```

也可以在首次创建数据目录前指定初始密码：

```bash
docker run -d \
  --name alist \
  --restart unless-stopped \
  -p 5244:5244 \
  -e ALIST_ADMIN_PASSWORD='请替换为高强度密码' \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

已有安装可通过以下命令重置管理员密码：

```bash
docker exec -it alist ./alist admin set '新的高强度密码'
```

生产环境建议将 `latest` 替换为 [Docker Hub](https://hub.docker.com/r/alliot/alist/tags) 中的固定版本标签，以获得可重复部署。

#### 镜像变体

| 镜像标签 | 包含组件 |
| --- | --- |
| `latest`、`vX.Y.Z` | OpenAList |
| `latest-ffmpeg`、`vX.Y.Z-ffmpeg` | OpenAList 和 FFmpeg |
| `latest-aria2`、`vX.Y.Z-aria2` | OpenAList 和 aria2 |
| `latest-aio`、`vX.Y.Z-aio` | OpenAList、FFmpeg 和 aria2 |

Docker 镜像支持 `linux/amd64` 和 `linux/arm64`。仓库也提供了 [docker-compose.yml](./docker-compose.yml) 示例；运行 `docker compose up -d` 前请先检查其中的数据挂载路径和环境变量。

### 预编译程序

从 [GitHub Releases](https://github.com/AlliotTech/openalist/releases) 下载适合操作系统和架构的压缩包，解压后运行：

```bash
./alist server
```

默认数据目录为 `./data`，默认访问地址为 `http://localhost:5244`。

### 从源码构建

环境要求：

- Git
- Go 1.25 或更高版本
- SQLite/CGO 所需的 C 编译器

```bash
git clone https://github.com/AlliotTech/openalist.git
cd openalist
go build -tags=jsoniter -o alist .
./alist server
```

`build.sh` 用于版本发布和 CI 流程，必须显式指定构建模式，不是常规的本地构建命令。

## 从 Alist 迁移

迁移前请备份数据目录。

OpenAList 与上游 Alist 使用的静态密码盐不同。复用上游 Alist 创建的数据目录时，已有密码哈希将无法通过验证。切换后请先重置管理员密码，再按需通过管理界面重置其他本地用户的密码。

全新安装的 OpenAList 不受此迁移说明影响。

## 配置与凭据安全

- 默认配置文件为 `data/config.json`。
- 也可以通过带 `ALIST_` 前缀的环境变量提供配置。
- 请使用本地工具或服务商认可的方式获取网盘凭据，不要将 Token 提交给不可信的在线服务。
- OneDrive 可考虑使用 [alist-onedrive-api](https://github.com/vtzp/alist-onedrive-api)，或通过 rclone/WebDAV 挂载服务商。

各存储的具体配置请查阅[文档站](https://alist.iots.vip/)。

## 开发

开发脚本会同时启动 Go 后端和前端开发服务器。默认情况下，后端与前端仓库需要位于同一级目录。

```bash
git clone https://github.com/AlliotTech/openalist.git
git clone --recurse-submodules https://github.com/AlliotTech/openalist-web.git

cd openalist-web
pnpm install --frozen-lockfile

cd ../openalist
./dev.sh
```

前端当前要求 Node.js 22.22.1 或更高版本以及 pnpm 11。如果前端仓库位于其他位置，请设置 `OPENALIST_WEB_DIR`。

提交 Pull Request 前请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 获取帮助

- 可复现的问题请提交到 [GitHub Issues](https://github.com/AlliotTech/openalist/issues)。
- 部署和存储配置问题请先查阅[文档站](https://alist.iots.vip/)。
- 反馈问题时请提供 OpenAList 版本、部署方式、相关日志和复现步骤。

## 相关仓库

- [OpenAList 后端](https://github.com/AlliotTech/openalist)
- [OpenAList 前端](https://github.com/AlliotTech/openalist-web)
- [OpenAList 文档](https://github.com/AlliotTech/openalist-docs)

## 项目背景

- [上游讨论 #8649](https://github.com/AlistGo/alist/issues/8649)
- [上游讨论 #8651](https://github.com/AlistGo/alist/issues/8651)

## 致谢

- 原始 [Alist 项目](https://github.com/alist-org/alist)
- 所有 OpenAList 与上游贡献者

## 开源协议

OpenAList 使用 [GNU Affero General Public License v3.0](./LICENSE) 开源。
