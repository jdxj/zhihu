package modules

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
	"zhihu/utils"

	"github.com/astaxie/beego/logs"
)

const (
	collectURLToken = iota + 1
)

const (
	FolloweeAPI    = `https://www.zhihu.com/api/v4/members/%s/followees?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
	FollowerAPI    = `https://www.zhihu.com/api/v4/members/%s/followers?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
	UserSumInfoAPI = `https://www.zhihu.com/api/v4/members/%s?include=allow_message,is_followed,is_following,is_org,is_blocking,employments,answer_count,follower_count,articles_count,gender,badge[?(type=best_answerer)].topics`
)

const (
	pauseDurationLimit = 2 * time.Second
)

func NewZhiHu(zhConfig *ZhiHuConfig, mysqlConfig *MySQLConfig) (*ZhiHu, error) {
	if zhConfig == nil {
		return nil, fmt.Errorf("invalid zhihu config")
	}
	if mysqlConfig == nil {
		return nil, fmt.Errorf("invalid mysql config")
	}

	cookies, err := utils.StringToCookies(zhConfig.Cookie, ".zhihu.com")
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

	dur, err := time.ParseDuration(zhConfig.PauseDuration)
	if err != nil {
		return nil, err
	}
	if dur < pauseDurationLimit {
		dur = pauseDurationLimit
	}

	zhiHu := &ZhiHu{
		config:        zhConfig,
		pauseDuration: dur,
		client:        client,
		dataSource:    ds,
		stop:          make(chan struct{}),
		stopFinish:    make(chan struct{}),
	}
	return zhiHu, nil
}

type ZhiHu struct {
	config        *ZhiHuConfig
	pauseDuration time.Duration

	client     *http.Client
	dataSource *DataSource

	stop       chan struct{}
	stopFinish chan struct{}
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
			ID: 0, // 从头开始?
		}
		startFolloweeURL = fmt.Sprintf(FolloweeAPI, zh.config.OwnURLToken)
		startFollowerURL = fmt.Sprintf(FollowerAPI, zh.config.OwnURLToken)
		logs.Warn("url token get gone when will save urlTokenProgress")
	} else if err != nil {
		urlToken = &URLToken{
			ID: 0, // 从头开始?
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

	// 使用 ticker 而不使用 timer 使得每次获取数据的时间近似
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
		logs.Error("get followee error: %s", err)
	}

	// 如果某个人的 "关注了" 或 "关注者" 为0, 那么需要给其初始值
	if nextURL == "" {
		nextURL = startURL
	}
	return nextURL
}

func (zh *ZhiHu) getFolloweeOrFollower(url string) (*PagingFollowee, error) {
	req, err := utils.NewRequestWithUserAgent("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := zh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	pf := &PagingFollowee{}
	decoder := json.NewDecoder(resp.Body)
	return pf, decoder.Decode(pf)
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
