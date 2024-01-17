package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"webookpro/internal/domain"
)

var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Oauth2Service interface {
	AuthURL(ctx context.Context) (string, error)
	VerifyCode(ctx *gin.Context, code string, state string) (domain.WeChatInfo, error)
}

type Oauth2WeChatService struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewOauth2WeChatService(appId string, appSecret string) *Oauth2WeChatService {
	return &Oauth2WeChatService{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

// AuthURL 构造跳转到微信扫码登录页面的url
func (s *Oauth2WeChatService) AuthURL(ctx context.Context) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	state := uuid.New()
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

// VerifyCode 微信扫码成功回调后 验证code
func (s *Oauth2WeChatService) VerifyCode(ctx *gin.Context, code string, state string) (domain.WeChatInfo, error) {
	// 如何验证呢? 其实就是用这个code去获取accesstoken，获取到了证明code正确，验证成功
	urlTpl := "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	urlStr := fmt.Sprintf(urlTpl, s.appId, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	var result Result
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	return domain.WeChatInfo{
		OpenID:  result.OpenID,
		UnionID: result.UnionID,
	}, nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenID  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionID string `json:"unionid"`
}
