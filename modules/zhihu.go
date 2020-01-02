package modules

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
	"zhihu/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/logs"
)

const (
	collectURLToken = iota + 1
	collectTopicID
	collectTopic
)

const (
	FolloweeAPI    = `https://www.zhihu.com/api/v4/members/%s/followees?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
	FollowerAPI    = `https://www.zhihu.com/api/v4/members/%s/followers?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
	UserSumInfoAPI = `https://www.zhihu.com/api/v4/members/%s?include=allow_message,is_followed,is_following,is_org,is_blocking,employments,answer_count,follower_count,articles_count,gender,badge[?(type=best_answerer)].topics`
	TopicIDAPI     = `https://www.zhihu.com/api/v3/topics/%s/children`
)

const (
	TopicWebPage = `https://www.zhihu.com/topic/%s/hot`
)

const (
	pauseDurationLimit = 3 * time.Second
	retryCountLimit    = 5
)

func NewZhiHu(config *Config) (*ZhiHu, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config")
	}

	zhiHuConfig := config.ZhiHu
	mysqlConfig := config.MySQL
	emailConfig := config.Email

	cookies, err := utils.StringToCookies(zhiHuConfig.Cookie, ".zhihu.com")
	if err != nil {
		return nil, err
	}
	cookieURL, err := url.Parse("https://www.zhihu.com")
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	jar.SetCookies(cookieURL, cookies)

	client := &http.Client{
		Jar: jar,
	}

	ds, err := NewDataSource(mysqlConfig)
	if err != nil {
		return nil, err
	}

	dur, err := time.ParseDuration(zhiHuConfig.PauseDuration)
	if err != nil {
		return nil, err
	}
	if dur < pauseDurationLimit {
		dur = pauseDurationLimit
	}

	emailSender, err := NewEmailSender(emailConfig)
	if err != nil {
		return nil, err
	}

	zhiHu := &ZhiHu{
		config:        zhiHuConfig,
		pauseDuration: dur,
		client:        client,
		emailSender:   emailSender,
		dataSource:    ds,
		stop:          make(chan struct{}),
		stopFinish:    make(chan struct{}),
	}
	return zhiHu, nil
}

type ZhiHu struct {
	config        *ZhiHuConfig
	pauseDuration time.Duration

	client      *http.Client
	dataSource  *DataSource
	emailSender *EmailSender

	stop       chan struct{}
	stopFinish chan struct{}
}

func (zh *ZhiHu) Start() {
	switch zh.config.Mode {
	case collectURLToken:
		go zh.CollectURLToken()
		go zh.sendURLTokenAmountRegularly()
	case collectTopicID:
		go zh.CollectTopicID()
	default:
		logs.Warn("unexpected start mode")
	}
}

