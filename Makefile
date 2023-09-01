all:
	go build -o gin-ws.server ./gin-upgrade/server/server.go
	go build -o gin-ws.client ./gin-upgrade/client/client.go
	go build -o net-http-ws.server ./net-http-upgrade/server/server.go
	go build -o net-http-ws.client ./net-http-upgrade/client/client.go
