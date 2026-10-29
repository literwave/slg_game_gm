.PHONY: mongo server web-install web-dev web-build build

mongo:
	docker compose -f deploy/docker-compose.yml up -d

server:
	cd server && go run ./cmd/gm-server -config configs/config.yaml

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# 构建前端并编译后端，产物：server/bin/gm-server + web/dist
build: web-build
	cd server && go build -o bin/gm-server ./cmd/gm-server
