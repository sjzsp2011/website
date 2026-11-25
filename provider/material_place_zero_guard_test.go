package provider

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/request"
)

// 素材更新应保留未传字段，并允许显式写入零值。

// TestSaveMaterialKeepsUnpassedFields 复现「改标题把正文/分类/状态清空」。
func TestSaveMaterialKeepsUnpassedFields(t *testing.T) {
	w := newZeroValueTestSite(t)
	suffix := time.Now().UnixNano()
	title := fmt.Sprintf("零值素材-%d", suffix)
	content := fmt.Sprintf("素材原始正文-%d", suffix)

	created, err := w.SaveMaterial(&request.PluginMaterial{
		Title:      library.Ptr(title),
		CategoryId: library.Ptr(uint(3)),
		Content:    library.Ptr(content),
		Status:     library.Ptr(uint(1)),
		AutoUpdate: library.Ptr(uint(1)),
	})
	if err != nil {
		t.Skipf("测试库无 materials 表或写入失败，跳过：%v", err)
	}
	defer func() {
		_ = w.DB.Unscoped().Delete(&model.Material{}, created.Id).Error
	}()

	if created.Content != content || created.CategoryId != 3 || created.Status != 1 {
		t.Fatalf("新建素材字段未落库：content=%q category_id=%d status=%d", created.Content, created.CategoryId, created.Status)
	}

	// 只改标题（PATCH：AI/MCP 通道的默认形态）。
	if _, err = w.SaveMaterial(&request.PluginMaterial{
		Id:      created.Id,
		Title:   library.Ptr(title + "-改"),
		Partial: true,
	}); err != nil {
		t.Fatalf("PATCH 素材失败：%v", err)
	}

	var patched model.Material
	if err := w.DB.Where("`id` = ?", created.Id).Take(&patched).Error; err != nil {
		t.Fatalf("重查素材失败：%v", err)
	}
	if patched.Title != title+"-改" {
		t.Errorf("title 应为 %q，实际 %q", title+"-改", patched.Title)
	}
	if patched.Content != content {
		t.Errorf("未传 content 却被清空：期望 %q，实际 %q（正文丢失，引用它的文档下次自动更新时会变成空白）", content, patched.Content)
	}
	if patched.CategoryId != 3 {
		t.Errorf("未传 category_id 却被改写：期望 3，实际 %d（素材从分类里消失）", patched.CategoryId)
	}
	if patched.Status != 1 {
		t.Errorf("未传 status 却被改写：期望 1（启用），实际 %d", patched.Status)
	}
	if patched.AutoUpdate != 1 {
		t.Errorf("未传 auto_update 却被改写：期望 1，实际 %d（自动同步静默关闭）", patched.AutoUpdate)
	}
}

// TestSaveMaterialStatusCanBeDisabled 复现「素材永远无法禁用」。
//
// 原实现无条件 `material.Status = 1`，请求里的 status 被完全忽略。
func TestSaveMaterialStatusCanBeDisabled(t *testing.T) {
	w := newZeroValueTestSite(t)
	suffix := time.Now().UnixNano()

	created, err := w.SaveMaterial(&request.PluginMaterial{
		Title:   library.Ptr(fmt.Sprintf("待停用素材-%d", suffix)),
		Content: library.Ptr(fmt.Sprintf("正文-%d", suffix)),
	})
	if err != nil {
		t.Skipf("测试库无 materials 表或写入失败，跳过：%v", err)
	}
	defer func() {
		_ = w.DB.Unscoped().Delete(&model.Material{}, created.Id).Error
	}()

	if _, err = w.SaveMaterial(&request.PluginMaterial{
		Id:      created.Id,
		Status:  library.Ptr(uint(0)),
		Partial: true,
	}); err != nil {
		t.Fatalf("停用素材失败：%v", err)
	}

	var patched model.Material
	if err := w.DB.Where("`id` = ?", created.Id).Take(&patched).Error; err != nil {
		t.Fatalf("重查素材失败：%v", err)
	}
	if patched.Status != 0 {
		t.Errorf("显式传 status=0 应写入 0（禁用），实际 %d —— 原实现硬编码 Status=1，素材永远无法禁用", patched.Status)
	}
}

