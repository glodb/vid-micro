## Running on local machine
go run main.go -env=DEV -con=AUTHSERVICE
go run main.go -env=DEV -con=TITLESSERVICE
go run main.go -env=DEV -con=CONTENTSERVICE
go run main.go -env=DEV -con=SESSIONSCLEANERSERVICE


## Compiling the docker file
sudo docker build -t vidmicro -f dockerfiles/Dockerfile .

## Running AuthService
sudo docker run --log-opt max-size=10m --log-opt max-file=3 -d --network host --restart unless-stopped --name authservice vidmicro -env=DEV -con=AUTHSERVICE

## Running TitlesService
sudo docker run --log-opt max-size=10m --log-opt max-file=3 -d --network host --restart unless-stopped --name titlesservice vidmicro -env=DEV -con=TITLESSERVICE

## Running ContentService
sudo docker run --log-opt max-size=10m --log-opt max-file=3 -d --network host --restart unless-stopped --name contentservice vidmicro -env=DEV -con=CONTENTSERVICE

## Running SessionService
sudo docker run --log-opt max-size=10m --log-opt max-file=3 -d --network host --restart unless-stopped --name sessionscleanerservice vidmicro -env=DEV -con=SESSIONSCLEANERSERVICE

## Applications to start on local machine
sudo docker start nats-server
sudo docker start my-postgres-container
sudo docker run -d -it --rm -p 7700:7700 -v $(pwd)/data.ms:/data.getmeili.io getmeili/meilisearch:latest