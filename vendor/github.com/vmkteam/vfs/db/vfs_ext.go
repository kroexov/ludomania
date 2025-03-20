package db

import (
	"context"
	"strings"

	"github.com/go-pg/pg/v10"
)

func (vr VfsRepo) FolderBranch(ctx context.Context, folderId int) (list []VfsFolder, err error) {
	_, err = vr.db.QueryContext(ctx, &list, `
WITH RECURSIVE r AS (
   SELECT *, 1 AS level from "vfsFolders"
   WHERE "folderId" = ?
   UNION SELECT ff.*, r.level + 1 AS level
   FROM "vfsFolders" ff
 		JOIN r ON ff."folderId" = r."parentFolderId"
)
SELECT * FROM r ORDER by level DESC`, folderId)

	return
}

func (vfs *VfsFileSearch) WithQuery(query *string) *VfsFileSearch {
	if query != nil && *query != "" {
		vfs.TitleILike = query
	}

	return vfs
}

func (vr VfsRepo) NextFileID() (int, error) {
	var max int
	_, err := vr.db.Query(pg.Scan(&max), `select nextval('"vfsFiles_fileId_seq"')`)

	return max, err
}

func (vr VfsRepo) UpdateFilesFolder(ctx context.Context, fileIDs []int64, newFolderId int) (bool, error) {
	q := vr.db.ModelContext(ctx, &VfsFile{FolderID: newFolderId}).
		Column(Columns.VfsFile.FolderID).
		Where("? IN (?)", pg.Ident(Columns.VfsFile.ID), pg.Ints(fileIDs))

	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

func (vr VfsRepo) DeleteVfsFiles(ctx context.Context, fileIDs []int64) (bool, error) {
	q := vr.db.ModelContext(ctx, &VfsFile{StatusID: StatusDeleted}).
		Column(Columns.VfsFile.StatusID).
		Where("? IN (?)", pg.Ident(Columns.VfsFile.ID), pg.Ints(fileIDs))

	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

func (vr VfsRepo) HashesForUpdate(ctx context.Context, limit uint64) (list []VfsHash, err error) {
	_, err = vr.db.QueryContext(
		ctx,
		&list,
		`SELECT "`+
			strings.Join([]string{
				Columns.VfsHash.Hash,
				Columns.VfsHash.Namespace,
				Columns.VfsHash.Extension,
			}, `", "`)+`"`+
			` FROM "`+Tables.VfsHash.Name+`"`+
			` WHERE "`+Columns.VfsHash.IndexedAt+`" IS NULL`+
			` LIMIT ?`+
			` FOR NO KEY UPDATE SKIP LOCKED`,
		limit,
	)
	return
}

// SaveVfsHash checks hash in DB and adds it if hash was not found.
func (vr VfsRepo) SaveVfsHash(ctx context.Context, hash *VfsHash) (err error) {
	h, err := vr.OneVfsHash(ctx, &VfsHashSearch{
		Hash:      &hash.Hash,
		Namespace: &hash.Namespace,
	})

	if err != nil {
		return err
	}

	// check hash for existence in db
	if h != nil {
		return nil
	}

	_, err = vr.AddVfsHash(ctx, hash)
	return err
}
