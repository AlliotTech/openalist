# OpenAList

[中文](./README_cn.md) | English

[![Release](https://img.shields.io/github/v/release/AlliotTech/openalist)](https://github.com/AlliotTech/openalist/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/alliot/alist)](https://hub.docker.com/r/alliot/alist)
[![License](https://img.shields.io/github/license/AlliotTech/openalist)](./LICENSE)

## Introduction

**OpenAList** is an independently maintained, community-driven fork that originated from [Alist](https://github.com/alist-org/alist) v3.45.0. It focuses on security hardening, storage compatibility, and practical self-hosted deployment.

- [Documentation](https://alist.iots.vip/)
- [Releases](https://github.com/AlliotTech/openalist/releases)
- [Docker Hub](https://hub.docker.com/r/alliot/alist)

## Highlights

- Supports 60+ cloud drive, object storage, local storage, and virtual storage drivers
- Provides WebDAV, FTP, SFTP, and S3-compatible access
- Includes security hardening for path traversal, XSS, SSRF, TLS, redirects, and remote downloads
- Continuously maintains storage drivers and download/upload compatibility
- Publishes multi-platform binaries and multi-architecture Docker images
- Keeps most Alist v3 behavior while remaining independently maintained

See the [release history](https://github.com/AlliotTech/openalist/releases) for detailed changes. Some upstream drivers, including Lark and Quqi, are not included in OpenAList.

## Quick Start

### Docker

```bash
docker run -d \
  --name alist \
  --restart unless-stopped \
  -p 5244:5244 \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

Open `http://localhost:5244` after the container starts. On a new installation, OpenAList generates a random administrator password and prints it in the logs:

```bash
docker logs alist
```

You can instead set the initial password before the data directory is created:

```bash
docker run -d \
  --name alist \
  --restart unless-stopped \
  -p 5244:5244 \
  -e ALIST_ADMIN_PASSWORD='replace-with-a-strong-password' \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

To reset the administrator password for an existing installation:

```bash
docker exec -it alist ./alist admin set 'new-strong-password'
```

For reproducible production deployments, replace `latest` with a version tag from [Docker Hub](https://hub.docker.com/r/alliot/alist/tags).

#### Image variants

| Image tags | Included components |
| --- | --- |
| `latest`, `vX.Y.Z` | OpenAList |
| `latest-ffmpeg`, `vX.Y.Z-ffmpeg` | OpenAList and FFmpeg |
| `latest-aria2`, `vX.Y.Z-aria2` | OpenAList and aria2 |
| `latest-aio`, `vX.Y.Z-aio` | OpenAList, FFmpeg, and aria2 |

Docker images are published for `linux/amd64` and `linux/arm64`. A sample [docker-compose.yml](./docker-compose.yml) is also included; review its volume path and environment variables before running `docker compose up -d`.

### Prebuilt binaries

Download the archive for your operating system and architecture from [GitHub Releases](https://github.com/AlliotTech/openalist/releases), extract it, and run:

```bash
./alist server
```

The default data directory is `./data`, and the default web address is `http://localhost:5244`.

### Build from source

Requirements:

- Git
- Go 1.25 or later
- A C compiler for SQLite/CGO

```bash
git clone https://github.com/AlliotTech/openalist.git
cd openalist
go build -tags=jsoniter -o alist .
./alist server
```

`build.sh` is intended for release and CI workflows and requires an explicit build mode. It is not the normal local build command.

## Migrating from Alist

Back up the data directory before migrating.

OpenAList uses a different static password salt from upstream Alist. When reusing a data directory created by upstream Alist, existing password hashes will no longer validate. Reset the administrator password after switching, then reset the passwords of other local users from the administration interface as needed.

Fresh OpenAList installations are not affected by this migration note.

## Configuration and Credentials

- The default configuration file is `data/config.json`.
- Configuration can also be supplied with `ALIST_`-prefixed environment variables.
- Use local or provider-approved tools to obtain cloud storage credentials. Do not submit tokens to untrusted online services.
- For OneDrive, consider [alist-onedrive-api](https://github.com/vtzp/alist-onedrive-api) or mount a provider through rclone/WebDAV.

Refer to the [documentation](https://alist.iots.vip/) for storage-specific configuration.

## Development

The development script starts the Go backend and the frontend development server together. By default, the backend and frontend repositories must be sibling directories.

```bash
git clone https://github.com/AlliotTech/openalist.git
git clone --recurse-submodules https://github.com/AlliotTech/openalist-web.git

cd openalist-web
pnpm install --frozen-lockfile

cd ../openalist
./dev.sh
```

The frontend currently requires Node.js 22.22.1 or later and pnpm 11. Set `OPENALIST_WEB_DIR` if the frontend repository is stored elsewhere.

See [CONTRIBUTING.md](./CONTRIBUTING.md) before submitting a pull request.

## Support

- Report reproducible bugs through [GitHub Issues](https://github.com/AlliotTech/openalist/issues).
- Use the [documentation](https://alist.iots.vip/) for deployment and storage setup.
- Include the OpenAList version, deployment method, relevant logs, and reproduction steps in bug reports.

## Related Repositories

- [OpenAList backend](https://github.com/AlliotTech/openalist)
- [OpenAList web](https://github.com/AlliotTech/openalist-web)
- [OpenAList documentation](https://github.com/AlliotTech/openalist-docs)

## Project Background

- [Upstream discussion #8649](https://github.com/AlistGo/alist/issues/8649)
- [Upstream discussion #8651](https://github.com/AlistGo/alist/issues/8651)

## Acknowledgments

- The original [Alist project](https://github.com/alist-org/alist)
- All OpenAList and upstream contributors

## License

OpenAList is licensed under the [GNU Affero General Public License v3.0](./LICENSE).
