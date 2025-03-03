#!/bin/bash
set -e

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Убедимся, что контейнеры запущены
echo -e "${YELLOW}Проверяем, запущены ли контейнеры...${NC}"
if ! docker-compose ps | grep -q "manga-reader-postgres"; then
    echo -e "${RED}Контейнер PostgreSQL не запущен. Запустите сначала контейнеры: docker-compose up -d${NC}"
    exit 1
fi

# Получаем информацию о БД из docker-compose.yml
PG_USER="lirika"
PG_PASSWORD="evil_god"
PG_DBNAME="manga_reader_app"
PG_HOST="localhost"
PG_PORT="5433"  # Порт на хост-машине, который мы изменили с 5432 на 5433

echo -e "${YELLOW}Применяем миграции к PostgreSQL (${PG_USER}@${PG_HOST}:${PG_PORT}/${PG_DBNAME})...${NC}"

# Проверка наличия migrate (golang-migrate)
if ! command -v migrate &> /dev/null; then
    echo -e "${YELLOW}Инструмент migrate не установлен. Устанавливаем...${NC}"

    # Определяем ОС
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm*) ARCH="arm64" ;;
    esac

    MIGRATE_VERSION="v4.16.2"

    # Скачиваем и устанавливаем migrate
    echo -e "${YELLOW}Скачиваем migrate для ${OS}-${ARCH}...${NC}"
    wget -q https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${OS}-${ARCH}.tar.gz -O migrate.tar.gz
    tar -xzf migrate.tar.gz
    chmod +x migrate
    sudo mv migrate /usr/local/bin/
    rm migrate.tar.gz

    echo -e "${GREEN}Инструмент migrate установлен.${NC}"
fi

# Формируем правильный URL для подключения
DB_URL="postgres://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DBNAME}?sslmode=disable"

echo -e "${YELLOW}Путь к миграциям: ./migrations/postgres${NC}"
echo -e "${YELLOW}URL базы данных: ${DB_URL}${NC}"

# Запускаем миграции
echo -e "${YELLOW}Применяем миграции...${NC}"
migrate -path ./migrations/postgres -database "${DB_URL}" up

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Миграции успешно применены!${NC}"
else
    echo -e "${RED}Произошла ошибка при применении миграций.${NC}"
    echo -e "${YELLOW}Проверяем, существуют ли файлы миграций...${NC}"
    ls -la ./migrations/postgres/

    echo -e "${YELLOW}Пробуем создать таблицы вручную...${NC}"
    docker-compose exec -T postgres psql -U ${PG_USER} -d ${PG_DBNAME} < manual_migration.sql
fi

# Проверяем состояние базы данных
echo -e "${GREEN}Проверяем таблицы в базе данных...${NC}"
docker-compose exec postgres psql -U ${PG_USER} -d ${PG_DBNAME} -c "\dt"