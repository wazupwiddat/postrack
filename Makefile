BIN := bin
SERVER := postrack
SRC := cmd/main.go
AWS_ECR = 026450499422.dkr.ecr.us-east-1.amazonaws.com

all: build

build: $(SRC)
	go build -o $(BIN)/$(SERVER) $(SRC)

copy:

build-image:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(AWS_ECR)
	docker build -t $(SERVER) .
	docker tag $(SERVER):latest $(AWS_ECR)/$(SERVER):latest
	docker push $(AWS_ECR)/$(SERVER):latest

run-image:
	docker run -it -p 8080:8080 -v ~/.aws:/.aws $(SERVER):latest
