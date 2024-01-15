.PHONY: docker
docker:
	@rm webook || true
	# 下面一行要注意 如果分开执行，要先 go env -w GOOS=linux，然后 go env -w GOARCH=arm（windows机器仍是amd64）
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f liyanlong/webookpro:v0.0.1
	@docker docker build -t liyanlong/webookpro:v0.0.1 .

mock:
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\service\code.go -package=svcmocks -destination=F:\GeekGoProjects\src\webookpro\internal\service\mock\code.mock.go
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\service\user.go -package=svcmocks -destination=F:\GeekGoProjects\src\webookpro\internal\service\mock\user.mock.go
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\repository\user.go -package=repomocks -destination=F:\GeekGoProjects\src\webookpro\internal\repository\mock\user.mock.go
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\repository\cache\user.go -package=cachemocks -destination=F:\GeekGoProjects\src\webookpro\internal\repository\cache\mock\user.mock.go
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\repository\cache\code.go -package=cachemocks -destination=F:\GeekGoProjects\src\webookpro\internal\repository\cache\mock\code.mock.go
	@mockgen -source=F:\GeekGoProjects\src\webookpro\internal\repository\dao\user.go -package=daomocks -destination=F:\GeekGoProjects\src\webookpro\internal\repository\dao\mock\user.mock.go