func (zh *ZhiHu) CollectURLToken() {
	defer func() {
		close(zh.stopFinish)
	}()

	ds := zh.dataSource
	// todo: 是否在这里关闭数据库?
	defer ds.db.Close()

	// 确保有一条数据
	if err := ds.InsertURLTokens([]string{zh.config.OwnURLToken}); err != nil {
		logs.Error("%s", err)
		return
	}

	var offset uint64
	var startFolloweeURL string
	var startFollowerURL string

	// 初始化进度
	utp, err := ds.GetURLTokenProgress()
	if err == sql.ErrNoRows { // 从未保存过记录
		offset = 0
		startFolloweeURL = fmt.Sprintf(FolloweeAPI, zh.config.OwnURLToken)
		startFollowerURL = fmt.Sprintf(FollowerAPI, zh.config.OwnURLToken)
	} else if err != nil {
		logs.Error("error when get urlTokenProgress: %s", err)
		return
	} else {
		if offset, err = ds.GetURLTokenOffset(utp.URLTokenID); err != nil {
			logs.Error("error when get urlTokenOffset: %s", err)
			return
		}

		startFolloweeURL = utp.NextFolloweeURL
		startFollowerURL = utp.NextFollowerURL

		logs.Debug("load urlTokenProgress success")
	}

loop:
	for {
		// 这里为 start* 赋值只是为了记录进度
		startFolloweeURL = zh.continueGetFollowee(startFolloweeURL)
		startFollowerURL = zh.continueGetFollowee(startFollowerURL)

		select {
		case <-zh.stop:
			logs.Info("stop collect url token, offset: %d", offset)
			break loop
		default:
		}

		offset++
		urlToken, err := ds.GetURLToken(offset)
		if err == sql.ErrNoRows {
			logs.Info("url token get gone")
			break loop
		} else if err != nil {
			logs.Error("%s", err)
			break loop
		}

		startFolloweeURL = fmt.Sprintf(FolloweeAPI, urlToken.URLToken)
		startFollowerURL = fmt.Sprintf(FollowerAPI, urlToken.URLToken)
	}

	// 保存进度
	urlToken, err := ds.GetURLToken(offset)
	if err == sql.ErrNoRows {
		urlToken = &URLToken{
			ID: 1, // 从头开始?
		}
		startFolloweeURL = fmt.Sprintf(FolloweeAPI, zh.config.OwnURLToken)
		startFollowerURL = fmt.Sprintf(FollowerAPI, zh.config.OwnURLToken)
		logs.Warn("url token get gone when will save urlTokenProgress")
	} else if err != nil {
		urlToken = &URLToken{
			ID: 1, // 从头开始?
		}
		startFolloweeURL = fmt.Sprintf(FolloweeAPI, zh.config.OwnURLToken)
		startFollowerURL = fmt.Sprintf(FollowerAPI, zh.config.OwnURLToken)
		logs.Error("url token get gone when will save urlTokenProgress: %s", err)
	}

	urlTokenProgress := &URLTokenProgress{}
	urlTokenProgress.URLTokenID = urlToken.ID
	urlTokenProgress.NextFolloweeURL = startFolloweeURL
	urlTokenProgress.NextFollowerURL = startFollowerURL
	if err := ds.InsertURLTokenProgress(urlTokenProgress); err != nil {
		logs.Error("error when insert urlTokenProgress: %s", err)
	}
}

func (zh *ZhiHu) Stop() {
	close(zh.stop)

	select {
	case <-zh.stopFinish:
	}
	logs.Info("zhihu stop finish")
}

func (zh *ZhiHu) Wait() <-chan struct{} {
	return zh.stopFinish
}

func (zh *ZhiHu) continueGetFollowee(startURL string) string {
	var pf *PagingFollowee
	var err error
	var nextURL string
	urlTokens := make([]string, 20)
	urlTokens = urlTokens[:0]

	ticker := time.NewTicker(zh.pauseDuration)
	defer ticker.Stop()

	// 如果接收到 stop 信号, 那么起码把当前 url 的数据获取完成再停止
loop:
	for pf, err = zh.getFolloweeOrFollower(startURL); err == nil && len(pf.Data) != 0; pf, err = zh.getFolloweeOrFollower(nextURL) {
		for _, follow := range pf.Data {
			urlTokens = append(urlTokens, follow.URLToken)
		}

		if err := zh.dataSource.InsertURLTokens(urlTokens); err != nil {
			logs.Error("insert followee error: %s", err)
		}
		urlTokens = urlTokens[:0]
		nextURL = pf.Paging.Next

		select {
		case <-zh.stop:
			logs.Warn("stop continue get followee or follower")
			break loop
		case <-ticker.C:
		}
	}
	if err != nil {
		logs.Error("get followee error: %s, startURL: %s, nextURL: %s", err, startURL, nextURL)

		// todo: 这是个未解决的错误, 需要在运行该程序时调试; 需要邮件通知
		msg := &EmailMsg{
			Subject: "Get Followee Error",
			Content: fmt.Sprintf("startURL: %s, nextURL: %s", startURL, nextURL),
		}
		if err := zh.emailSender.SendEmail(msg); err != nil {
			logs.Error("error when send 'Get Followee Error' email: %s", err)
		}
	}

	// 如果某个人的 "关注了" 或 "关注者" 为0, 那么需要给其初始值
	if nextURL == "" {
		nextURL = startURL
	}
	return nextURL
}

