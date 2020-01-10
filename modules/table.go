package modules

type URLToken struct {
	ID       uint64
	URLToken string
}

func (ut *URLToken) ToScan() []interface{} {
	var fields []interface{}
	fields = append(fields, &ut.ID)
	fields = append(fields, &ut.URLToken)
	return fields
}

type URLTokenProgress struct {
	ID              uint64
	URLTokenID      uint64
	NextFolloweeURL string
	NextFollowerURL string
}

func (utp *URLTokenProgress) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &utp.ID)
	fields = append(fields, &utp.URLTokenID)
	fields = append(fields, &utp.NextFolloweeURL)
	fields = append(fields, &utp.NextFollowerURL)

	return fields
}

func (utp *URLTokenProgress) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, utp.URLTokenID)
	fields = append(fields, utp.NextFolloweeURL)
	fields = append(fields, utp.NextFollowerURL)

	return fields
}

type TopicID struct {
	ID      uint64
	TopicID string
	Name    string
}

func (ti *TopicID) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &ti.ID)
	fields = append(fields, &ti.TopicID)
	fields = append(fields, &ti.Name)
	return fields
}

func (ti *TopicID) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, ti.TopicID)
	fields = append(fields, ti.Name)
	return fields
}

type TopicIDProgress struct {
	ID             uint64
	TopicID        uint64
	NextTopicIDURL string
}

func (tip *TopicIDProgress) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &tip.ID)
	fields = append(fields, &tip.TopicID)
	fields = append(fields, &tip.NextTopicIDURL)
	return fields
}

func (tip *TopicIDProgress) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, tip.TopicID)
	fields = append(fields, tip.NextTopicIDURL)
	return fields
}

// TopicTable 有重名的结构, 所以使用了 Table 后缀
type TopicTable struct {
	ID            uint64
	TopicID       uint64
	FollowerCount uint64
	QuestionCount uint64
}

func (tt *TopicTable) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, tt.TopicID)
	fields = append(fields, tt.FollowerCount)
	fields = append(fields, tt.QuestionCount)
	return fields
}

type TopicProgress struct {
	ID      uint64
	TopicID uint64
}

func (tp *TopicProgress) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &tp.ID)
	fields = append(fields, &tp.TopicID)
	return fields
}

func (tp *TopicProgress) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, tp.TopicID)
	return fields
}

type Industry struct {
	ID   uint64
	Name string
}

func (ind *Industry) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &ind.ID)
	fields = append(fields, &ind.Name)
	return fields
}

// todo: 实现
type People struct {
	URLTokenID    uint64
	Name          string        `json:"name"`
	Headline      string        `json:"headline"`
	Description   string        `json:"description"`
	Gender        int           `json:"gender"`
	FolloweeCount uint64        `json:"followingCount"`
	FollowerCount uint64        `json:"followerCount"`
	AnswerCount   uint64        `json:"answerCount"`
	QuestionCount uint64        `json:"questionCount"`
	ArticlesCount uint64        `json:"articlesCount"`
	ColumnsCount  uint64        `json:"columnsCount"`
	Business      *Business     `json:"business"`
	Locations     []*Location   `json:"locations"`
	Educations    []*Education  `json:"educations"`
	Employments   []*Employment `json:"employments"`
}

func (p *People) ToInsert() []interface{} {
	var fields []interface{}

	fields = append(fields, p.URLTokenID)
	fields = append(fields, p.Name)
	fields = append(fields, p.Headline)
	fields = append(fields, p.Description)
	fields = append(fields, p.Gender)
	fields = append(fields, p.FolloweeCount)
	fields = append(fields, p.FollowerCount)
	fields = append(fields, p.AnswerCount)
	fields = append(fields, p.QuestionCount)
	fields = append(fields, p.ArticlesCount)
	fields = append(fields, p.ColumnsCount)

	var industry, address, school, major, company, job string
	var entrance, graduation int

	// 行业
	if p.Business != nil {
		industry = p.Business.Name
	}
	fields = append(fields, industry)

	// 住址
	if len(p.Locations) != 0 {
		address = p.Locations[0].Name
	}
	fields = append(fields, address)

	// 教育
	if len(p.Educations) != 0 {
		school = p.Educations[0].School.Name
		major = p.Educations[0].Major.Name
		entrance = p.Educations[0].EntranceYear
		graduation = p.Educations[0].GraduationYear
	}
	fields = append(fields, school, major, entrance, graduation)

	// 职业
	if len(p.Employments) != 0 {
		company = p.Employments[0].Company.Name
		job = p.Employments[0].Job.Name
	}
	fields = append(fields, company, job)
	return fields
}

type Business struct {
	Name string `json:"name"`
}

type Location struct {
	Name string `json:"name"`
}

type Education struct {
	School         *School `json:"school"`
	Major          *Major  `json:"major"`
	EntranceYear   int     `json:"entranceYear"`
	GraduationYear int     `json:"graduationYear"`
}

type School struct {
	Name string `json:"name"`
}

type Major struct {
	Name string `json:"name"`
}

type PeopleProgress struct {
	ID         uint64
	URLTokenID uint64
}

func (pp *PeopleProgress) ToScan() []interface{} {
	var fields []interface{}

	fields = append(fields, &pp.ID)
	fields = append(fields, &pp.ID)
	return fields
}
