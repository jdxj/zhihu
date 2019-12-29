package modules

import (
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
	FolloweeAPI = `https://www.zhihu.com/api/v4/members/%s/followees?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
	FollowerAPI = `https://www.zhihu.com/api/v4/members/%s/followers?include=data[*].answer_count,articles_count,gender,follower_count,is_followed,is_following,badge[?(type=best_answerer)].topics&offset=0&limit=20`
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

	zhiHu := &ZhiHu{
		config:     zhConfig,
		client:     client,
		dataSource: ds,
		stop:       make(chan struct{}),
		stopFinish: make(chan struct{}),
	}
	return zhiHu, nil
}

type ZhiHu struct {
	config     *ZhiHuConfig
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
	defer ds.db.Close()

	// 先清除一次
	if err := ds.Truncate(urlTokenTable); err != nil {
		logs.Error("%s", err)
		return
	}

	if err := ds.InsertURLTokens([]string{zh.config.OwnURLToken}); err != nil {
		logs.Error("%s", err)
		return
	}

	var offset int
	for {
		select {
		case <-zh.stop:
			logs.Info("stop collect url token")
			return
		default:
		}

		// todo: 测试, 先抓一个人的 "关注了" 和 "关注者"
		if offset >= 3 {
			return
		}

		urlToken, err := ds.GetURLToken(offset)
		if err != nil {
			logs.Error("%s", err)
			return
		}

		followeeFirstURLToken := fmt.Sprintf(FolloweeAPI, urlToken)
		followerFirstURLToken := fmt.Sprintf(FollowerAPI, urlToken)
		zh.continueGetFollowee(followeeFirstURLToken, followerFirstURLToken)

		offset++
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

func (zh *ZhiHu) continueGetFollowee(followeeFirstURLToken, followerFirstURLToken string) {
	var pf *PagingFollowee
	var err error
	var nextURL string

	// 遍历 "关注了"
	for pf, err = zh.getFolloweeOrFollower(followeeFirstURLToken); err == nil && len(pf.Data) != 0; pf, err = zh.getFolloweeOrFollower(nextURL) {
		logs.Debug("current followee data len: %d", len(pf.Data))

		urlTokens := make([]string, 0)
		for _, follow := range pf.Data {
			urlTokens = append(urlTokens, follow.URLToken)
		}
		logs.Debug("followee data: %v", urlTokens)

		if err := zh.dataSource.InsertURLTokens(urlTokens); err != nil {
			logs.Error("insert followee error: %s", err)
		}

		nextURL = pf.Paging.Next
	}
	if err != nil {
		logs.Error("get followee error: %s", err)
	}

	// 遍历 "关注者"
	for pf, err = zh.getFolloweeOrFollower(followerFirstURLToken); err == nil && len(pf.Data) != 0; pf, err = zh.getFolloweeOrFollower(nextURL) {
		logs.Debug("current follower data len: %d", len(pf.Data))

		urlTokens := make([]string, 0)
		for _, follow := range pf.Data {
			urlTokens = append(urlTokens, follow.URLToken)
		}
		logs.Debug("follower data: %v", urlTokens)

		if err := zh.dataSource.InsertURLTokens(urlTokens); err != nil {
			logs.Error("insert follower error: %s", err)
		}

		nextURL = pf.Paging.Next
	}
	if err != nil {
		logs.Error("get follower error: %s", err)
	}
}

func (zh *ZhiHu) getFolloweeOrFollower(url string) (*PagingFollowee, error) {
	// 避免过快
	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()

	select {
	case <-zh.stop:
		return nil, fmt.Errorf("stop get followee or follower")
	case <-timer.C:
	}

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
	IsFollowing       bool           `json:"is_following"`
	IsFollowed        bool           `json:"is_followed"`
	FollowerCount     int            `json:"follower_count"`
	AnswerCount       int            `json:"answer_count"`
	ArticlesCount     int            `json:"articles_count"`
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

type SelfRecommend struct {
}
