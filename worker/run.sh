# Prebuild image
APP_NAME=br2 docker-compose build

go run main.go -owner=$1