func (zh *ZhiHu) getFolloweeOrFollower(url string) (*PagingFollowee, error) {
	timer := time.NewTimer(zh.pauseDuration)
	defer timer.Stop()

	var retryCount int
	for {
		req, err := utils.NewRequestWithUserAgent("GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := zh.client.Do(req)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		pf := &PagingFollowee{}
		if err := json.Unmarshal(data, pf); err != nil {
			logs.Error("error when get followee or follower, url: %s, data: %s, err: %s", url, data, err)

			if strings.HasPrefix(err.Error(), "invalid character") {
				// 第一次也统计到重试次数中
				retryCount++
				logs.Debug("get followee or follower retry count: %d", retryCount)
				if retryCount >= retryCountLimit {
					logs.Error("retry count over")
					return nil, err
				}

				timer.Reset(zh.pauseDuration)
				select {
				case <-timer.C:
				}
			} else {
				// 其他错误
				return nil, err
			}
		} else {
			return pf, nil
		}
	}
}

func (zh *ZhiHu) sendURLTokenAmountRegularly() {
	msg := &EmailMsg{
		Subject: "URLToken Amount",
	}

	// 先发送一次数据
	count, err := zh.dataSource.CountURLToken()
	msg.Content = fmt.Sprintf("count: %d, err: %s", count, err)

	if err := zh.emailSender.SendEmail(msg); err != nil {
		logs.Error("%s", err)
	}

	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-zh.stop:
			// 立即释放
			ticker.Stop()
			return
		case <-ticker.C:
		}

		count, err := zh.dataSource.CountURLToken()
		msg.Content = fmt.Sprintf("count: %d, err: %s", count, err)

		if err := zh.emailSender.SendEmail(msg); err != nil {
			logs.Error("%s", err)
		}
	}
}

type PagingFollowee struct {
	Paging *Paging     `json:"paging"`
	Data   []*Followee `json:"data"`
}

