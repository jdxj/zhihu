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
