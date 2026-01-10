# Requirements Document

## Introduction

本文档定义了 V 多协议代理面板项目的重构需求，包括项目结构优化和 Docker 部署支持。当前项目存在结构混乱、职责不清、重复代码等问题，需要进行系统性重构以提高可维护性和部署便利性。

## Glossary

- **V_Panel**: V 多协议代理面板系统，提供代理服务管理功能
- **Backend**: Go 语言编写的后端服务
- **Frontend**: Vue.js 编写的 Web 管理界面
- **Xray_Core**: 底层代理引擎，支持多种代理协议
- **Docker_Container**: Docker 容器化部署单元
- **Docker_Compose**: 多容器编排工具

## Requirements

### Requirement 1: 后端项目结构重构

**User Story:** As a developer, I want a clean and well-organized project structure, so that I can easily understand, maintain, and extend the codebase.

#### Acceptance Criteria

1. THE Backend SHALL organize code into the following top-level directories: `cmd/`, `internal/`, `pkg/`, `configs/`, `scripts/`, `deployments/`
2. THE Backend SHALL consolidate duplicate packages (e.g., `model/` and `database/`, `db/` and `database/`) into single coherent modules
3. THE Backend SHALL separate business logic from HTTP handlers using a service layer pattern
4. THE Backend SHALL use `internal/` directory for private application code that should not be imported by other projects
5. THE Backend SHALL use `pkg/` directory for code that can be safely imported by external applications
6. WHEN the Backend starts, THE V_Panel SHALL load configuration from environment variables with fallback to config files

### Requirement 2: 数据库层统一

**User Story:** As a developer, I want a single, consistent database abstraction layer, so that I can avoid confusion and reduce code duplication.

#### Acceptance Criteria

1. THE Backend SHALL have exactly one database package located at `internal/database/`
2. THE Backend SHALL remove redundant database packages (`db/`, `model/sqlite.go`, `model/db.go`)
3. THE Backend SHALL use repository pattern for data access with interfaces defined in `internal/repository/`
4. WHEN database operations fail, THE Backend SHALL return structured errors with context

### Requirement 3: API 层重构

**User Story:** As a developer, I want a clean API layer with clear separation of concerns, so that I can easily add new endpoints and maintain existing ones.

#### Acceptance Criteria

1. THE Backend SHALL organize API handlers in `internal/api/handlers/` grouped by domain (auth, proxy, system, etc.)
2. THE Backend SHALL define all routes in a single `internal/api/routes.go` file
3. THE Backend SHALL use middleware for cross-cutting concerns (logging, auth, CORS) in `internal/api/middleware/`
4. THE Backend SHALL remove duplicate handler implementations (consolidate `api/handlers/`, `server/handlers/`)
5. WHEN an API request is received, THE Backend SHALL validate input before processing

### Requirement 4: 代理协议模块化

**User Story:** As a developer, I want proxy protocols to be modular and pluggable, so that I can easily add or modify protocol support.

#### Acceptance Criteria

1. THE Backend SHALL organize proxy protocols in `internal/proxy/protocols/` with each protocol in its own subdirectory
2. THE Backend SHALL define a common `Protocol` interface that all proxy implementations must satisfy
3. THE Backend SHALL consolidate duplicate proxy implementations (e.g., `proxy/vless.go` and `proxy/vless/vless.go`)
4. WHEN a new protocol is added, THE Backend SHALL only require implementing the Protocol interface

### Requirement 5: 配置管理优化

**User Story:** As an operator, I want flexible configuration management, so that I can easily configure the application for different environments.

#### Acceptance Criteria

1. THE Backend SHALL support configuration via environment variables with `V_` prefix
2. THE Backend SHALL support configuration via YAML files in `configs/` directory
3. THE Backend SHALL merge configurations with precedence: environment variables > config files > defaults
4. THE Backend SHALL validate configuration at startup and fail fast with clear error messages
5. WHEN configuration is invalid, THE Backend SHALL log the specific validation error and exit

### Requirement 6: Docker 单容器部署

**User Story:** As an operator, I want to deploy the application using Docker, so that I can easily deploy and manage the application in any environment.

#### Acceptance Criteria