type Paging struct {
	IsEnd    bool   `json:"is_end"`
	IsStart  bool   `json:"is_start"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Totals   int    `json:"totals"`
}

type Followee struct {
	ID                string         `json:"id"`
	URLToken          string         `json:"url_token"`
	Name              string         `json:"name"`
	UseDefaultAvatar  bool           `json:"use_default_avatar"`
	AvatarURL         string         `json:"avatar_url"`
	AvatarURLTemplate string         `json:"avatar_url_template"`
	IsOrg             bool           `json:"is_org"`
	Type              string         `json:"type"`
	URL               string         `json:"url"`
	UserType          string         `json:"user_type"`
	HeadLine          string         `json:"head_line"`
	Gender            int            `json:"gender"`
	IsAdvertiser      bool           `json:"is_advertiser"`
	VIPInfo           *VIPInfo       `json:"vip_info"`
	Badge             []*Badge       `json:"badge"`
	AllowMessage      bool           `json:"allow_message"`
	IsFollowing       bool           `json:"is_following"`
	IsFollowed        bool           `json:"is_followed"`
	IsBlocking        bool           `json:"is_blocking"`
	FollowerCount     int            `json:"follower_count"`
	AnswerCount       int            `json:"answer_count"`
	ArticlesCount     int            `json:"articles_count"`
	Employments       []*Employment  `json:"employments"`
	SelfRecommend     *SelfRecommend `json:"self_recommend"`
}

type VIPInfo struct {
	IsVIP      bool   `json:"is_vip"`
	RenameDays string `json:"rename_days"`
}

type Badge struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Employment struct {
	Job     *Job     `json:"job"`
	Company *Company `json:"company"`
}

type Job struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Company struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
type SelfRecommend struct {
}

func (zh *ZhiHu) CollectTopicID() {
	defer func() {
		close(zh.stopFinish)
	}()

	ds := zh.dataSource
	// todo: 是否在这里关闭数据库?
	defer ds.db.Close()

	rootTopicID := &TopicID{
		TopicID: zh.config.RootTopicID,
		Name:    "根话题", // 比较懒
	}
	if err := ds.InsertTopicsID([]*TopicID{rootTopicID}); err != nil {
		logs.Error("%s", err)
		return
	}

	var offset uint64
	var startTopicIDURL string

	tip, err := ds.GetTopicIDProgress()
	if err == sql.ErrNoRows {
		offset = 0
		startTopicIDURL = fmt.Sprintf(TopicIDAPI, zh.config.RootTopicID)
	} else if err != nil {
		logs.Error("%s", err)
		return
	} else {
		if offset, err = ds.GetTopicIDOffset(tip.TopicID); err != nil {
			logs.Error("error when get topicIDOffset: %s", err)
			return
		}

		startTopicIDURL = tip.NextTopicIDURL
		logs.Debug("load topicIDProgress success")
	}

loop:
	for {
		startTopicIDURL = zh.continueGetTopicID(startTopicIDURL)

		select {
		case <-zh.stop:
			logs.Info("stop collect topic id, offset: %d", offset)
			break loop
		default:
		}

		offset++
		topicID, err := ds.GetTopicID(offset)
		if err == sql.ErrNoRows {
			logs.Info("topic id get gone")
			break loop
		} else if err != nil {
			logs.Error("%s", err)
			break loop
		}

		startTopicIDURL = fmt.Sprintf(TopicIDAPI, topicID.TopicID)
	}

	topicID, err := ds.GetTopicID(offset)
	if err == sql.ErrNoRows {
		topicID = &TopicID{
			ID: 1,
		}
		startTopicIDURL = fmt.Sprintf(TopicIDAPI, zh.config.RootTopicID)
		logs.Warn("topic id get gone when will save topicIDProgress")
	} else if err != nil {
		topicID = &TopicID{
			ID: 1,
		}
		startTopicIDURL = fmt.Sprintf(TopicIDAPI, zh.config.RootTopicID)
		logs.Warn("topic id get gone when will save topicIDProgress: %s", err)
	}

	topicIDProgress := &TopicIDProgress{
		TopicID:        topicID.ID,
		NextTopicIDURL: startTopicIDURL,
	}
	if err := ds.InsertTopicIDProgress(topicIDProgress); err != nil {
		logs.Error("error when insert topicIDProgress: %s", err)
	}
}

func (zh *ZhiHu) continueGetTopicID(startTopicIDURL string) string {
	var pt *PagingTopic
	var err error
	var nextURL string
	topicsID := make([]*TopicID, 20)
	topicsID = topicsID[:0]

	ticker := time.NewTicker(zh.pauseDuration)
	defer ticker.Stop()

loop:
	for pt, err = zh.getTopicID(startTopicIDURL); err == nil && len(pt.Data) != 0; pt, err = zh.getTopicID(nextURL) {
		for _, topic := range pt.Data {
			topicID := &TopicID{
				TopicID: topic.ID,
				Name:    topic.Name,
			}
			topicsID = append(topicsID, topicID)
		}

		if err := zh.dataSource.InsertTopicsID(topicsID); err != nil {
			logs.Error("%s", err)
		}
		topicsID = topicsID[0:]
		nextURL = pt.Paging.Next

		select {
		case <-zh.stop:
			logs.Warn("stop continue get topicID")
			break loop
		case <-ticker.C:
		}
	}
	if err != nil {
		logs.Error("get topicID error: %s", err)
	}

	if nextURL == "" {
		nextURL = startTopicIDURL
	}
	return nextURL
}

func (zh *ZhiHu) getTopicID(topicIDURL string) (*PagingTopic, error) {
	timer := time.NewTimer(zh.pauseDuration)
	defer timer.Stop()

	var retryCount int
	for {
		req, err := utils.NewRequestWithUserAgent("GET", topicIDURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := zh.client.Do(req)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		pt := &PagingTopic{}
		if err := json.Unmarshal(data, pt); err != nil {
			logs.Error("error when get topic id: url: %s, data: %s, err: %s", topicIDURL, data, err)

			if strings.HasPrefix(err.Error(), "invalid character") {
				retryCount++
				logs.Debug("get topic id retry count: %d: ", retryCount)
				if retryCount >= retryCountLimit {
					logs.Error("retry count over")
					return nil, err
				}

				timer.Reset(zh.pauseDuration)
				select {
				case <-timer.C:
				}
			}
		} else {
			return pt, nil
		}
	}
}

type PagingTopic struct {
	Paging *Paging  `json:"paging"`
	Data   []*Topic `json:"data"`
}

type Topic struct {
	IsBlack      bool   `json:"is_black"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Excerpt      string `json:"excerpt"`
	IsVote       bool   `json:"is_vote"`
	Introduction string `json:"introduction"`
	AvatarURL    string `json:"avatar_url"`
	Type         string `json:"type"`
	ID           string `json:"id"`
}

