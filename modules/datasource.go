package modules

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	urlTokenTable         = "urlToken"
	urlTokenProgressTable = "urlTokenProgress"
	topicIDTable          = "topicID"
	topicIDProgressTable  = "topicIDProgress"
	topicTable            = "topic"
	topicProgressTable    = "topicProgress"
	industryTable         = "industry"
	peopleTable           = "people"
	peopleProgressTable   = "peopleProgress"
)

func NewDataSource(config *MySQLConfig) (*DataSource, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid mysql config")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?loc=Local&parseTime=true",
		config.User, config.Password, config.Host, config.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ds := &DataSource{
		db: db,
	}
	return ds, nil
}

type DataSource struct {
	db *sql.DB
}

func (ds *DataSource) Close() error {
	return ds.db.Close()
}

func (ds *DataSource) InsertURLTokens(urlTokens []string) error {
	db := ds.db

	queryInsert := fmt.Sprintf(`INSERT INTO %s (urlToken) VALUES (?)`, urlTokenTable)
	stmtInsert, err := db.Prepare(queryInsert)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	for _, urlToken := range urlTokens {
		if _, err := stmtInsert.Exec(urlToken); err != nil {
			if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry") {
				continue
			}
			return err
		}
	}

	return nil
}

func (ds *DataSource) GetURLToken(offset uint64) (*URLToken, error) {
	ut := &URLToken{}
	//query := fmt.Sprintf(`SELECT id,urlToken FROM %s ORDER BY id LIMIT ?,1`, urlTokenTable)
	query := fmt.Sprintf(`SELECT id,urlToken FROM %s WHERE id>? ORDER BY id LIMIT 0,1`, urlTokenTable)
	row := ds.db.QueryRow(query, offset)
	return ut, row.Scan(ut.ToScan()...)
}

func (ds *DataSource) GetURLTokenOffset(urlTokenID uint64) (uint64, error) {
	var offset uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id<? ORDER BY id", urlTokenTable)
	row := ds.db.QueryRow(query, urlTokenID)
	return offset, row.Scan(&offset)
}

func (ds *DataSource) GetURLTokenProgress() (*URLTokenProgress, error) {
	utp := &URLTokenProgress{}
	query := fmt.Sprintf(`SELECT id,urlTokenID,nextFolloweeURL,nextFollowerURL
FROM %s ORDER BY id DESC LIMIT 1`, urlTokenProgressTable)
	row := ds.db.QueryRow(query)
	return utp, row.Scan(utp.ToScan()...)
}

func (ds *DataSource) InsertURLTokenProgress(utp *URLTokenProgress) error {
	query := fmt.Sprintf(`INSERT INTO %s (urlTokenID,nextFolloweeURL,nextFollowerURL) VALUES (?,?,?)`,
		urlTokenProgressTable)
	_, err := ds.db.Exec(query, utp.ToInsert()...)
	return err
}

func (ds *DataSource) Truncate(tableName string) error {
	query := fmt.Sprintf(`TRUNCATE TABLE %s`, tableName)
	_, err := ds.db.Exec(query)
	return err
}

func (ds *DataSource) CountURLToken() (count uint64, err error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, urlTokenTable)
	row := ds.db.QueryRow(query)

	return count, row.Scan(&count)
}

func (ds *DataSource) InsertTopicsID(topicsID []*TopicID) error {
	db := ds.db

	queryInsert := fmt.Sprintf(`INSERT INTO %s (topicID,name) VALUES (?,?)`, topicIDTable)
	stmtInsert, err := db.Prepare(queryInsert)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	for _, topicID := range topicsID {
		if _, err := stmtInsert.Exec(topicID.ToInsert()...); err != nil {
			if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry") {
				continue
			}
			return err
		}
	}

	return nil
}

func (ds *DataSource) GetTopicID(offset uint64) (*TopicID, error) {
	ti := &TopicID{}
	query := fmt.Sprintf(`SELECT id,topicID,name FROM %s ORDER BY id LIMIT ?,1`, topicIDTable)
	row := ds.db.QueryRow(query, offset)
	return ti, row.Scan(ti.ToScan()...)
}

func (ds *DataSource) GetTopicIDOffset(topicID uint64) (uint64, error) {
	var offset uint64
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE id<? ORDER BY id`, topicIDTable)
	row := ds.db.QueryRow(query, topicID)
	return offset, row.Scan(&offset)
}

func (ds *DataSource) GetTopicIDProgress() (*TopicIDProgress, error) {
	tip := &TopicIDProgress{}
	query := fmt.Sprintf(`SELECT id,topicID,nextTopicIDURL
FROM %s ORDER BY id DESC LIMIT 1`, topicIDProgressTable)
	row := ds.db.QueryRow(query)
	return tip, row.Scan(tip.ToScan()...)
}

func (ds *DataSource) InsertTopicIDProgress(tip *TopicIDProgress) error {
	query := fmt.Sprintf(`INSERT INTO %s (topicID,nextTopicIDURL) VALUES (?,?)`, topicIDProgressTable)
	_, err := ds.db.Exec(query, tip.ToInsert()...)
	return err
}

func (ds *DataSource) InsertTopic(tt *TopicTable) error {
	db := ds.db

	querySelect := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE topicID=?`, topicTable)
	row := db.QueryRow(querySelect, tt.TopicID)

	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return nil
	}

	queryInsert := fmt.Sprintf(`INSERT INTO %s (topicID,followerCount,questionCount) VALUES (?,?,?)`, topicTable)
	_, err := db.Exec(queryInsert, tt.TopicID, tt.FollowerCount, tt.QuestionCount)
	return err
}

func (ds *DataSource) GetTopicProgress() (*TopicProgress, error) {
	query := fmt.Sprintf(`SELECT id,topicID FROM %s ORDER BY id DESC LIMIT 1`, topicProgressTable)
	row := ds.db.QueryRow(query)
	tp := &TopicProgress{}
	return tp, row.Scan(tp.ToScan()...)
}

func (ds *DataSource) InsertTopicProgress(tp *TopicProgress) error {
	query := fmt.Sprintf(`INSERT INTO %s (topicID) VALUES (?)`, topicProgressTable)
	_, err := ds.db.Exec(query, tp.ToInsert()...)
	return err
}

func (ds *DataSource) InsertIndustry(industry string) (uint64, error) {
	db := ds.db

	querySelect := fmt.Sprintf(`SELECT id,name FROM %s WHERE name=?`, industryTable)
	queryInsert := fmt.Sprintf(`INSERT INTO %s (name) VALUES (?)`, industryTable)

	row := db.QueryRow(querySelect, industry)

	ind := &Industry{}
	err := row.Scan(ind.ToScan()...)
	if err == sql.ErrNoRows {
		if res, err := db.Exec(queryInsert, industry); err != nil {
			return 0, err
		} else {
			id, err := res.LastInsertId()
			return uint64(id), err
		}
	} else if err != nil {
		return 0, err
	} else {
		return ind.ID, nil
	}
}

func (ds *DataSource) InsertPeople(people *People) error {
	if people == nil {
		return fmt.Errorf("invalid people data")
	}

	query := fmt.Sprintf(`INSERT INTO %s (urlTokenID,name,headline,description,gender,followeeCount,followerCount,answerCount,questionCount,articlesCount,columnsCount,industry,address,school,major,entranceYear,graduationYear,company,job) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, peopleTable)
	_, err := ds.db.Exec(query, people.ToInsert()...)
	return err
}
