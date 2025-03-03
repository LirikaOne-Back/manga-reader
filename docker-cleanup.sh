#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Принудительно останавливаем все контейнеры, связанные с проектом...${NC}"

docker-compose down 2>/dev/null || true

containers=$(docker ps -a | grep -E 'manga-reader|postgres|redis' | awk '{print $1}')
if [ ! -z "$containers" ]; then
    echo -e "${YELLOW}Найдены следующие контейнеры:${NC}"
    docker ps -a | grep -E 'manga-reader|postgres|redis'

    echo -e "${YELLOW}Принудительно удаляем контейнеры...${NC}"
    for container in $containers; do
        echo -e "${YELLOW}Удаление контейнера $container${NC}"
        docker rm -f $container 2>/dev/null || echo -e "${RED}Не удалось удалить контейнер $container${NC}"
    done
else
    echo -e "${GREEN}Контейнеры проекта не найдены.${NC}"
fi

echo -e "${YELLOW}Проверяем и удаляем неиспользуемые volumes...${NC}"
docker volume prune -f

echo -e "${YELLOW}Проверяем сети Docker...${NC}"
networks=$(docker network ls | grep manga-network | awk '{print $1}')
if [ ! -z "$networks" ]; then
    echo -e "${YELLOW}Удаляем сети проекта...${NC}"
    for network in $networks; do
        docker network rm $network 2>/dev/null || echo -e "${RED}Не удалось удалить сеть $network${NC}"
    done
fi

echo -e "${GREEN}Очистка завершена. Теперь можно запустить проект заново.${NC}"
echo -e "${YELLOW}Запустите: ./start.sh${NC}"