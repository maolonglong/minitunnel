# minitunnel

A Go port of [bore](https://github.com/ekzhang/bore).

> A modern, simple TCP tunnel in Rust that exposes local ports to a remote server, bypassing standard NAT connection firewalls. **That's all it does: no more, and no less.**

```bash
# Installation (requires Go)
go install go.chensl.me/minitunnel/cmd/mt@latest

# On your local machine
mt local 8000
```

This will expose your local port at `localhost:8000` to the public internet at `minitunnel.icu:<PORT>`, where the port number is assigned randomly.
