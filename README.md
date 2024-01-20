go run main.go -env=DEV -con=AUTHSERVICE
go run main.go -env=DEV -con=TITLESSERVICE
go run main.go -env=DEV -con=CONTENTSERVICE


sudo docker start nats-server
sudo docker start my-postgres-container
