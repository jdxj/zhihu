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
