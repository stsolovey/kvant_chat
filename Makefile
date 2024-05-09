# Пути
CMD_SEVERR_PATH=./cmd/chat_server/
BIN_PATH=./bin/
SERVER_EXECUTABLE=chat_server


# Запуск окружения и сервера
ups: up-deps run_server

# Запуск сервера
run_server:
	go run cmd/chat_server/main.go

# Запуск окружения
up-deps:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml up -d

# Остановка окружения
down-deps:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml down

# Компиляция проекта
build:
	mkdir -p $(BIN_PATH)
	go build -o $(BIN_PATH)$(SERVER_EXECUTABLE) $(CMD_PATH)main.go

# View output (компоуза)
logs:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml logs

# Запуск приложения в терминале
run-app: build
	$(BIN_PATH)$(SERVER_EXECUTABLE)

# Запуск приложения в фоне (сохраняем pid чтобы кикнуть позже)
run-app-background: build
	$(BIN_PATH)$(SERVER_EXECUTABLE) & echo $$! > $(BIN_PATH)PID

# Остановка приложения по pid
stop-app:
	if [ -f $(BIN_PATH)PID ]; then \
		kill `cat $(BIN_PATH)PID` || true; \
		rm $(BIN_PATH)PID; \
	fi



# Остановка всего: приложения и зависимостей
down: stop-app down-deps

# Тестирование: старт окружения и приложения, тест, стоп
test: up-deps # run-app-background
	sleep 5 # Даём приложению время для запуска
	go test ./... -count=1; result=$$?; \
	# make stop-app; \
	# make down-deps; \
	# exit $$result

testv: up-deps # run-app-background
	sleep 5 # Allow time for the application to start
	go test ./... -count=1 -v; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

tidy:
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	go mod tidy

lint: tidy
	golangci-lint run ./...
