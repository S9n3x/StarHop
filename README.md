# StarHop – Latency-Driven P2P Proxy Mesh
wait

# schedule
- [x] Complete multi node registratio  
# build
**server**
```bash
go build -ldflags="-s -w" -trimpath -o StarHop.exe .\cmd\server\main.go
```
# develop
**grpc**
```bash
protoc --go_out=plugins=grpc:./ proto\hop_tunnel.proto
```
