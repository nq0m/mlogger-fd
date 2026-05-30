.PHONY: dev build clean

dev:
	cd frontend && npm run build
	go build -o fdlogger .
	./fdlogger

build:
	cd frontend && npm run build
	go build -o fdlogger .

clean:
	rm -f fdlogger
	rm -rf frontend/build
