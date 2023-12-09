# please replace it with the path where you have your custom fixture(s).
MOUNT_SOURCE=$(HOME)/data
deps:
	go mod download
test:
	go test -v -count=1  ./... 
build:
	docker build -t recipe-stats-calculator .
run:
	docker run -it -v $(MOUNT_SOURCE):/app/data recipe-stats-calculator