func (zh *ZhiHu) CollectTopic() {
	var offset uint64
	tp, err := zh.dataSource.GetTopicProgress()
	if err == sql.ErrNoRows {
		offset = 0
	} else if err != nil {
		logs.Error("%s", err)
		return
	} else {
		if offset, err = zh.dataSource.GetTopicIDOffset(tp.TopicID); err != nil {
			logs.Error("%s", err)
			return
		}
	}

	ticker := time.NewTicker(zh.pauseDuration)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-zh.stop:
			logs.Info("stop collect topic")
			break loop
		case <-ticker.C:
		}

		ti, err := zh.dataSource.GetTopicID(offset)
		if err == sql.ErrNoRows {
			logs.Info("topic id get gone when collect topic")
			break loop
		} else if err != nil {
			logs.Error("%s", err)
			break loop
		}

		startTopicURL := fmt.Sprintf(TopicWebPage, ti.TopicID)
		if err := zh.crawlTopic(startTopicURL, ti.ID); err != nil {
			logs.Error("%s", err)
		}
		offset++
	}

	ti, err := zh.dataSource.GetTopicID(offset)
	if err == sql.ErrNoRows {
		ti = &TopicID{
			ID: 1,
		}
		logs.Warn("topic id get gone when will save topicProgress")
	} else if err != nil {
		ti = &TopicID{
			ID: 1,
		}
		logs.Warn("topic id get gone when will save topicProgress: %s", err)
	}

	topicProgress := &TopicProgress{
		TopicID: ti.ID,
	}
	if err := zh.dataSource.InsertTopicProgress(topicProgress); err != nil {
		logs.Error("error when insert topicProgress: %s", err)
	}
}

func (zh *ZhiHu) crawlTopic(webPageURL string, id uint64) error {
	// todo: 是否有重试逻辑
	req, err := utils.NewRequestWithUserAgent("GET", webPageURL, nil)
	if err != nil {
		return err
	}
	resp, err := zh.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	tt := &TopicTable{
		TopicID: id,
	}

	selection := doc.Find(".NumberBoard-itemValue")
	if value, ok := selection.First().Attr("title"); ok {
		count, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("error when strconv topic follower count, topicID: %d, title: %s, err: %s",
				id, value, err)
		}
		tt.FollowerCount = count
	}

	if value, ok := selection.Last().Attr("title"); ok {
		count, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("error when strconv topic question count, topicID: %d, title: %s, err: %s",
				id, value, err)
		}
		tt.QuestionCount = count
	}

	return zh.dataSource.InsertTopic(tt)
}
