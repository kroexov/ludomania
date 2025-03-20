package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type VfsRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewVfsRepo returns new repository
func NewVfsRepo(db orm.DB) VfsRepo {
	return VfsRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.VfsFile.Name:   {StatusFilter},
			Tables.VfsFolder.Name: {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.VfsFile.Name:   {{Column: Columns.VfsFile.CreatedAt, Direction: SortDesc}},
			Tables.VfsFolder.Name: {{Column: Columns.VfsFolder.CreatedAt, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.VfsFile.Name:   {TableColumns, Columns.VfsFile.Folder},
			Tables.VfsFolder.Name: {TableColumns, Columns.VfsFolder.ParentFolder},
		},
	}
}

// WithTransaction is a function that wraps VfsRepo with pg.Tx transaction.
func (vr VfsRepo) WithTransaction(tx *pg.Tx) VfsRepo {
	vr.db = tx
	return vr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (vr VfsRepo) WithEnabledOnly() VfsRepo {
	f := make(map[string][]Filter, len(vr.filters))
	for i := range vr.filters {
		f[i] = make([]Filter, len(vr.filters[i]))
		copy(f[i], vr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	vr.filters = f

	return vr
}

/*** VfsFile ***/

// FullVfsFile returns full joins with all columns
func (vr VfsRepo) FullVfsFile() OpFunc {
	return WithColumns(vr.join[Tables.VfsFile.Name]...)
}

// DefaultVfsFileSort returns default sort.
func (vr VfsRepo) DefaultVfsFileSort() OpFunc {
	return WithSort(vr.sort[Tables.VfsFile.Name]...)
}

// VfsFileByID is a function that returns VfsFile by ID(s) or nil.
func (vr VfsRepo) VfsFileByID(ctx context.Context, id int, ops ...OpFunc) (*VfsFile, error) {
	return vr.OneVfsFile(ctx, &VfsFileSearch{ID: &id}, ops...)
}

// OneVfsFile is a function that returns one VfsFile by filters. It could return pg.ErrMultiRows.
func (vr VfsRepo) OneVfsFile(ctx context.Context, search *VfsFileSearch, ops ...OpFunc) (*VfsFile, error) {
	obj := &VfsFile{}
	err := buildQuery(ctx, vr.db, obj, search, vr.filters[Tables.VfsFile.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// VfsFilesByFilters returns VfsFile list.
func (vr VfsRepo) VfsFilesByFilters(ctx context.Context, search *VfsFileSearch, pager Pager, ops ...OpFunc) (vfsFiles []VfsFile, err error) {
	err = buildQuery(ctx, vr.db, &vfsFiles, search, vr.filters[Tables.VfsFile.Name], pager, ops...).Select()
	return
}

// CountVfsFiles returns count
func (vr VfsRepo) CountVfsFiles(ctx context.Context, search *VfsFileSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, vr.db, &VfsFile{}, search, vr.filters[Tables.VfsFile.Name], PagerOne, ops...).Count()
}

// AddVfsFile adds VfsFile to DB.
func (vr VfsRepo) AddVfsFile(ctx context.Context, vfsFile *VfsFile, ops ...OpFunc) (*VfsFile, error) {
	q := vr.db.ModelContext(ctx, vfsFile)
	applyOps(q, ops...)
	_, err := q.ExcludeColumn(Columns.VfsFile.CreatedAt).Insert()

	return vfsFile, err
}

// UpdateVfsFile updates VfsFile in DB.
func (vr VfsRepo) UpdateVfsFile(ctx context.Context, vfsFile *VfsFile, ops ...OpFunc) (bool, error) {
	q := vr.db.ModelContext(ctx, vfsFile).WherePK()
	applyOps(q, ops...)
	res, err := q.ExcludeColumn(Columns.VfsFile.CreatedAt).Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteVfsFile set statusId to deleted in DB.
func (vr VfsRepo) DeleteVfsFile(ctx context.Context, id int) (deleted bool, err error) {
	vfsFile := &VfsFile{ID: id, StatusID: StatusDeleted}

	return vr.UpdateVfsFile(ctx, vfsFile, WithColumns(Columns.VfsFile.StatusID))
}

/*** VfsFolder ***/

// FullVfsFolder returns full joins with all columns
func (vr VfsRepo) FullVfsFolder() OpFunc {
	return WithColumns(vr.join[Tables.VfsFolder.Name]...)
}

// DefaultVfsFolderSort returns default sort.
func (vr VfsRepo) DefaultVfsFolderSort() OpFunc {
	return WithSort(vr.sort[Tables.VfsFolder.Name]...)
}

// VfsFolderByID is a function that returns VfsFolder by ID(s) or nil.
func (vr VfsRepo) VfsFolderByID(ctx context.Context, id int, ops ...OpFunc) (*VfsFolder, error) {
	return vr.OneVfsFolder(ctx, &VfsFolderSearch{ID: &id}, ops...)
}

// OneVfsFolder is a function that returns one VfsFolder by filters. It could return pg.ErrMultiRows.
func (vr VfsRepo) OneVfsFolder(ctx context.Context, search *VfsFolderSearch, ops ...OpFunc) (*VfsFolder, error) {
	obj := &VfsFolder{}
	err := buildQuery(ctx, vr.db, obj, search, vr.filters[Tables.VfsFolder.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// VfsFoldersByFilters returns VfsFolder list.
func (vr VfsRepo) VfsFoldersByFilters(ctx context.Context, search *VfsFolderSearch, pager Pager, ops ...OpFunc) (vfsFolders []VfsFolder, err error) {
	err = buildQuery(ctx, vr.db, &vfsFolders, search, vr.filters[Tables.VfsFolder.Name], pager, ops...).Select()
	return
}

// CountVfsFolders returns count
func (vr VfsRepo) CountVfsFolders(ctx context.Context, search *VfsFolderSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, vr.db, &VfsFolder{}, search, vr.filters[Tables.VfsFolder.Name], PagerOne, ops...).Count()
}

// AddVfsFolder adds VfsFolder to DB.
func (vr VfsRepo) AddVfsFolder(ctx context.Context, vfsFolder *VfsFolder, ops ...OpFunc) (*VfsFolder, error) {
	q := vr.db.ModelContext(ctx, vfsFolder)
	applyOps(q, ops...)
	_, err := q.ExcludeColumn(Columns.VfsFolder.CreatedAt).Insert()

	return vfsFolder, err
}

// UpdateVfsFolder updates VfsFolder in DB.
func (vr VfsRepo) UpdateVfsFolder(ctx context.Context, vfsFolder *VfsFolder, ops ...OpFunc) (bool, error) {
	q := vr.db.ModelContext(ctx, vfsFolder).WherePK()
	applyOps(q, ops...)
	res, err := q.ExcludeColumn(Columns.VfsFolder.CreatedAt).Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteVfsFolder set statusId to deleted in DB.
func (vr VfsRepo) DeleteVfsFolder(ctx context.Context, id int) (deleted bool, err error) {
	vfsFolder := &VfsFolder{ID: id, StatusID: StatusDeleted}

	return vr.UpdateVfsFolder(ctx, vfsFolder, WithColumns(Columns.VfsFolder.StatusID))
}