1. THE V_Panel SHALL provide a `Dockerfile` that builds both backend and frontend
2. THE Docker_Container SHALL use multi-stage build to minimize final image size
3. THE Docker_Container SHALL expose port 8080 for the web interface
4. THE Docker_Container SHALL persist data in `/data` volume
5. THE Docker_Container SHALL support configuration via environment variables
6. WHEN the Docker_Container starts, THE V_Panel SHALL automatically initialize the database if not exists

### Requirement 7: Docker Compose 编排

**User Story:** As an operator, I want to use Docker Compose for deployment, so that I can easily manage the application with its dependencies.

#### Acceptance Criteria

1. THE V_Panel SHALL provide a `docker-compose.yml` file for orchestration
2. THE Docker_Compose SHALL define named volumes for data persistence
3. THE Docker_Compose SHALL support environment variable configuration via `.env` file
4. THE Docker_Compose SHALL include health checks for the application container
5. WHEN using Docker_Compose, THE operator SHALL be able to start the application with a single command

### Requirement 8: 前端构建集成

**User Story:** As a developer, I want the frontend build to be integrated into the Docker build process, so that deployment is simplified.

#### Acceptance Criteria

1. THE Dockerfile SHALL build the frontend as part of the multi-stage build
2. THE Backend SHALL serve static frontend files from an embedded filesystem or `/app/web/dist`
3. THE Docker_Container SHALL include pre-built frontend assets
4. WHEN the frontend is updated, THE Docker build SHALL automatically rebuild frontend assets

### Requirement 9: 日志和监控统一

**User Story:** As an operator, I want consistent logging and monitoring, so that I can effectively troubleshoot and monitor the application.

#### Acceptance Criteria

1. THE Backend SHALL consolidate logging into a single `internal/logger/` package
2. THE Backend SHALL output logs in JSON format for container environments
3. THE Backend SHALL support configurable log levels via environment variable
4. THE Docker_Container SHALL output logs to stdout/stderr for container log collection
5. WHEN an error occurs, THE Backend SHALL log structured error information with request context

### Requirement 10: 健康检查和优雅关闭

**User Story:** As an operator, I want the application to support health checks and graceful shutdown, so that I can safely manage container lifecycle.

#### Acceptance Criteria

1. THE Backend SHALL expose `/health` endpoint for liveness checks
2. THE Backend SHALL expose `/ready` endpoint for readiness checks
3. WHEN receiving SIGTERM, THE Backend SHALL gracefully shutdown within 30 seconds
4. WHEN shutting down, THE Backend SHALL complete in-flight requests before terminating
5. THE Docker_Container SHALL use health check endpoint for container health monitoring

### Requirement 11: 清理冗余文件和目录

**User Story:** As a developer, I want a clean project without redundant or orphaned files, so that I can focus on relevant code and reduce confusion.

#### Acceptance Criteria

1. THE V_Panel SHALL remove duplicate `src/` directory at project root (frontend code should only exist in `web/src/`)
2. THE V_Panel SHALL remove orphaned configuration files (`package.json`, `package-lock.json` at root level)
3. THE V_Panel SHALL consolidate duplicate router implementations (`router/` and `web/src/router/`)
4. THE V_Panel SHALL remove unused or duplicate view files (e.g., `*.vue.new` files, duplicate Stats/Roles views)
5. THE V_Panel SHALL organize all deployment-related files in `deployments/` directory
6. THE V_Panel SHALL move database migration files to `internal/database/migrations/`
7. WHEN a file or directory serves no purpose, THE V_Panel SHALL remove it from the repository

### Requirement 12: 目录结构标准化

**User Story:** As a developer, I want a standardized directory structure following Go best practices, so that the project is familiar to other Go developers.

#### Acceptance Criteria

1. THE V_Panel SHALL follow the standard Go project layout with `cmd/`, `internal/`, `pkg/` directories
2. THE V_Panel SHALL place the main application entry point in `cmd/v/main.go`
3. THE V_Panel SHALL organize internal packages by domain: `internal/auth/`, `internal/proxy/`, `internal/monitor/`, etc.
4. THE V_Panel SHALL place shared utilities in `pkg/` only if they are designed for external use
5. THE V_Panel SHALL keep frontend code exclusively in `web/` directory
6. THE V_Panel SHALL place all configuration templates in `configs/` directory
7. THE V_Panel SHALL place all scripts (build, deploy, etc.) in `scripts/` directory
