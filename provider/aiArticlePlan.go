package provider

import (
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
)

func (w *Website) GetAiArticlePlanByReqId(reqId int64) (*model.AiArticlePlan, error) {
	var plan model.AiArticlePlan
	err := w.DB.Where("`req_id` = ?", reqId).Take(&plan).Error

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (w *Website) GetAiArticlePlanById(id uint) (*model.AiArticlePlan, error) {
	var plan model.AiArticlePlan
	err := w.DB.Where("`id` = ?", id).Take(&plan).Error

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (w *Website) GetAiArticlePlanByKeyword(planType int, keyword string) (*model.AiArticlePlan, error) {
	var plan model.AiArticlePlan
	err := w.DB.Where("`type` = ? and `keyword` = ? AND (use_self = ? OR status = ?)", planType, keyword, true, config.AiArticleStatusCompleted).Take(&plan).Error

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (w *Website) SaveAiArticlePlan(resp *AnqiAiResult, useSelf bool) (*model.AiArticlePlan, error) {
	if resp.ReqId > 0 {
		plan, err := w.GetAiArticlePlanByReqId(resp.ReqId)
		if err == nil {
			// 已存在
			return plan, nil
		}
	}

	plan := &model.AiArticlePlan{
		Type:       resp.Type,
		ReqId:      resp.ReqId,
		Language:   resp.Language,
		ToLanguage: resp.ToLanguage,
		Keyword:    resp.Keyword,
		Demand:     resp.Demand,
		ArticleId:  resp.ArticleId,
		PayCount:   resp.PayCount,
		UseSelf:    useSelf,
		Status:     resp.Status,
	}

	err := w.DB.Save(plan).Error
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// CancelLegacyAiArticlePlans closes pending jobs from the retired cloud writer.
func (w *Website) CancelLegacyAiArticlePlans() error {
	return w.DB.Model(&model.AiArticlePlan{}).
		Where("status IN ? AND use_self = ?", []int{config.AiArticleStatusWaiting, config.AiArticleStatusDoing}, false).
		UpdateColumn("status", config.AiArticleStatusCanceled).Error
}
