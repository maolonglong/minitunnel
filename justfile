build:
  CGO_ENABLED=0 go build -v -trimpath -ldflags="-s -w" ./cmd/mt

test:
  @go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --keep-going --race --trace --timeout=30m

fmt:
  golines --max-len=88 --base-formatter="gofumpt -extra" -w .
  gosimports -local go.chensl.me -w .

deps:
  go install github.com/onsi/ginkgo/v2/ginkgo@latest
  go install github.com/segmentio/golines@latest
  go install mvdan.cc/gofumpt@latest
  go install github.com/rinchsan/gosimports/cmd/gosimports@latest

clean:
  rm -rf dist
  rm -f mt
