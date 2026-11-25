package request

// PluginGoogleAuthRequest Google OAuth 回调请求。
type PluginGoogleAuthRequest struct {
	Code  string `json:"code"`  // 授权码，Google 回调带回
	State string `json:"state"` // 防 CSRF 的 state，必须与 auth/url 返回的一致
}
