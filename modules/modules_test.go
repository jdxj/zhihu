package modules

import (
	"fmt"
	"net/url"
	"testing"
)

const configPath = "/home/jdxj/workspace/zhihu/config.json"

func TestReadConfig(t *testing.T) {
	config, err := ReadConfig(configPath)
	if err != nil {
		t.Fatalf("%s", err)
	}

	fmt.Printf("%+v", config)
}

func getTestConfig() (*Config, error) {
	return ReadConfig(configPath)
}

func TestDataSource(t *testing.T) {
	config, err := getTestConfig()
	if err != nil {
		t.Fatalf("%s", err)
	}

	ds, err := NewDataSource(config.MySQL)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer ds.db.Close()

	urlTokens := []string{
		"abc",
		"def",
	}
	if err := ds.InsertURLTokens(urlTokens); err != nil {
		t.Fatalf("%s", err)
	}

	urlToken, err := ds.GetURLToken(0)
	if err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Println(urlToken)

	if err := ds.Truncate(urlTokenTable); err != nil {
		t.Fatalf("%s", err)
	}
}

func TestNewZhiHu(t *testing.T) {
	config, err := getTestConfig()
	if err != nil {
		t.Fatalf("%s", err)
	}

	zh, err := NewZhiHu(config.ZhiHu, config.MySQL)
	if err != nil {
		t.Fatalf("%s", err)
	}

	pf, err := zh.getFolloweeOrFollower(fmt.Sprintf(FolloweeAPI, config.ZhiHu.OwnURLToken))
	if err != nil {
		t.Fatalf("%s", err)
	}

	fmt.Printf("next: %s\n", pf.Paging.Next)
	fmt.Printf("totals: %d\n", pf.Paging.Totals)
}

func TestURLParse(t *testing.T) {
	rawURL := "https://www.zhihu.com/api/v4/members/wang-you-qiang-36/followees?include=data%5B*%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F(type%3Dbest_answerer)%5D.topics&offset=0&limit=20"
	rawURL2 := "https://www.zhihu.com/api/v4/members/wang-you-qiang-36/followers?include=data%5B*%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F(type%3Dbest_answerer)%5D.topics&offset=0&limit=20"
	str, err := url.QueryUnescape(rawURL)
	if err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Println(str)

	str, err = url.QueryUnescape(rawURL2)
	if err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Println(str)
}