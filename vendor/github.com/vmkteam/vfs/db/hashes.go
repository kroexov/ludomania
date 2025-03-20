package db

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-pg/pg/v10"
)

const tempTableName = `tmp_vfsHashes`

// CreateTempHashesTable creates a temporary table for hashes.
func (db DB) CreateTempHashesTable(ctx context.Context, tx *pg.Tx) error {
	query := fmt.Sprintf(
		`CREATE TEMP TABLE "%s"
(
	hash varchar(40) default ''::character varying not null,
	namespace varchar(32) default 'default'::character varying,
	"fileSize" integer default 0 not null,
	extension varchar(4) default 'jpg'::character varying
) ON COMMIT DROP`, tempTableName,
	)
	_, err := tx.ExecContext(ctx, query)
	return err
}

// CopyHashesFromSTDIN fills temporary hashes table with CSV data
func (db DB) CopyHashesFromSTDIN(tx *pg.Tx, r io.Reader) (int, error) {
	sql := fmt.Sprintf(`COPY "%s" FROM STDIN DELIMITER ';' CSV`, tempTableName)
	res, err := tx.CopyFrom(r, sql)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// UpsertHashesTable upserts data from tempTable into vfsHashes.
func (db DB) UpsertHashesTable(ctx context.Context, tx *pg.Tx) (int, time.Duration, error) {
	t0 := time.Now()
	query := fmt.Sprintf(`INSERT INTO "vfsHashes" ("hash", "namespace", "fileSize", "extension")
    (SELECT "hash", COALESCE("namespace", 'default'), "fileSize", COALESCE("extension", 'jpg') FROM "%s")
	ON CONFLICT ("hash", "namespace") DO NOTHING`, tempTableName)
	res, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, time.Duration(0), err
	}
	return res.RowsAffected(), time.Since(t0), nil
}
