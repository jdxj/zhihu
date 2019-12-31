package modules

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

const configPath = "/home/jdxj/workspace/zhihu/config.json"
const configPath2 = "/Users/okni-12/workspace/zhihu/config.json"

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

	// test insert
	//urlTokens := []string{
	//	"abc",
	//	"def",
	//}
	//if err := ds.InsertURLTokens(urlTokens); err != nil {
	//	t.Fatalf("%s", err)
	//}

	// test get urlToken offset
	//offset := uint64(5)
	//urlToken, err := ds.GetURLToken(offset)
	//if err != nil {
	//	t.Fatalf("%s", err)
	//}
	//fmt.Println("id:", urlToken.ID)
	//fmt.Println("urlToken:", urlToken.URLToken)
	//
	//offsetRet, err := ds.GetURLTokenOffset(urlToken)
	//if err != nil {
	//	t.Fatalf("%s", err)
	//}
	//if offset != offsetRet {
	//	t.Fatalf("get urlToken offset failed")
	//}

	//utp, err := ds.GetURLTokenProgress()
	//if err == sql.ErrNoRows {
	//	fmt.Println(err)
	//} else if err != nil {
	//	t.Fatalf("%s", err)
	//}
	//fmt.Printf("%+v\n", utp)

	//utp := &URLTokenProgress{
	//	URLTokenID:      3,
	//	NextFolloweeURL: "nextFolloweeURL",
	//	NextFollowerURL: "nextFollowerURL",
	//}
	//if err := ds.InsertURLTokenProgress(utp); err != nil {
	//	t.Fatalf("%s", err)
	//}

	//if err := ds.Truncate(urlTokenTable); err != nil {
	//	t.Fatalf("%s", err)
	//}

	count, err := ds.CountURLToken()
	if err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Println(count)
}

func TestNewZhiHu(t *testing.T) {
	config, err := getTestConfig()
	if err != nil {
		t.Fatalf("%s", err)
	}

	zh, err := NewZhiHu(config)
	if err != nil {
		t.Fatalf("%s", err)
	}

	//pf, err := zh.getFolloweeOrFollower(fmt.Sprintf(FolloweeAPI, config.ZhiHu.OwnURLToken))
	//if err != nil {
	//	t.Fatalf("%s", err)
	//}
	//
	//fmt.Printf("next: %s\n", pf.Paging.Next)
	//fmt.Printf("totals: %d\n", pf.Paging.Totals)
	//zh.sendURLTokenAmountRegularly()
	pf, err := zh.getFolloweeOrFollower(`http://www.zhihu.com/api/v4/members/qi-xu-42-64/followees?include=data%5B%2A%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F%28type%3Dbest_answerer%29%5D.topics&limit=20&offset=40`)
	if err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Printf("%+v\n", *pf.Paging)

	for _, data := range pf.Data {
		fmt.Printf("%+v\n", *data)
	}
}

func TestURLParse(t *testing.T) {
	rawURL1 := "https://www.zhihu.com/api/v4/members/wang-you-qiang-36/followees?include=data%5B*%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F(type%3Dbest_answerer)%5D.topics&offset=0&limit=20"
	rawURL2 := "https://www.zhihu.com/api/v4/members/wang-you-qiang-36/followers?include=data%5B*%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F(type%3Dbest_answerer)%5D.topics&offset=0&limit=20"
	rawURL3 := "https://www.zhihu.com/api/v4/members/eluosixiongmei?include=allow_message%2Cis_followed%2Cis_following%2Cis_org%2Cis_blocking%2Cemployments%2Canswer_count%2Cfollower_count%2Carticles_count%2Cgender%2Cbadge%5B%3F(type%3Dbest_answerer)%5D.topics"

	unescaped := urlQueryUnescape(rawURL1)
	fmt.Println("rawURL1", unescaped)
	unescaped = urlQueryUnescape(rawURL2)
	fmt.Println("rawURL2", unescaped)
	unescaped = urlQueryUnescape(rawURL3)
	fmt.Println("rawURL3", unescaped)
}

func urlQueryUnescape(rawURL string) string {
	result, err := url.QueryUnescape(rawURL)
	if err != nil {
		panic(err)
	}
	return result
}

func TestEmailSender_SendEmail(t *testing.T) {
	config, err := getTestConfig()
	if err != nil {
		t.Fatalf("%s", err)
	}
	ec := config.Email

	es, err := NewEmailSender(ec)
	if err != nil {
		t.Fatalf("%s", err)
	}

	msg := &EmailMsg{
		Subject: "test zhihu send email",
		Content: "ok",
	}
	if err := es.SendEmail(msg); err != nil {
		t.Fatalf("%s", err)
	}
}

func TestReSlice(t *testing.T) {
	s := make([]int, 10)
	s = s[:0]
	for i := 0; i < 1000; i++ {
		s = append(s, i)
		if i%10 == 0 {
			fmt.Printf("%v\n", s)
			s = s[:0]
		}
	}
}

func TestJsonParse(t *testing.T) {
	str := `{"paging": {"is_end": true, "totals": 0, "previous": "http://www.zhihu.com/api/v4/members/eluosixiongmei/followees?include=data%5B%2A%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F%28type%3Dbest_answerer%29%5D.topics&limit=20&offset=0", "is_start": true, "next": "http://www.zhihu.com/api/v4/members/eluosixiongmei/followees?include=data%5B%2A%5D.answer_count%2Carticles_count%2Cgender%2Cfollower_count%2Cis_followed%2Cis_following%2Cbadge%5B%3F%28type%3Dbest_answerer%29%5D.topics&limit=20&offset=20"}, "data": []}`

	pf := &PagingFollowee{}
	if err := json.Unmarshal([]byte(str), pf); err != nil {
		t.Fatalf("%s", err)
	}
}
