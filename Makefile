.PHONY: all react backend clean

all: react backend



react:
	cd frontend/hangman-frontend-react/ && rm -rf ../../dist/build && npm run build && mv build/ ../../dist/

backend:
	cd backend/hangman-backend/ && rm -f ../../dist/hangman && go build -o ../../dist/hangman && cd ../..

dictionary:
	go run dictionarysetup.go && mv words.db dist/
clean:
	rm -rf dist

