.PHONY: docker
docker:
	@rm webook || true
	# 下面一行要注意 如果分开执行，要先 go env -w GOOS=linux，然后 go env -w GOARCH=arm（windows机器仍是amd64）
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f liyanlong/webookpro:v0.0.1
	@docker docker build -t liyanlong/webookpro:v0.0.1 .