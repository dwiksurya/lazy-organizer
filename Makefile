.PHONY: cli gui test clean

cli:
	go build -o lazy-organizer .

gui:
	go build -o lazy-organizer-gui ./cmd/gui/

test:
	go test ./... -v

clean:
	rm -f lazy-organizer lazy-organizer-gui

# Cross-compile CLI (pure Go, no CGO needed)
cross-cli:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/lazy-organizer-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/lazy-organizer-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-windows-amd64.exe .

# GUI requires CGO + OpenGL. Cross-compile with fyne-cross:
#   go install github.com/fyne-io/fyne-cross@latest
#   fyne-cross linux -arch amd64,arm64
#   fyne-cross darwin -arch amd64,arm64
#   fyne-cross windows -arch amd64
