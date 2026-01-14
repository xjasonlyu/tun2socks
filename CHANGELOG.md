# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-01-14

### Added

#### Web Interface
- **Dashboard**: Introduced a modern, single-page, GitHub-style web dashboard for management.
- **Real-time Monitoring**:
  - Service status (Running/Stopped).
  - Process metrics: PID, Uptime, Memory Usage.
  - Traffic statistics: Real-time Upload/Download speeds and total transfer.
- **Proxy Configuration**:
  - Support for SOCKS5, SOCKS4, HTTP, and HTTPS protocols.
  - **Dynamic Update**: Ability to change proxy settings instantly without restarting the service.
  - **Persistence**: Configuration is automatically saved to `~/.config/tun2socks/config.yaml`.
- **Route Management**:
  - Visual interface for adding and deleting network routes.
  - Automatic gateway IP detection for new routes.
  - Smart CIDR auto-completion (defaults to `/32`).
- **Device Status**: Real-time monitoring of TUN device state (UP/DOWN), IP address, and MTU.

#### Backend
- **REST API**:
  - `GET /api/v1/service/events`: Server-Sent Events (SSE) for real-time status updates.
  - `POST /api/v1/proxy`: Endpoint for dynamic proxy switching.
  - `GET/POST/DELETE /api/v1/routes`: Endpoints for routing table management.
- **Engine**:
  - Added logic to automatically create, configure IP (default `198.18.0.1/15`), and bring UP the TUN device on startup.
  - Implemented `UpdateProxy` mechanism for runtime configuration changes.

### Changed
- **CLI**:
  - Added `--restapi` flag to enable the web interface (e.g., `--restapi http://token@127.0.0.1:7777`).
  - `--proxy` flag is now optional if a valid `config.yaml` is present.
- **Frontend**:
  - Refactored from a multi-page sidebar layout to a unified single-page Dashboard.
  - Standardized UI components with a clean, minimal design system.
- **Build System**:
  - All builds now default to embedding the web interface assets.
  - Simplified build process: `make` now produces a complete binary with embedded web dashboard.
  - Added `make web` target for standalone frontend builds.

### Fixed
- Fixed an issue where the TUN device created via API would remain in `DOWN` state.
- Fixed correct PID reporting (previously showing 0).
- Resolved circular dependency issues between `engine` and `restapi` packages.
- Fixed nil pointer panic when accessing proxy configuration before initialization.
