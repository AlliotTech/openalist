# OpenAList
[English](./README.md) | 中文

## 项目简介

**OpenAList** 是基于 [Alist](https://github.com/alist-org/alist) v3.45.0 的社区驱动分支，致力于提供更安全、可定制、易用的文件列表与管理解决方案。

- **文档站**: http://alist.iots.vip/

## 主要特性

- 🗂️ 支持多种主流网盘和本地存储
- 🔒 更安全的 Token 获取方式（已移除原有不安全 API）
- 🛠️ UI 优化与驱动增强
- 🚀 镜像持续集成，开箱即用
- 🧩 易于二次开发和自定义
- 📝 兼容原版 Alist 的大部分功能

## 快速开始

### 使用 Docker

```bash
docker run -d \
  --name=alist \
  -p 5244:5244 \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

> **注意：**
> 由于静态密码 salt 已更改，首次使用请重置管理员密码：
> ```
> docker exec -it alist /bin/sh
> ./alist admin set <你的新密码>
> ```

### 本地部署

1. 克隆仓库：
   ```bash
   git clone https://github.com/AlliotTech/openalist.git
   cd openalist
   ```
2. 构建并运行：
   ```bash
   ./build.sh
   ./alist server
   ```

## 配置说明

- 配置文件路径：`data/config.json`
- 强烈建议使用离线/本地工具获取各网盘 Token，避免安全风险
- OneDrive 推荐方案：使用 [alist-onedrive-api](https://github.com/vtzp/alist-onedrive-api) 或 rclone 挂载 WebDAV

## 常见问题

- **Q: 如何安全获取各网盘 Token？**
  A: 建议使用本地或离线工具获取，切勿使用不明在线服务。

- **Q: 镜像/程序无法启动？**
  A: 请检查端口占用、数据目录权限等常见问题。

- **Q: 如何反馈 Bug 或建议？**
  A: 欢迎通过 [GitHub Issues](https://github.com/AlliotTech/openalist/issues) 提交。

## 贡献指南

欢迎贡献代码、文档或建议。请先阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 并提交 Pull Request。

## 镜像与相关仓库

- [openalist](https://github.com/AlliotTech/openalist)
- [openalist-web](https://github.com/AlliotTech/openalist-web)
- [openalist-docs](https://github.com/AlliotTech/openalist-docs)

## 致谢

- 原始 [Alist 项目](https://github.com/alist-org/alist)
- 所有开源贡献者

## 更多

- [AlistGo/alist/issues/8649](https://github.com/AlistGo/alist/issues/8649)
- [AlistGo/alist/issues/8651](https://github.com/AlistGo/alist/issues/8651)