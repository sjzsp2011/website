package provider

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
)

func (w *Website) GetUserByFunc(ops func(tx *gorm.DB) *gorm.DB) (*model.User, error) {
	var user model.User
	err := ops(w.DB).Take(&user).Error
	if err != nil {
		return nil, err
	}
	user.GetThumb(w.PluginStorage.StorageUrl)
	user.Link = w.GetUrl(PatternPeople, &user, 0)
	user.Extra = w.GetUserExtra(user.Id)
	return &user, nil
}

func (w *Website) GetUserInfoById(userId uint) (*model.User, error) {
	if userId == 0 {
		return nil, errors.New("no user")
	}
	return w.GetUserByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`id` = ?", userId)
	})
}

func (w *Website) GetUserInfoByUrlToken(urlToken string) (*model.User, error) {
	return w.GetUserByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`url_token` = ?", urlToken)
	})
}

func (w *Website) GetUserInfoByEmail(email string) (*model.User, error) {
	return w.GetUserByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`email` = ?", email)
	})
}

// ── 用户名唯一性（2026-10-27）────────────────────────────────────────────
//
// user_name 不只是展示名：LoginViaPassword 有一条「按 user_name 登录」的分支，
// 两个同名账号会让人登进别人那里。所以：
//   - 后台/AI 创建（SaveUserInfo，用户名由调用方自选）→ 重名直接拒绝，不改写；
//   - 第三方自动注册（Google/公众号/小程序，用户名取自平台昵称，用户无法自选）
//     → 先查重，重复时追加数字后缀：张三、张三2、张三3……
//
// 之所以两条路径策略不同：前者用户能改，报错让他换一个就行；后者他改不了昵称，
// 拒绝注册等于把第三方登录整个堵死。
const (
	// userNameMaxLen 与 model.User.UserName 的 varchar(64) 对齐。
	// 后缀最多 4 位（2..9999），所以 base 压到 60 个字符即可保证总长不越界。
	userNameMaxLen = 60
	// defaultThirdPartyUserName 是平台没给昵称时的兜底名。
	// 允许空用户名会让「按用户名登录」命中任意一条空名记录，必须堵掉。
	defaultThirdPartyUserName = "用户"
)

func (w *Website) GetUserGroupInfo(groupId uint) (*model.UserGroup, error) {
	var group model.UserGroup

	err := w.DB.Where("`id` = ?", groupId).Take(&group).Error

	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (w *Website) GetUserWechatByOpenid(openid string) (*model.UserWechat, error) {
	var userWechat model.UserWechat
	if err := w.DB.Where("`openid` = ?", openid).First(&userWechat).Error; err != nil {
		return nil, err
	}

	return &userWechat, nil
}

func (w *Website) VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// GetUserDiscount 获取用户优惠比例， 优先级：用户组优惠比例 > 通过分享链接下单优惠比例
func (w *Website) GetUserDiscount(userId uint, user *model.User) (discount, shippingDiscount int64) {
	if user == nil && userId > 0 {
		user, _ = w.GetUserInfoById(userId)
	}
	if user != nil {
		group, err := w.GetUserGroupInfo(user.GroupId)
		if err == nil {
			if group.Setting.Discount > 0 {
				discount = group.Setting.Discount
			}
			if group.Setting.ShippingDiscount > 0 {
				shippingDiscount = group.Setting.ShippingDiscount
			}
		}
		if user.ParentId > 0 {
			parent, err := w.GetUserInfoById(user.ParentId)
			if err == nil {
				group, err := w.GetUserGroupInfo(parent.GroupId)
				if err == nil {
					if group.Setting.ShareDiscount > 0 && group.Setting.ShareDiscount > discount {
						discount = group.Setting.ShareDiscount
					}
				}
			}
		}
	}

	return discount, shippingDiscount
}

func (w *Website) GetUserFields() []*config.CustomField {
	//这里有默认的设置
	fields := w.PluginUser.Fields

	return fields
}

func (w *Website) MigrateUserTable(fields []*config.CustomField, focus bool) {
	tableName := "users"
	// 根据表单字段，生成数据
	for _, field := range fields {
		field.CheckSetFilter()
		column := field.GetFieldColumn()
		if !w.DB.Migrator().HasColumn(model.User{}, field.FieldName) {
			//创建语句
			w.DB.Exec("ALTER TABLE ? ADD COLUMN ?", clause.Table{Name: tableName}, gorm.Expr(column))
		} else if focus {
			//更新语句
			w.DB.Exec("ALTER TABLE ? MODIFY COLUMN ?", clause.Table{Name: tableName}, gorm.Expr(column))
		}
	}
}

func (w *Website) GetUserExtra(id uint) map[string]*config.CustomField {
	//读取extra
	result := map[string]interface{}{}
	// 从数据库中取出来
	extraFields := map[string]*config.CustomField{}
	var fields []string
	for _, v := range w.PluginUser.Fields {
		fields = append(fields, "`"+v.FieldName+"`")
	}
	//从数据库中取出来
	if len(fields) > 0 {
		w.DB.Model(model.User{}).Where("`id` = ?", id).Select(strings.Join(fields, ",")).Scan(&result)
		//extra的CheckBox的值
		for _, v := range w.PluginUser.Fields {
			value, ok := result[v.FieldName].(string)
			if ok {
				if v.Type == config.CustomFieldTypeImage || v.Type == config.CustomFieldTypeFile || v.Type == config.CustomFieldTypeEditor {
					result[v.FieldName] = w.ReplaceContentUrl(value, true)
				} else if v.Type == config.CustomFieldTypeImages {
					// json 还原
					var images []string
					err := json.Unmarshal([]byte(value), &images)
					if err == nil {
						for i := range images {
							images[i] = w.ReplaceContentUrl(images[i], true)
						}
						result[v.FieldName] = images
					}
				} else if v.Type == config.CustomFieldTypeTexts {
					var texts []config.CustomFieldTexts
					err := json.Unmarshal([]byte(value), &texts)
					if err == nil {
						result[v.FieldName] = texts
					}
				} else if v.Type == config.CustomFieldTypeTimeline {
					var val config.TimelineField
					_ = json.Unmarshal([]byte(value), &val)
					result[v.FieldName] = val
				} else if v.Type == config.CustomFieldTypeArchive {
					var arcIds []int64
					err := json.Unmarshal([]byte(value), &arcIds)
					if err == nil {
						result[v.FieldName] = arcIds
					}
				} else if v.Type == config.CustomFieldTypeCategory || v.Type == config.CustomFieldTypeNumber {
					result[v.FieldName], _ = strconv.ParseInt(value, 10, 64)
				}
			}
			extraFields[v.FieldName] = &config.CustomField{
				Name:        v.Name,
				Value:       result[v.FieldName],
				Default:     v.Content,
				FollowLevel: v.FollowLevel,
			}
		}
	}

	return extraFields
}
