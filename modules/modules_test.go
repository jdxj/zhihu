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
	return ReadConfig(configPath2)
}

var industryStr = `高新科技 互联网 电子商务 电子游戏 计算机软件 计算机硬件 信息传媒 出版业 电影录音 广播电视 通信 金融 银行 资本投资 证券投资 保险 信贷 财务 审计 服务业 法律 餐饮 酒店 旅游 广告 公关 景观 咨询分析 市场推广 人力资源 社工服务 养老服务 教育 高等教育 基础教育 职业教育 幼儿教育 特殊教育 培训 医疗服务 临床医疗 制药 保健 美容 医疗器材 生物工程 疗养服务 护理服务 艺术娱乐 创意艺术 体育健身 娱乐休闲 图书馆 博物馆 策展 博彩 制造加工 食品饮料业 纺织皮革业 服装业 烟草业 造纸业 印刷业 化工业 汽车 家具 电子电器 机械设备 塑料工业 金属加工 军火 地产建筑 房地产 装饰装潢 物业服务 特殊建造 建筑设备 贸易零售 零售 大宗交易 进出口贸易 公共服务 政府 国防军事 航天 科研 给排水 水利能源 电力电网 公共管理 环境保护 非营利组织 开采冶金 煤炭工业 石油工业 黑色金属 有色金属 土砂石开采 地热开采 交通仓储 邮政 物流递送 地面运输 铁路运输 管线运输 航运业 民用航空业 农林牧渔 种植业 畜牧养殖业 林业 渔业`

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

	ti := &TopicID{
		TopicID: "19776749",
	}
	if err := ds.InsertTopicsID([]*TopicID{ti}); err != nil {
		t.Fatalf("%s", err)
	}
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

	if err := zh.crawlPeople(3, "liaoxuefeng"); err != nil {
		t.Fatalf("%s", err)
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
	str := `{"paging": {"is_end": false, "previous": "http://www.zhihu.com/api/v3/topics/19776749/children?limit=10&offset=0", "is_start": true, "next": "http://www.zhihu.com/api/v3/topics/19776749/children?limit=10&offset=10"}, "data": [{"is_black": false, "name": "\u300c\u672a\u5f52\u7c7b\u300d\u8bdd\u9898", "url": "http://www.zhihu.com/api/v3/topics/19776751", "excerpt": "\u77e5\u4e4e\u7684\u5168\u90e8\u8bdd\u9898\u901a\u8fc7\u7236\u5b50\u5173\u7cfb\u6784\u6210\u4e00\u4e2a\u6709\u6839\u65e0\u5faa\u73af\u7684\u6709\u5411\u56fe\u3002 \u6240\u6709\u6ca1\u6709\u76f4\u63a5\u6dfb\u52a0\u7236\u8bdd\u9898\u7684\u8bdd\u9898\u4f1a\u81ea\u52a8\u6210\u4e3a\u300c\u672a\u5f52\u7c7b\u300d\u8bdd\u9898\u7684\u5b50\u8bdd\u9898\uff0c\u4ece\u800c\u4e0e\u6574\u4e2a\u8bdd\u9898\u6811\u8fde\u63a5\u8d77\u6765\u3002", "is_vote": false, "introduction": "\u77e5\u4e4e\u7684\u5168\u90e8\u8bdd\u9898\u901a\u8fc7\u7236\u5b50\u5173\u7cfb\u6784\u6210\u4e00\u4e2a<b>\u6709\u6839\u65e0\u5faa\u73af\u7684\u6709\u5411\u56fe<\/b>\u3002<br>\u6240\u6709\u6ca1\u6709\u76f4\u63a5\u6dfb\u52a0\u7236\u8bdd\u9898\u7684\u8bdd\u9898\u4f1a\u81ea\u52a8\u6210\u4e3a<b>\u300c\u672a\u5f52\u7c7b\u300d\u8bdd\u9898<\/b>\u7684\u5b50\u8bdd\u9898\uff0c\u4ece\u800c\u4e0e\u6574\u4e2a\u8bdd\u9898\u6811\u8fde\u63a5\u8d77\u6765\u3002", "avatar_url": "https://pic3.zhimg.com/50/v2-a6302d07f9514c05c19885e76be4dd42_qhd.jpg", "type": "topic", "id": "19776751"}, {"is_black": false, "name": "\u5b66\u79d1", "url": "http://www.zhihu.com/api/v3/topics/19618774", "excerpt": "\u5b66\u79d1\u8be5\u8bcd\u6709\u4ee5\u4e0b\u4e24\u79cd\u542b\u4e49\uff1a\u2460\u76f8\u5bf9\u72ec\u7acb\u7684\u77e5\u8bc6\u4f53\u7cfb\u3002\u4eba\u7c7b\u6240\u6709\u7684\u77e5\u8bc6\u5212\u5206\u4e3a\u4e94\u5927\u95e8\u7c7b\uff1a\u81ea\u7136\u79d1\u5b66\uff0c\u519c\u4e1a\u79d1\u5b66\uff0c\u533b\u836f\u79d1\u5b66\uff0c\u5de5\u7a0b\u4e0e\u6280\u672f\u79d1\u5b66\uff0c\u4eba\u6587\u4e0e\u793e\u4f1a\u79d1\u5b66\u3002\u2461\u6211\u56fd\u9ad8\u7b49\u5b66\u6821\u672c\u79d1\u6559\u80b2\u4e13\u4e1a\u8bbe\u7f6e\u7684\u5b66\u79d1\u5206\u7c7b\uff0c\u6211\u56fd\u9ad8\u7b49\u6559\u80b2\u5212\u5206\u4e3a13\u4e2a\u5b66\u79d1\u95e8\u7c7b\uff1a\u54f2\u5b66\u3001\u7ecf\u6d4e\u5b66\u3001\u6cd5\u5b66\u3001\u6559\u80b2\u5b66\u3001\u6587\u5b66\u3001\u5386\u53f2\u5b66\u3001\u7406\u5b66\u3001\u5de5\u5b66\u3001\u519c\u5b66\u3001\u533b\u5b66\u3001\u519b\u4e8b\u5b66\u3001\u7ba1\u7406\u5b66\u3001\u827a\u672f\u5b66\u3002", "is_vote": false, "introduction": "\u5b66\u79d1\u8be5\u8bcd\u6709\u4ee5\u4e0b\u4e24\u79cd\u542b\u4e49\uff1a\u2460\u76f8\u5bf9\u72ec\u7acb\u7684\u77e5\u8bc6\u4f53\u7cfb\u3002\u4eba\u7c7b\u6240\u6709\u7684\u77e5\u8bc6\u5212\u5206\u4e3a\u4e94\u5927\u95e8\u7c7b\uff1a\u81ea\u7136\u79d1\u5b66\uff0c\u519c\u4e1a\u79d1\u5b66\uff0c\u533b\u836f\u79d1\u5b66\uff0c\u5de5\u7a0b\u4e0e\u6280\u672f\u79d1\u5b66\uff0c\u4eba\u6587\u4e0e\u793e\u4f1a\u79d1\u5b66\u3002\u2461\u6211\u56fd\u9ad8\u7b49\u5b66\u6821\u672c\u79d1\u6559\u80b2\u4e13\u4e1a\u8bbe\u7f6e\u7684\u5b66\u79d1\u5206\u7c7b\uff0c\u6211\u56fd\u9ad8\u7b49\u6559\u80b2\u5212\u5206\u4e3a13\u4e2a\u5b66\u79d1\u95e8\u7c7b\uff1a\u54f2\u5b66\u3001\u7ecf\u6d4e\u5b66\u3001\u6cd5\u5b66\u3001\u6559\u80b2\u5b66\u3001\u6587\u5b66\u3001\u5386\u53f2\u5b66\u3001\u7406\u5b66\u3001\u5de5\u5b66\u3001\u519c\u5b66\u3001\u533b\u5b66\u3001\u519b\u4e8b\u5b66\u3001\u7ba1\u7406\u5b66\u3001\u827a\u672f\u5b66\u3002", "avatar_url": "https://pic2.zhimg.com/50/8658418bc_qhd.jpg", "type": "topic", "id": "19618774"}, {"is_black": false, "name": "\u5b9e\u4f53", "url": "http://www.zhihu.com/api/v3/topics/19778287", "excerpt": "\u5b9e\u4f53\uff08entity\uff09\u662f\u6709\u53ef\u533a\u522b\u6027\u4e14\u72ec\u7acb\u5b58\u5728\u7684\u67d0\u79cd\u4e8b\u7269\uff0c\u4f46\u5b83\u4e0d\u9700\u8981\u662f\u7269\u8d28\u4e0a\u7684\u5b58\u5728\u3002\u5c24\u5176\u662f\u62bd\u8c61\u548c\u6cd5\u5f8b\u62df\u5236\u4e5f\u901a\u5e38\u88ab\u89c6\u4e3a\u5b9e\u4f53\u3002\u5b9e\u4f53\u53ef\u88ab\u770b\u6210\u662f\u4e00\u5305\u542b\u6709\u5b50\u96c6\u7684\u96c6\u5408\u3002\u5728\u54f2\u5b66\u91cc\uff0c\u8fd9\u79cd\u96c6\u5408\u88ab\u79f0\u4e3a\u5ba2\u4f53\u3002\u5b9e\u4f53\u53ef\u88ab\u4f7f\u7528\u6765\u6307\u6d89\u67d0\u4e2a\u53ef\u80fd\u662f\u4eba\u3001\u52a8\u7269\u3001\u690d\u7269\u6216\u771f\u83cc\u7b49\u4e0d\u4f1a\u601d\u8003\u7684\u751f\u547d\u3001\u65e0\u751f\u547d\u7269\u4f53\u6216\u4fe1\u5ff5\u7b49\u7684\u4e8b\u7269\u3002\u5728\u8fd9\u4e00\u65b9\u9762\uff0c\u5b9e\u4f53\u53ef\u4ee5\u88ab\u89c6\u4e3a\u4e00\u5168\u5305\u7684\u8bcd\u8bed\u3002\u6709\u65f6\uff0c\u5b9e\u4f53\u88ab\u5f53\u505a\u672c\u8d28\u7684\u5e7f\u4e49\uff0c\u4e0d\u8bba\u5373\u6307\u7684\u662f\u5426\u4e3a\u7269\u8d28\u4e0a\u7684\u5b58\u5728\uff0c\u5982\u65f6\u5e38\u4f1a\u6307\u6d89\u5230\u7684\u65e0\u7269\u8d28\u5f62\u5f0f\u7684\u5b9e\u4f53\u2026", "is_vote": false, "introduction": "\u5b9e\u4f53\uff08entity\uff09\u662f\u6709\u53ef\u533a\u522b\u6027\u4e14\u72ec\u7acb\u5b58\u5728\u7684\u67d0\u79cd\u4e8b\u7269\uff0c\u4f46\u5b83\u4e0d\u9700\u8981\u662f\u7269\u8d28\u4e0a\u7684\u5b58\u5728\u3002\u5c24\u5176\u662f\u62bd\u8c61\u548c\u6cd5\u5f8b\u62df\u5236\u4e5f\u901a\u5e38\u88ab\u89c6\u4e3a\u5b9e\u4f53\u3002\u5b9e\u4f53\u53ef\u88ab\u770b\u6210\u662f\u4e00\u5305\u542b\u6709\u5b50\u96c6\u7684\u96c6\u5408\u3002\u5728\u54f2\u5b66\u91cc\uff0c\u8fd9\u79cd\u96c6\u5408\u88ab\u79f0\u4e3a\u5ba2\u4f53\u3002\u5b9e\u4f53\u53ef\u88ab\u4f7f\u7528\u6765\u6307\u6d89\u67d0\u4e2a\u53ef\u80fd\u662f\u4eba\u3001\u52a8\u7269\u3001\u690d\u7269\u6216\u771f\u83cc\u7b49\u4e0d\u4f1a\u601d\u8003\u7684\u751f\u547d\u3001\u65e0\u751f\u547d\u7269\u4f53\u6216\u4fe1\u5ff5\u7b49\u7684\u4e8b\u7269\u3002\u5728\u8fd9\u4e00\u65b9\u9762\uff0c\u5b9e\u4f53\u53ef\u4ee5\u88ab\u89c6\u4e3a\u4e00\u5168\u5305\u7684\u8bcd\u8bed\u3002\u6709\u65f6\uff0c\u5b9e\u4f53\u88ab\u5f53\u505a\u672c\u8d28\u7684\u5e7f\u4e49\uff0c\u4e0d\u8bba\u5373\u6307\u7684\u662f\u5426\u4e3a\u7269\u8d28\u4e0a\u7684\u5b58\u5728\uff0c\u5982\u65f6\u5e38\u4f1a\u6307\u6d89\u5230\u7684\u65e0\u7269\u8d28\u5f62\u5f0f\u7684\u5b9e\u4f53\uff0d\u8bed\u8a00\u3002\u66f4\u6709\u751a\u8005\uff0c\u5b9e\u4f53\u6709\u65f6\u4ea6\u6307\u5b58\u5728\u6216\u672c\u8d28\u672c\u8eab\u3002\u5728\u6cd5\u5f8b\u4e0a\uff0c\u5b9e\u4f53\u662f\u6307\u80fd\u5177\u6709\u6743\u5229\u548c\u4e49\u52a1\u7684\u4e8b\u7269\u3002\u8fd9\u901a\u5e38\u662f\u6307\u6cd5\u4eba\uff0c\u4f46\u4e5f\u5305\u62ec\u81ea\u7136\u4eba\u3002\u00a0<br><b>\u8b66\u544a\uff1a\u8bf7\u4e0d\u8981\u5c06\u672c\u8bdd\u9898\u5f53\u505a\u5e7f\u544a\u5ba3\u4f20\u7684\u573a\u6240\uff0c\u4e5f\u4e0d\u8981\u5c06\u5173\u4e8e\u5b9e\u4f53\u7ecf\u6d4e\u3001\u5b9e\u4f53\u5e97\u7b49\u7684\u63d0\u95ee\u76f4\u63a5\u7ed1\u5b9a\u5230\u8be5\u8bdd\u9898\u4e0a\u3002<\/b>", "avatar_url": "https://pic4.zhimg.com/50/37c2b84c88dea33eb4e36a1bdb558386_qhd.jpg", "type": "topic", "id": "19778287"}, {"is_black": false, "name": "\u300c\u5f62\u800c\u4e0a\u300d\u8bdd\u9898", "url": "http://www.zhihu.com/api/v3/topics/19778298", "excerpt": "\u300c\u5f62\u800c\u4e0a\u300d\u8bdd\u9898\u4e0b\u6536\u5f55\u4e86\u4e00\u4e9b\u8ba8\u8bba\u6982\u5ff5\u3001\u903b\u8f91\u3001\u542b\u4e49\u548c\u539f\u56e0\u7b49\u62bd\u8c61\u5185\u5bb9\u7684\u5b50\u8bdd\u9898\u3002 \u300c\u5f62\u800c\u4e0a\u300d\u662f\u65e5\u672c\u4eba\u4e95\u4e0a\u54f2\u6b21\u90ce\u5bf9 metaphysic \u4e00\u8bcd\u7684\u6c49\u5b57\u7ffb\u8bd1\uff0c\u8bed\u51fa\u300a\u6613\u7ecf\u300b\u4e2d\u300c\u5f62\u800c\u4e0a\u8005\u8c13\u4e4b\u9053\uff0c\u5f62\u800c\u4e0b\u8005\u8c13\u4e4b\u5668\u300d \u8bf7\u4e0d\u8981\u5728\u95ee\u9898\u4e0a\u76f4\u63a5\u7ed1\u5b9a\u300c\u5f62\u800c\u4e0a\u300d\u8bdd\u9898", "is_vote": false, "introduction": "<b>\u300c\u5f62\u800c\u4e0a\u300d\u8bdd\u9898<\/b>\u4e0b\u6536\u5f55\u4e86\u4e00\u4e9b\u8ba8\u8bba\u6982\u5ff5\u3001\u903b\u8f91\u3001\u542b\u4e49\u548c\u539f\u56e0\u7b49\u62bd\u8c61\u5185\u5bb9\u7684\u5b50\u8bdd\u9898\u3002<br><br>\u300c\u5f62\u800c\u4e0a\u300d\u662f\u65e5\u672c\u4eba\u4e95\u4e0a\u54f2\u6b21\u90ce\u5bf9 metaphysic \u4e00\u8bcd\u7684\u6c49\u5b57\u7ffb\u8bd1\uff0c\u8bed\u51fa\u300a\u6613\u7ecf\u300b\u4e2d\u300c\u5f62\u800c\u4e0a\u8005\u8c13\u4e4b\u9053\uff0c\u5f62\u800c\u4e0b\u8005\u8c13\u4e4b\u5668\u300d<br><br><i>\u8bf7\u4e0d\u8981\u5728\u95ee\u9898\u4e0a\u76f4\u63a5\u7ed1\u5b9a\u300c\u5f62\u800c\u4e0a\u300d\u8bdd\u9898<\/i>", "avatar_url": "https://pic1.zhimg.com/50/e4875269379c310d8cffa414b7905caa_qhd.jpg", "type": "topic", "id": "19778298"}, {"category": "n_wiki", "is_black": false, "name": "\u4ea7\u4e1a", "url": "http://www.zhihu.com/api/v3/topics/19560891", "excerpt": "\u4ea7\u4e1a\u662f\u793e\u4f1a\u5206\u5de5\u548c\u751f\u4ea7\u529b\u4e0d\u65ad\u53d1\u5c55\u7684\u4ea7\u7269\u3002\u4ea7\u4e1a\u662f\u793e\u4f1a\u5206\u5de5\u7684\u4ea7\u7269\uff0c\u5b83\u968f\u7740\u793e\u4f1a\u5206\u5de5\u7684\u4ea7\u751f\u800c\u4ea7\u751f\uff0c\u5e76\u968f\u7740\u793e\u4f1a\u5206\u5de5\u7684\u53d1\u5c55\u800c\u53d1\u5c55\u3002\u5728\u8fdc\u53e4\u65f6\u4ee3\uff0c\u4eba\u7c7b\u5171\u540c\u52b3\u52a8\uff0c\u5171\u540c\u751f\u6d3b\u3002", "is_vote": false, "introduction": "\u4ea7\u4e1a\u662f\u793e\u4f1a\u5206\u5de5\u548c\u751f\u4ea7\u529b\u4e0d\u65ad\u53d1\u5c55\u7684\u4ea7\u7269\u3002\u4ea7\u4e1a\u662f\u793e\u4f1a\u5206\u5de5\u7684\u4ea7\u7269\uff0c\u5b83\u968f\u7740\u793e\u4f1a\u5206\u5de5\u7684\u4ea7\u751f\u800c\u4ea7\u751f\uff0c\u5e76\u968f\u7740\u793e\u4f1a\u5206\u5de5\u7684\u53d1\u5c55\u800c\u53d1\u5c55\u3002\u5728\u8fdc\u53e4\u65f6\u4ee3\uff0c\u4eba\u7c7b\u5171\u540c\u52b3\u52a8\uff0c\u5171\u540c\u751f\u6d3b\u3002", "avatar_url": "https://pic2.zhimg.com/50/358d89bd7de70e5891d5fd47138ff242_qhd.jpg", "type": "topic", "id": "19560891"}, {"is_black": false, "name": "\u751f\u6d3b\u3001\u827a\u672f\u3001\u6587\u5316\u4e0e\u6d3b\u52a8", "url": "http://www.zhihu.com/api/v3/topics/19778317", "excerpt": "\u4ee5\u4eba\u7c7b\u96c6\u4f53\u884c\u4e3a\u548c\u4eba\u7c7b\u793e\u4f1a\u6587\u660e\u4e3a\u4e3b\u4f53\u7684\u8bdd\u9898\uff0c\u5176\u5185\u5bb9\u4e3b\u8981\u5305\u542b\u751f\u6d3b\u3001\u827a\u672f\u3001\u6587\u5316\u3001\u6d3b\u52a8\u56db\u4e2a\u65b9\u9762\u3002", "is_vote": false, "introduction": "\u4ee5\u4eba\u7c7b\u96c6\u4f53\u884c\u4e3a\u548c\u4eba\u7c7b\u793e\u4f1a\u6587\u660e\u4e3a\u4e3b\u4f53\u7684\u8bdd\u9898\uff0c\u5176\u5185\u5bb9\u4e3b\u8981\u5305\u542b\u751f\u6d3b\u3001\u827a\u672f\u3001\u6587\u5316\u3001\u6d3b\u52a8\u56db\u4e2a\u65b9\u9762\u3002", "avatar_url": "https://pic3.zhimg.com/50/6df49c633_qhd.jpg", "type": "topic", "id": "19778317"}]}`

	pt := &PagingTopic{}
	if err := json.Unmarshal([]byte(str), pt); err != nil {
		t.Fatalf("%s", err)
	}
	fmt.Printf("paging: %+v\n", *pt.Paging)
	for _, topic := range pt.Data {
		fmt.Printf("topic: %+v\n", *topic)
	}
}