// TestSaveMaterialMd5FollowsNewContent 复现「md5 存的是上一次内容的摘要」。
//
// 原顺序：`md5Str := library.Md5(material.Content)`（此时还是旧内容，新建时为 ""），
// 之后才 `material.Content = req.Content`。所以每条素材的 md5 都滞后一次，
// 新建素材的 md5 恒为 md5("") = d41d8cd98f00b204e9800998ecf8427e。
func TestSaveMaterialMd5FollowsNewContent(t *testing.T) {
	w := newZeroValueTestSite(t)
	suffix := time.Now().UnixNano()
	first := fmt.Sprintf("第一版正文-%d", suffix)
	second := fmt.Sprintf("第二版正文-%d", suffix)

	created, err := w.SaveMaterial(&request.PluginMaterial{
		Title:   library.Ptr(fmt.Sprintf("md5素材-%d", suffix)),
		Content: library.Ptr(first),
	})
	if err != nil {
		t.Skipf("测试库无 materials 表或写入失败，跳过：%v", err)
	}
	defer func() {
		_ = w.DB.Unscoped().Delete(&model.Material{}, created.Id).Error
	}()

	if created.Md5 != library.Md5(first) {
		t.Errorf("新建素材的 md5 应是当前内容的摘要：期望 %s（md5 of %q），实际 %s（若为 md5(\"\") 说明取的是旧值）",
			library.Md5(first), first, created.Md5)
	}

	if _, err = w.SaveMaterial(&request.PluginMaterial{
		Id:      created.Id,
		Content: library.Ptr(second),
		Partial: true,
	}); err != nil {
		t.Fatalf("PATCH 素材内容失败：%v", err)
	}

	var patched model.Material
	if err := w.DB.Where("`id` = ?", created.Id).Take(&patched).Error; err != nil {
		t.Fatalf("重查素材失败：%v", err)
	}
	if patched.Content != second {
		t.Fatalf("content 应为 %q，实际 %q", second, patched.Content)
	}
	if patched.Md5 != library.Md5(second) {
		t.Errorf("md5 应跟随新内容：期望 %s，实际 %s（滞后一次会让按 md5 的去重与「内容是否变过」全部错位）",
			library.Md5(second), patched.Md5)
	}
}

// TestSaveMaterialCategoryStatusCanBeDisabled 复现「素材分类永远无法停用」。
func TestSaveMaterialCategoryStatusCanBeDisabled(t *testing.T) {
	w := newZeroValueTestSite(t)
	suffix := time.Now().UnixNano()

	category, err := w.SaveMaterialCategory(&request.PluginMaterialCategory{
		Title: library.Ptr(fmt.Sprintf("待停用素材分类-%d", suffix)),
	})
	if err != nil {
		t.Skipf("测试库无 material_categories 表或写入失败，跳过：%v", err)
	}
	defer func() {
		_ = w.DB.Unscoped().Delete(&model.MaterialCategory{}, category.Id).Error
	}()

	if category.Status != 1 {
		t.Errorf("新建素材分类默认应启用（status=1），实际 %d", category.Status)
	}

	if _, err = w.SaveMaterialCategory(&request.PluginMaterialCategory{
		Id:      category.Id,
		Status:  library.Ptr(uint(0)),
		Partial: true,
	}); err != nil {
		t.Fatalf("停用素材分类失败：%v", err)
	}

	var patched model.MaterialCategory
	if err := w.DB.Where("`id` = ?", category.Id).Take(&patched).Error; err != nil {
		t.Fatalf("重查素材分类失败：%v", err)
	}
	if patched.Status != 0 {
		t.Errorf("显式传 status=0 应写入 0，实际 %d —— 原实现硬编码 Status=1，分类永远无法停用", patched.Status)
	}
	if patched.Title != category.Title {
		t.Errorf("未传 title 却被清空：期望 %q，实际 %q", category.Title, patched.Title)
	}
}

// TestMaterialRequestFieldsArePointers 锁住请求结构体的指针性。
//
// 这几个字段必须是指针，否则「未传」与「传零值」无法区分；
// Id 必须保持值类型（0 = 新建）。
func TestMaterialRequestFieldsArePointers(t *testing.T) {
	cases := []struct {
		value    any
		mustPtr  []string
		mustKeep []string
	}{
		{
			request.PluginMaterial{},
			[]string{"Title", "CategoryId", "Content", "Status", "AutoUpdate"},
			[]string{"Id"},
		},
		{
			request.PluginMaterialCategory{},
			[]string{"Title", "Status"},
			[]string{"Id"},
		},
	}
	for _, c := range cases {
		typ := reflect.TypeOf(c.value)
		for _, name := range c.mustPtr {
			f, ok := typ.FieldByName(name)
			if !ok {
				t.Errorf("%s 缺少字段 %s", typ.Name(), name)
				continue
			}
			if f.Type.Kind() != reflect.Ptr {
				t.Errorf("%s.%s 必须是指针（nil=未传），实际 %s", typ.Name(), name, f.Type)
			}
		}
		for _, name := range c.mustKeep {
			f, ok := typ.FieldByName(name)
			if !ok {
				t.Errorf("%s 缺少字段 %s", typ.Name(), name)
				continue
			}
			if f.Type.Kind() == reflect.Ptr {
				t.Errorf("%s.%s 不应是指针，实际 %s", typ.Name(), name, f.Type)
			}
		}
	}
}
