# Путь к main-пакету клиента.
CLIENT_PKG := ./cmd/client
SERVER_PKG := ./cmd/server
MIGRATOR_PKG := ./cmd/migrator

# Версия берётся из последнего git-тега (например, v1.2.3), либо "dev",
# если тегов нет / вне git-репозитория. Можно переопределить снаружи:
#   make build VERSION=1.0.0
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Дата сборки в UTC, ISO 8601 — воспроизводимо и однозначно для любого таймзон.
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)

OUT_DIR := bin

.PHONY: build build-linux build-windows build-darwin clean

# Сборка под текущую платформу (GOOS/GOARCH возьмутся из окружения).
build:
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client $(CLIENT_PKG)
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server $(SERVER_PKG)
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator $(MIGRATOR_PKG)

build-linux:
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client-linux-amd64 $(CLIENT_PKG)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client-linux-arm64 $(CLIENT_PKG)

	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server-linux-amd64 $(SERVER_PKG)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server-linux-arm64 $(SERVER_PKG)

	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator-linux-amd64 $(MIGRATOR_PKG)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator-linux-arm64 $(MIGRATOR_PKG)

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client-windows-amd64.exe $(CLIENT_PKG)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server-windows-amd64.exe $(SERVER_PKG)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator-windows-amd64.exe $(MIGRATOR_PKG)

build-darwin:
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client-darwin-amd64 $(CLIENT_PKG)
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-client-darwin-arm64 $(CLIENT_PKG)

	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server-darwin-amd64 $(SERVER_PKG)
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-server-darwin-arm64 $(SERVER_PKG)

	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator-darwin-amd64 $(MIGRATOR_PKG)
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/passwordmanager-migrator-darwin-arm64 $(MIGRATOR_PKG)

# Собрать бинарники под все требуемые платформы одной командой.
build-all: build-linux build-windows build-darwin

clean:
	rm -rf $(OUT_DIR)