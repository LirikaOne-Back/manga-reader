#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Запускаем манга-читалку в Docker...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker не установлен. Пожалуйста, установите Docker.${NC}"
    exit 1
fi

if lsof -Pi :5433 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${YELLOW}Порт 5433 уже используется. Возможно, контейнер PostgreSQL уже запущен.${NC}"
fi

if lsof -Pi :6380 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${YELLOW}Порт 6380 уже используется. Возможно, контейнер Redis уже запущен.${NC}"
fi

echo -e "${YELLOW}Останавливаем существующие контейнеры, если они есть...${NC}"
docker-compose down 2>/dev/null || true

problematic_containers=$(docker ps -a --filter "name=manga-reader" -q)
if [ ! -z "$problematic_containers" ]; then
    echo -e "${YELLOW}Удаляем проблемные контейнеры...${NC}"
    docker rm -f $problematic_containers 2>/dev/null || true
fi

echo -e "${GREEN}Запускаем контейнеры...${NC}"
docker-compose up -d --build

echo -e "${YELLOW}Ожидаем готовности PostgreSQL...${NC}"
RETRIES=10
until docker-compose exec postgres pg_isready -U lirika -d manga_reader_app || [ $RETRIES -eq 0 ]; do
    echo -e "${YELLOW}Ожидаем запуска PostgreSQL, осталось попыток: $RETRIES${NC}"
    RETRIES=$((RETRIES-1))
    sleep 5
done

if [ $RETRIES -eq 0 ]; then
    echo -e "${RED}PostgreSQL не запустился вовремя. Проверьте логи: docker-compose logs postgres${NC}"
    exit 1
fi

echo -e "${GREEN}Применяем миграции PostgreSQL...${NC}"
docker-compose exec app /app/manga-reader migrate up

echo -e "${YELLOW}Проверка статуса приложения...${NC}"
RETRIES=5
until curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ | grep -q "200" || [ $RETRIES -eq 0 ]; do
    echo -e "${YELLOW}Ожидаем запуска приложения, осталось попыток: $RETRIES${NC}"
    RETRIES=$((RETRIES-1))
    sleep 2
done

echo -e "${GREEN}Манга-читалка успешно запущена!${NC}"
echo -e "${GREEN}Сервер доступен по адресу: http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}PostgreSQL доступен на порту 5433${NC}"
echo -e "${YELLOW}Redis доступен на порту 6380${NC}"
echo ""
echo -e "${YELLOW}Для остановки используйте: docker-compose down${NC}"
echo -e "${YELLOW}Для просмотра логов используйте: docker-compose logs -f${NC}"
echo -e "${YELLOW}Для доступа к БД используйте: docker-compose exec postgres psql -U lirika -d manga_reader_app${NC}"