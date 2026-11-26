package request

type TranslateArticleRequest struct {
	Id         int64  `json:"id"`          // 文档 ID
	ToLanguage string `json:"to_language"` // 目标语言，如：en
}

type AiPseudoArticleRequest struct {
	Id int64 `json:"id"` // 文档 ID
}
