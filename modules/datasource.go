package modules

import (
	"database/sql"
	"fmt"

	"github.com/astaxie/beego/logs"

	_ "github.com/go-sql-driver/mysql"
)

const (
	urlTokenTable = "urlToken"
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

	querySelect := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE urlToken=?`, urlTokenTable)
	stmtSelect, err := db.Prepare(querySelect)
	if err != nil {
		return err
	}
	defer stmtSelect.Close()

	queryInsert := fmt.Sprintf(`INSERT INTO %s (urlToken) VALUES (?)`, urlTokenTable)
	stmtInsert, err := db.Prepare(queryInsert)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	for _, urlToken := range urlTokens {
		row := stmtSelect.QueryRow(urlToken)
		var count int
		if err := row.Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			logs.Warn("duplicate urlToken: %s", urlToken)
			continue
		}

		if _, err := stmtInsert.Exec(urlToken); err != nil {
			return err
		}
	}

	return nil
}

func (ds *DataSource) GetURLToken(offset int) (string, error) {
	query := fmt.Sprintf(`SELECT urlToken FROM %s LIMIT ?,1`, urlTokenTable)
	row := ds.db.QueryRow(query, offset)

	var urlToken string
	if err := row.Scan(&urlToken); err != nil {
		return "", err
	}
	return urlToken, nil
}

func (ds *DataSource) Truncate(tableName string) error {
	query := fmt.Sprintf(`TRUNCATE TABLE %s`, tableName)
	if _, err := ds.db.Exec(query); err != nil {
		return err
	}

	return nil
}
