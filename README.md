# OpenAList
[中文](./README_cn.md) | English

## Introduction

**OpenAList** is a community-driven fork of [Alist](https://github.com/alist-org/alist) (based on v3.45.0), aiming to provide a more secure, customizable, and user-friendly file listing and management solution.

- **Document site**: http://alist.iots.vip/

## Features

- 🗂️ Supports multiple mainstream cloud drives and local storage
- 🔒 Safer token acquisition (removed insecure original APIs)
- 🛠️ UI and driver enhancements
- 🚀 Continuous integration Docker image, ready to use
- 🧩 Easy for secondary development and customization
- 📝 Compatible with most original Alist features

## Quick Start

### Using Docker

```bash
docker run -d \
  --name=alist \
  -p 5244:5244 \
  -v /path/to/data:/opt/alist/data \
  alliot/alist:latest
```

> **Note:**
> The static password salt has changed. Please reset the admin password on first use:
> ```
> docker exec -it alist /bin/sh
> ./alist admin set <your_new_password>
> ```

### Local Deployment

1. Clone the repository:
   ```bash
   git clone https://github.com/AlliotTech/openalist.git
   cd openalist
   ```
2. Build and run:
   ```bash
   ./build.sh
   ./alist server
   ```

## Configuration

- Config file path: `data/config.json`
- Strongly recommend using offline/local tools to obtain cloud drive tokens for security
- For OneDrive: use [alist-onedrive-api](https://github.com/vtzp/alist-onedrive-api) or mount WebDAV via rclone

## FAQ

- **Q: How to securely obtain cloud drive tokens?**
  A: Use local or offline tools, never use untrusted online services.

- **Q: The image/program fails to start?**
  A: Check for port conflicts, data directory permissions, etc.

- **Q: How to report bugs or suggestions?**
  A: Please submit via [GitHub Issues](https://github.com/AlliotTech/openalist/issues).

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) and submit a Pull Request.

## Images & Related Repositories

- [openalist](https://github.com/AlliotTech/openalist)
- [openalist-web](https://github.com/AlliotTech/openalist-web)
- [openalist-docs](https://github.com/AlliotTech/openalist-docs)

## Acknowledgments

- Original [Alist Project](https://github.com/alist-org/alist)
- All open-source contributors

## More

- [AlistGo/alist/issues/8649](https://github.com/AlistGo/alist/issues/8649)
- [AlistGo/alist/issues/8651](https://github.com/AlistGo/alist/issues/8651)
