package provider

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"gorm.io/gorm/clause"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/request"
)

type AnqiAiRequest struct {
	Keyword    string `json:"keyword"`
	Demand     string `json:"demand"`
	Prompt     string `json:"prompt"`
	Language   string `json:"language"`
	ToLanguage string `json:"to_language"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	TextLength int64  `json:"text_length"`
	AiRemain   int64  `json:"ai_remain"`
	Async      bool   `json:"async"`      // 是否异步处理
	Cost       bool   `json:"cost"`       // 支付需要支付
	ArticleId  int64  `json:"article_id"` // 本地的文档ID

	Type  int   `json:"type"`
	ReqId int64 `json:"req_id"`
}

type AnqiTranslateTextRequest struct {
	Language   string   `json:"language"`
	ToLanguage string   `json:"to_language"`
	Text       []string `json:"text"`   // 需要翻译的文本。可以是字符串，也可以是HTML文章
	Usage      int64    `json:"usage"`  // 消耗Token
	Status     int      `json:"status"` // 返回的时候包含
	UseSelf    bool     `json:"-"`
	Uri        string   `json:"uri"`
	Remark     string   `json:"remark"`
	Count      int64    `json:"count"` // 总量

	OriginTitle   string `json:"origin_title"`
	OriginContent string `json:"origin_content"`
}

// AnqiAiResult 是结合了2中结构体的内容，一种是plan，一种是article
type AnqiAiResult struct {
	Id          int64  `json:"id"`
	CreatedTime int64  `json:"created_time"`
	Type        int    `json:"type"`
	Language    string `json:"language"`
	ToLanguage  string `json:"to_language"`
	Keyword     string `json:"keyword"`
	Demand      string `json:"demand"`
	PayCount    int64  `json:"pay_count"`
	Status      int    `json:"status"`

	Title         string `json:"title"`
	Content       string `json:"content"`
	ReqId         int64  `json:"req_id"`
	ArticleId     int64  `json:"-"`
	UseSelf       bool   `json:"-"`
	OriginTitle   string `json:"origin_title"`
	OriginContent string `json:"origin_content"`
}

type AnqiTranslateHtmlResult struct {
	Uri        string `json:"uri"` // 这个一般不用传
	Html       string `json:"html"`
	Language   string `json:"language"`
	ToLanguage string `json:"to_language"`
	Status     int    `json:"status"`
	Count      int64  `json:"count"` // 总量
	Usage      int64  `json:"usage"` // 用量
	Remark     string `json:"remark"`
}

// AnqiTranslateString 翻译任意文本内容
func (w *Website) AnqiTranslateString(req *AnqiTranslateTextRequest) (result *AnqiTranslateTextRequest, err error) {
	if req.Language == "" {
		req.Language = w.System.Language
	}
	if req.ToLanguage == "" {
		return nil, errors.New(w.Tr("PleaseSelectTargetLanguage"))
	}
	// 如果不是正常的文本，如读取的是二进制文件，如图片内容等，则不进行翻译
	if !utf8.ValidString(strings.Join(req.Text, "")) {
		return req, nil
	}
	if w.TranslationSettings().Engine == config.TranslateEngineDefault {
		return nil, errors.New(w.Tr("TranslationEngineRequired"))
	}
	result, err = w.SelfAiTranslateResult(req)
	if err != nil {
		return nil, err
	}
	if len(result.Text) < len(req.Text) {
		return nil, errors.New("text num error")
	}
	result.UseSelf = true
	AddTranslateLog(result)
	return result, nil
}

func (w *Website) AnqiAiPseudoArticle(archive *model.Archive, isDraft bool) error {
	archiveData, err := w.GetArchiveDataById(archive.Id)
	if err != nil {
		return err
	}

	req := &AnqiAiRequest{
		Title:    archive.Title,
		Content:  archiveData.Content,
		Language: w.System.Language, // 以系统语言为标准
		Async:    true,              // 异步返回结果
	}
	req, err = w.SelfAiPseudoResult(req)
	if err != nil {
		return err
	}
	archive.Title = req.Title
	archive.Description = library.ParseDescription(strings.ReplaceAll(library.StripTags(req.Content), "\n", " "), 250)
	archive.HasPseudo = 1
	tx := w.DB
	if isDraft {
		tx = tx.Model(&model.ArchiveDraft{})
	} else {
		tx = tx.Model(&model.Archive{})
	}
	err = tx.Where("id = ?", archive.Id).UpdateColumns(map[string]interface{}{
		"title":       archive.Title,
		"description": archive.Description,
		"has_pseudo":  archive.HasPseudo,
	}).Error
	// 再保存内容
	archiveData.Content = req.Content
	w.DB.Save(archiveData)
	// 添加到plan，并标记完成
	result := AnqiAiResult{
		Type:      config.AiArticleTypeAiPseudo,
		Language:  req.Language,
		Keyword:   req.Keyword,
		Demand:    req.Demand,
		Status:    config.AiArticleStatusCompleted,
		Title:     req.Title,
		Content:   req.Content,
		ArticleId: archive.Id,
	}
	_, err = w.SaveAiArticlePlan(&result, true)
	return err
}

// AnqiAiGenerateArticle 该函数尝试采用同步返回
func (w *Website) AnqiAiGenerateArticle(keyword *model.Keyword) (int, error) {
	// 检查是否生成过
	_, err := w.GetAiArticlePlanByKeyword(config.AiArticleTypeGenerate, keyword.Title)
	if err == nil {
		//log.Println("已存在于数据库", keyword.Title)
		return 1, nil
	}

	req := &AnqiAiRequest{
		Keyword:  keyword.Title,
		Language: w.System.Language, // 以系统语言为标准
		Demand:   w.AiGenerateConfig.Demand,
		Async:    true,
	}
	req, err = w.SelfAiGenerateResult(req)
	if err != nil {
		return 0, err
	}
	var content = strings.Split(req.Content, "\n")
	if w.AiGenerateConfig.InsertImage == config.CollectImageInsert && len(w.AiGenerateConfig.Images) > 0 {
		rd := rand.New(rand.NewSource(time.Now().UnixNano()))
		img := w.AiGenerateConfig.Images[rd.Intn(len(w.AiGenerateConfig.Images))]
		index := len(content) / 3
		content = append(content, "")
		copy(content[index+1:], content[index:])
		imgTag := "<img src='" + img + "' alt='" + req.Title + "' />"
		// ![新的图片](http://xxx/xxx.webp)
		if w.Content.Editor == "markdown" {
			imgTag = fmt.Sprintf("![%s](%s)", req.Title, img)
		}
		content[index] = imgTag
	}
	if w.AiGenerateConfig.InsertImage == config.CollectImageCategory {
		// 根据分类每次只取其中一张
		img := w.GetRandImageFromCategory(w.AiGenerateConfig.ImageCategoryId, keyword.Title)
		if len(img) > 0 {
			index := len(content) / 3
			content = append(content, "")
			copy(content[index+1:], content[index:])
			imgTag := "<img src='" + img + "' alt='" + req.Title + "' />"
			// ![新的图片](http://xxx/xxx.webp)
			if w.Content.Editor == "markdown" {
				imgTag = fmt.Sprintf("![%s](%s)", req.Title, img)
			}
			content[index] = imgTag
		}
	}

	categoryId := keyword.CategoryId
	if categoryId == 0 {
		if len(w.AiGenerateConfig.CategoryIds) > 0 {
			categoryId = w.AiGenerateConfig.CategoryIds[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(w.AiGenerateConfig.CategoryIds))]
		} else if w.AiGenerateConfig.CategoryId > 0 {
			categoryId = w.AiGenerateConfig.CategoryId
		}
		if categoryId == 0 {
			var category model.Category
			w.DB.Where("module_id = 1").Take(&category)
			w.AiGenerateConfig.CategoryIds = []uint{category.Id}
			categoryId = category.Id
		}
	}

	archiveReq := request.Archive{
		Title:      req.Title,
		ModuleId:   0,
		CategoryId: categoryId,
		Keywords:   library.Ptr(keyword.Title),
		Content:    strings.Join(content, "\n"),
		KeywordId:  library.Ptr(keyword.Id),
		OriginUrl:  keyword.Title,
		OriginId:   library.Ptr(config.ArchiveFromAi),
		ForceSave:  true,
	}
	if w.AiGenerateConfig.SaveType == 0 {
		archiveReq.Draft = true
	} else {
		archiveReq.Draft = false
	}
	archive, err2 := w.SaveArchive(&archiveReq)
	if err2 != nil {
		log.Println("保存AI文章出错：", archiveReq.Title, err2.Error())
		return 0, nil
	}
	//文章计数
	w.UpdateTodayArticleCount(0, 1)
	// 添加到plan，并标记完成
	result := AnqiAiResult{
		Type:      config.AiArticleTypeGenerate,
		Language:  req.Language,
		Keyword:   req.Keyword,
		Demand:    req.Demand,
		Status:    config.AiArticleStatusCompleted,
		Title:     req.Title,
		Content:   req.Content,
		ArticleId: archive.Id,
	}
	_, err = w.SaveAiArticlePlan(&result, true)
	return 1, err
}

type StreamData struct {
	Content  string `json:"content"`
	Err      string `json:"err"`
	Finished bool   `json:"finished"`
}

type AiStreamStore struct {
	mu   sync.Mutex
	list map[string]*StreamData
}

func (a *AiStreamStore) UpdateStreamData(streamId, content, err string, finished bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	data, ok := a.list[streamId]
	if !ok {
		data = &StreamData{}
		a.list[streamId] = data
	}
	data.Content = data.Content + content
	data.Err = err
	data.Finished = finished
}

func (a *AiStreamStore) LoadStreamData(streamId string) (content, err string, finished bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	data, ok := a.list[streamId]
	if !ok {
		return "", "", true
	}
	content = data.Content
	err = data.Err
	finished = data.Finished
	data.Content = ""

	return
}

func (a *AiStreamStore) DeleteStreamData(streamId string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.list, streamId)
}

var AiStreamResults = AiStreamStore{
	mu:   sync.Mutex{},
	list: map[string]*StreamData{},
}

func (w *Website) AnqiLoadStreamData(streamId string) (content, err string, finished bool) {
	return AiStreamResults.LoadStreamData(streamId)
}

func (w *Website) AnqiAiGenerateStream(keyword *request.KeywordRequest) (string, error) {
	req := &AnqiAiRequest{
		Keyword:  keyword.Title,
		Demand:   keyword.Demand,
		Language: w.System.Language, // 以系统语言为标准
	}

	streamId := strconv.FormatInt(time.Now().UnixMilli(), 10)
	prompt := "以\"" + req.Keyword + "\"为题生成一篇SEO文章。 要求如下: 1.充分理解标题的意思,为文章确定一个主题;2.文章字数1000-1500字,避免冗长,追求表达清晰; 3.自然引用,无明显痕迹;逻辑严谨,内容连贯无歧义; 4.关键词恰当融入,避免堆砌; 5.文章需要一个标题，标题放在第一行; 6.重点内容采用加粗、斜体等标记并且确保文章原创度高于90%;7.文章无需结束语。"
	if strings.HasPrefix(req.Language, config.LanguageEn) || strings.HasPrefix(w.AiGenerateConfig.Language, config.LanguageEn) {
		prompt = "Generate an SEO article titled '" + req.Keyword + "'. The requirements are as follows: 1. Fully understand the meaning of the title and determine a theme for the article; 2. The word count of the article should be 1000-1500 words, avoiding being lengthy and pursuing clear expression; 3. Natural citation without obvious traces; Rigorous logic, coherent and unambiguous content; 4. Incorporate keywords appropriately and avoid piling them up; 5. The article needs a title, which should be placed on the first line; 6. Key content should be marked in bold, italics, etc., and the originality of the article should be ensured to be above 90%; 8. The article does not require a conclusion."
	}
	if len(req.Demand) > 0 {
		prompt += "\n" + req.Demand
	}
	if w.AiGenerateConfig.AiEngine == config.AiEngineOpenAI {
		if !w.AiGenerateConfig.ApiValid {
			return "", errors.New(w.Tr("InterfaceUnavailable"))
		}
		key := w.GetOpenAIKey()
		if key == "" {
			return "", errors.New(w.Tr("NoAvailableKey"))
		}
		stream, err := w.GetOpenAIStreamResponse(key, prompt)
		if err != nil {
			msg := err.Error()
			re, _ := regexp.Compile(`code: (\d+),`)
			match := re.FindStringSubmatch(msg)
			if len(match) > 1 {
				if match[1] == "401" || match[1] == "429" {
					// Key 已失效
					w.SetOpenAIKeyInvalid(key)
				}
			}
			return "", err
		}
		go func() {
			defer stream.Close()
			for {
				resp, err2 := stream.Recv()
				if errors.Is(err2, io.EOF) {
					break
				}
				if err2 != nil {
					err = err2
					fmt.Printf("\nStream error: %v\n", err2)
					break
				}
				AiStreamResults.UpdateStreamData(streamId, resp.Choices[0].Delta.Content, "", false)
			}
			if err != nil {
				if strings.Contains(err.Error(), "You exceeded your current quota") {
					w.SetOpenAIKeyInvalid(key)
				}

				AiStreamResults.UpdateStreamData(streamId, "", err.Error(), true)
			} else {
				AiStreamResults.UpdateStreamData(streamId, "", "", true)
			}

			time.AfterFunc(5*time.Second, func() {
				AiStreamResults.DeleteStreamData(streamId)
			})
		}()
	} else if w.AiGenerateConfig.AiEngine == config.AiEngineSpark {
		buf, _, err := GetSparkStream(w.AiGenerateConfig.Spark, prompt)
		if err != nil {
			return "", err
		}
		go func() {
			for {
				line := <-buf

				if line == "EOF" {
					break
				}

				AiStreamResults.UpdateStreamData(streamId, line, "", false)
			}
			AiStreamResults.UpdateStreamData(streamId, "", "", true)

			time.AfterFunc(5*time.Second, func() {
				AiStreamResults.DeleteStreamData(streamId)
			})
		}()
	} else {
		// 错误
		return "", errors.New(w.Tr("NoAiGenerationSourceSelected"))
	}
	return streamId, nil
}

func AddTranslateLog(result *AnqiTranslateTextRequest) {
	if len(result.Text) == 0 {
		return
	}
	// translate log 只记录到主站点的库里
	db := GetDefaultDB()
	title := result.Text[0]
	if utf8.RuneCountInString(title) > 190 {
		title = string([]rune(title)[:190])
	}
	// md5 的值 = 原始字符串+翻译的语言
	md5Str := library.Md5(fmt.Sprint(result.Text) + "-" + result.ToLanguage)
	logData := model.TranslateLog{
		Language:    result.Language,
		ToLanguage:  result.ToLanguage,
		OriginTitle: title,
		Md5:         md5Str,
	}

	db.Model(&model.TranslateLog{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "md5"}},
		UpdateAll: true,
	}).Where("`md5` = ?", md5Str).Create(&logData)
}

func AddTranslateHtmlLog(result *AnqiTranslateHtmlResult) {
	// translate log 只记录到主站点的库里
	db := GetDefaultDB()
	if len(result.Uri) > 190 {
		result.Uri = result.Uri[:190]
	}
	logData := model.TranslateHtmlLog{
		Uri:        result.Uri,
		Language:   result.Language,
		ToLanguage: result.ToLanguage,
		Count:      result.Count,
		UseCount:   result.Usage,
		Status:     result.Status,
		Remark:     result.Remark,
	}

	db.Model(&model.TranslateHtmlLog{}).Create(&logData)
}

// Restart first need to stop iris. so it will call after iris shutdown complete.
func Restart() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	// if the file is old
	if strings.HasSuffix(self, ".old") {
		self = strings.TrimRight(self, ".old")
	}
	args := os.Args
	env := os.Environ()
	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		err = cmd.Run()
		if err == nil {
			os.Exit(0)
		}
		return err
	}
	return syscall.Exec(self, args, env)
}

func Shutdown() {
	sites := GetWebsites()
	for _, w := range sites {
		// 关闭数据库
		if w.DB != nil {
			db, err := w.DB.DB()
			if err == nil {
				_ = db.Close()
			}
		}
		// 关闭日志文件
		if w.StatisticLog != nil {
			w.StatisticLog.Close()
		}
	}
}
