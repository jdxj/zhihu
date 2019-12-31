package modules

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	urlTokenTable         = "urlToken"
	urlTokenProgressTable = "urlTokenProgress"
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
			continue
		}

		if _, err := stmtInsert.Exec(urlToken); err != nil {
			return err
		}
	}

	return nil
}

func (ds *DataSource) GetURLToken(offset uint64) (*URLToken, error) {
	ut := &URLToken{}

	query := fmt.Sprintf(`SELECT id,urlToken FROM %s ORDER BY id LIMIT ?,1`, urlTokenTable)
	row := ds.db.QueryRow(query, offset)

	if err := row.Scan(ut.ToScan()...); err != nil {
		return nil, err
	}
	return ut, nil
}

func (ds *DataSource) GetURLTokenOffset(urlTokenID uint64) (uint64, error) {
	var offset uint64

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id<? ORDER BY id", urlTokenTable)
	row := ds.db.QueryRow(query, urlTokenID)
	if err := row.Scan(&offset); err != nil {
		return 0, err
	}
	return offset, nil
}

func (ds *DataSource) GetURLTokenProgress() (*URLTokenProgress, error) {
	utp := &URLTokenProgress{}

	query := fmt.Sprintf(`SELECT id,urlTokenID,nextFolloweeURL,nextFollowerURL
FROM %s ORDER BY id DESC LIMIT 1`, urlTokenProgressTable)

	row := ds.db.QueryRow(query)
	if err := row.Scan(utp.ToScan()...); err != nil {
		return nil, err
	}
	return utp, nil
}

func (ds *DataSource) InsertURLTokenProgress(utp *URLTokenProgress) error {
	query := fmt.Sprintf(`INSERT INTO %s (urlTokenID,nextFolloweeURL,nextFollowerURL) VALUES (?,?,?)`,
		urlTokenProgressTable)
	if _, err := ds.db.Exec(query, utp.ToInsert()...); err != nil {
		return err
	}
	return nil
}

func (ds *DataSource) Truncate(tableName string) error {
	query := fmt.Sprintf(`TRUNCATE TABLE %s`, tableName)
	if _, err := ds.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (ds *DataSource) CountURLToken() (count uint64, err error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, urlTokenTable)
	row := ds.db.QueryRow(query)

	return count, row.Scan(&count)
}
