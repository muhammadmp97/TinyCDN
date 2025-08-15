#!/bin/bash

source .env

header=$(cat <<'EOF'
  _______             __________  _   __
 /_  __(_)___  __  __/ ____/ __ \/ | / /
  / / / / __ \/ / / / /   / / / /  |/ / 
 / / / / / / / /_/ / /___/ /_/ / /|  /  
/_/ /_/_/ /_/\__, /\____/_____/_/ |_/   
            /____/                      
EOF
)

up() {
    if ! docker compose ps -a -q | grep -q .; then
        docker compose up -d
        while true; do
            if (docker compose ps | grep minio | grep Up) > /dev/null 2>&1; then
                cat ./docker/minio.json | docker exec -i tinycdn-minio /bin/sh -c "mc alias set local http://$MINIO_ADDRESS $MINIO_USER $MINIO_PASSWORD && mc mb local/$MINIO_BUCKET_NAME && mc ilm import local/$MINIO_BUCKET_NAME" > /dev/null 2>&1 \
                    && echo -e "\e[32mBucket created successfully!\e[0m"
                break
            fi
            sleep 2
        done
    else
        echo "You've already created the containers."
    fi
}

down() {
    docker compose down
}

start() {
    if ! docker compose start > /dev/null 2>&1; then
        echo -e "\e[31mYou haven't created the containers yet!\nRun the command \e[3m./tinycdn up\e[0m\e[0m" >&2
    fi
}

stop() {
    docker compose stop
}

case "$1" in
    up)
        up
        ;;
    down)
        down
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    *)
        echo "$header"
        echo -e "\nHow can I help you?"
        echo -e "\e[3m./tinycdn.sh up|down|start|stop\e[0m"
        ;;
esac
