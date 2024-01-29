.PHONY: all react backend clean

all: react backend

react:
	cd frontend/hangman-frontend-react/ && rm -rf ../../dist/build && npm run build && mv build/ ../../dist/

backend:
	cd backend/hangman-backend/ && rm -f ../../dist/hangman && go build -o ../../dist/hangman && cd ../..

clean:
	rm -rf dist

