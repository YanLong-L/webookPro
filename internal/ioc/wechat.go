package ioc

import (
	"webookpro/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Oauth2Service {
	//appId, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_ID ")
	//}
	//appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic("没有找到环境变量 WECHAT_APP_SECRET")
	//}
	appId, appKey := "", ""
	return wechat.NewOauth2WeChatService(appId, appKey)
}
