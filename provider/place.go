package provider

import (
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
)

// GetPlaceTemplate 获取模板，如果检测到不继承，则停止获取
func (w *Website) GetPlaceTemplate(place *model.Place) string {
	if place == nil {
		return ""
	}

	if place.Template != "" {
		return place.Template
	}

	//查找上级
	if place.ParentId > 0 {
		parent := w.GetPlaceFromCache(place.ParentId)
		if parent != nil {
			// 如果上级存在模板，并且选择不继承，从这里阻止
			if parent.Template != "" && parent.IsInherit == 0 {
				return ""
			}
		}
		return w.GetPlaceTemplate(parent)
	}

	//不存在，则返回空
	return ""
}

func (w *Website) GetParentPlaces(parentId uint) []*model.Place {
	var places []*model.Place
	if parentId == 0 {
		return nil
	}
	for {
		place := w.GetPlaceFromCache(parentId)
		if place == nil {
			break
		}
		places = append(places, place)
		parentId = place.ParentId
	}
	// 将 places 翻转
	for i, j := 0, len(places)-1; i < j; i, j = i+1, j-1 {
		places[i], places[j] = places[j], places[i]
	}

	return places
}

func (w *Website) GetTopPlaceId(placeId uint) uint {
	if placeId == 0 {
		return 0
	}
	for {
		place := w.GetPlaceFromCache(placeId)
		if place == nil {
			break
		}
		if place.ParentId == 0 {
			break
		}
		placeId = place.ParentId
	}

	return placeId
}

func (w *Website) GetCachePlaces() []*model.Place {
	if w.DB == nil {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	var places []*model.Place

	err := w.Cache.Get("places", &places)

	if err == nil && len(places) > 0 {
		return places
	}

	err = w.DB.Model(model.Place{}).Order("sort asc").Find(&places).Error
	if err != nil {
		return nil
	}
	for i := range places {
		places[i].GetThumb(w.PluginStorage.StorageUrl, w.GetDefaultThumb(int(places[i].Id)))
	}
	placeTree := NewPlaceTree(places)
	places = placeTree.GetTree(0, "")

	if len(places) > 0 {
		_ = w.Cache.Set("places", places, 0)
	}

	return places
}

func (w *Website) GetCachePlacesByIds(ids []uint) []*model.Place {
	places := w.GetCachePlaces()
	var tmpPlaces = make([]*model.Place, 0, len(ids))
	for _, place := range places {
		if place.Link == "" {
			place.Link = w.GetUrl("place", place, 0)
		}
		for _, id := range ids {
			if place.Id == id {
				tmpPlaces = append(tmpPlaces, place)
			}
		}
	}

	return tmpPlaces
}

func (w *Website) GetSubPlaceIds(placeId uint, places []*model.Place) []uint {
	var subIds []uint
	if places == nil {
		places = w.GetCachePlaces()
	}

	for i := range places {
		if places[i].Status != config.ContentStatusOK {
			continue
		}
		if places[i].ParentId == placeId {
			subIds = append(subIds, places[i].Id)
			subIds = append(subIds, w.GetSubPlaceIds(places[i].Id, places)...)
		}
	}

	return subIds
}

func (w *Website) GetPlaceFromCache(placeId uint) *model.Place {
	if placeId == 0 {
		return nil
	}
	places := w.GetCachePlaces()
	for i := range places {
		if places[i].Id == placeId {
			if places[i].Link == "" {
				places[i].Link = w.GetUrl(PatternPlace, places[i], 0)
			}
			return places[i]
		}
	}

	return nil
}

func (w *Website) GetPlaceFromCacheByToken(urlToken string, parents ...*model.Place) *model.Place {
	// 空 token 不能参与匹配：库里只要有一条 url_token 为空的记录，
	// 任何拿空串来查的调用方都会命中它（子域名解析、路由兜底都走这里）。
	if urlToken == "" {
		return nil
	}
	places := w.GetCachePlaces()
	var parent *model.Place
	if len(parents) > 0 {
		parent = parents[0]
	}
	if parent != nil {
		for i := range places {
			if places[i].UrlToken == urlToken && parent.Id == places[i].ParentId {
				return places[i]
			}
		}
	} else {
		for i := range places {
			if places[i].Link == "" {
				places[i].Link = w.GetUrl("place", places[i], 0)
			}
			if places[i].UrlToken == urlToken {
				return places[i]
			}
		}
	}

	return nil
}

func (w *Website) GetPlacesFromCache(parentId uint, all bool) []*model.Place {
	places := w.GetCachePlaces()
	var tmpPlaces = make([]*model.Place, 0, len(places))
	for i := range places {
		if places[i].Status != config.ContentStatusOK {
			// 跳过隐藏的
			continue
		}
		if all || places[i].ParentId == parentId {
			if places[i].Link == "" {
				places[i].Link = w.GetUrl("place", places[i], 0)
			}
			tmpPlaces = append(tmpPlaces, places[i])
		}
	}

	return tmpPlaces
}
