## all: Building binaries
build:
	mkdir -p ./storage
	go build -o ./file-api/producer ./file-api/  
	go build -o ./processing-api/consumer ./processing-api/ 

start-producer:
	./file-api/producer 

start-consumer:
	./processing-api/consumer 