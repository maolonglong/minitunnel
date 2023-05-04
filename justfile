build:
  CGO_ENABLED=0 go build -v -trimpath -ldflags="-s -w" ./cmd/mt

fmt:
  gosimports -local go.chensl.me -w .
  golines --max-len=88 --base-formatter=gofumpt -w .

clean:
  rm -rf dist
  rm -f mt
