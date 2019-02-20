dist:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/cron2_linux_amd64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/cron2_darwin_amd64