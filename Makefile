build:
	CGO_ENABLED=0 go build  -ldflags="-X 'main.Version=$$(git describe --tags --always --dirty)' -s -w" -o video2live .
docker: build
	docker build . -t shynome/video2live:$$(git describe --tags --always --dirty)
push: docker
	docker push shynome/video2live:$$(git describe --tags --always --dirty)
