.PHONY: all react backend clean

all: react backend serve
build: react backend


react:
	cd frontend/hangman-frontend-react/ && rm -rf ../../dist/build && npm run build && mv build/ ../../dist/ && cd ../..

backend:
	cd backend/hangman-backend/ && rm -f ../../dist/hangman && go.exe build -o ../../dist/hangman && cd ../..


serve:
	cd dist/&&./hangman



dictionary:
	go run dictionarysetup.go && mv words.db dist/
clean:
	rm -rf dist

