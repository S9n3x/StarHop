# StarHop – Latency-Driven P2P Proxy Mesh
wait

# Schedule
- [x] Complete multi node registratio  
- [x] Node auto hop registration
- [x] Delayed pre heartbeat of node
- [ ] Proxy protocol adaptation
# Build
**server**
```bash
go build -ldflags="-s -w" -trimpath -o StarHop.exe .\cmd\server\main.go
```
# Develop
**grpc**
```bash
protoc --go_out=plugins=grpc:./ proto\hop_tunnel.proto
```
