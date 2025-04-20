package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type CommonRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewCommonRepo returns new repository
func NewCommonRepo(db orm.DB) CommonRepo {
	return CommonRepo{
		db:      db,
		filters: map[string][]Filter{},
		sort: map[string][]SortField{
			Tables.Ludoman.Name:     {{Column: Columns.Ludoman.ID, Direction: SortDesc}},
			Tables.Transaction.Name: {{Column: Columns.Transaction.CreatedAt, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.Ludoman.Name:     {TableColumns},
			Tables.Transaction.Name: {TableColumns, Columns.Transaction.FromLudoman, Columns.Transaction.ToLudoman},
		},
	}
}

// WithTransaction is a function that wraps CommonRepo with pg.Tx transaction.
func (cr CommonRepo) WithTransaction(tx *pg.Tx) CommonRepo {
	cr.db = tx
	return cr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (cr CommonRepo) WithEnabledOnly() CommonRepo {
	f := make(map[string][]Filter, len(cr.filters))
	for i := range cr.filters {
		f[i] = make([]Filter, len(cr.filters[i]))
		copy(f[i], cr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	cr.filters = f

	return cr
}

/*** Ludoman ***/

// FullLudoman returns full joins with all columns
func (cr CommonRepo) FullLudoman() OpFunc {
	return WithColumns(cr.join[Tables.Ludoman.Name]...)
}

// DefaultLudomanSort returns default sort.
func (cr CommonRepo) DefaultLudomanSort() OpFunc {
	return WithSort(cr.sort[Tables.Ludoman.Name]...)
}

// LudomanByID is a function that returns Ludoman by ID(s) or nil.
func (cr CommonRepo) LudomanByID(ctx context.Context, id int, ops ...OpFunc) (*Ludoman, error) {
	return cr.OneLudoman(ctx, &LudomanSearch{ID: &id}, ops...)
}

// OneLudoman is a function that returns one Ludoman by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneLudoman(ctx context.Context, search *LudomanSearch, ops ...OpFunc) (*Ludoman, error) {
	obj := &Ludoman{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Ludoman.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// LudomenByFilters returns Ludoman list.
func (cr CommonRepo) LudomenByFilters(ctx context.Context, search *LudomanSearch, pager Pager, ops ...OpFunc) (ludomen []Ludoman, err error) {
	err = buildQuery(ctx, cr.db, &ludomen, search, cr.filters[Tables.Ludoman.Name], pager, ops...).Select()
	return
}

// CountLudomen returns count
func (cr CommonRepo) CountLudomen(ctx context.Context, search *LudomanSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Ludoman{}, search, cr.filters[Tables.Ludoman.Name], PagerOne, ops...).Count()
}

// AddLudoman adds Ludoman to DB.
func (cr CommonRepo) AddLudoman(ctx context.Context, ludoman *Ludoman, ops ...OpFunc) (*Ludoman, error) {
	q := cr.db.ModelContext(ctx, ludoman)
	applyOps(q, ops...)
	_, err := q.Insert()

	return ludoman, err
}

// UpdateLudoman updates Ludoman in DB.
func (cr CommonRepo) UpdateLudoman(ctx context.Context, ludoman *Ludoman, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, ludoman).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Ludoman.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteLudoman deletes Ludoman from DB.
func (cr CommonRepo) DeleteLudoman(ctx context.Context, id int) (deleted bool, err error) {
	ludoman := &Ludoman{ID: id}

	res, err := cr.db.ModelContext(ctx, ludoman).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** Transaction ***/

// FullTransaction returns full joins with all columns
func (cr CommonRepo) FullTransaction() OpFunc {
	return WithColumns(cr.join[Tables.Transaction.Name]...)
}

// DefaultTransactionSort returns default sort.
func (cr CommonRepo) DefaultTransactionSort() OpFunc {
	return WithSort(cr.sort[Tables.Transaction.Name]...)
}

// TransactionByID is a function that returns Transaction by ID(s) or nil.
func (cr CommonRepo) TransactionByID(ctx context.Context, id int, ops ...OpFunc) (*Transaction, error) {
	return cr.OneTransaction(ctx, &TransactionSearch{ID: &id}, ops...)
}

// OneTransaction is a function that returns one Transaction by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneTransaction(ctx context.Context, search *TransactionSearch, ops ...OpFunc) (*Transaction, error) {
	obj := &Transaction{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Transaction.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// TransactionsByFilters returns Transaction list.
func (cr CommonRepo) TransactionsByFilters(ctx context.Context, search *TransactionSearch, pager Pager, ops ...OpFunc) (transactions []Transaction, err error) {
	err = buildQuery(ctx, cr.db, &transactions, search, cr.filters[Tables.Transaction.Name], pager, ops...).Select()
	return
}

// CountTransactions returns count
func (cr CommonRepo) CountTransactions(ctx context.Context, search *TransactionSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Transaction{}, search, cr.filters[Tables.Transaction.Name], PagerOne, ops...).Count()
}

// AddTransaction adds Transaction to DB.
func (cr CommonRepo) AddTransaction(ctx context.Context, transaction *Transaction, ops ...OpFunc) (*Transaction, error) {
	q := cr.db.ModelContext(ctx, transaction)
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Transaction.CreatedAt)
	}
	applyOps(q, ops...)
	_, err := q.Insert()

	return transaction, err
}

// UpdateTransaction updates Transaction in DB.
func (cr CommonRepo) UpdateTransaction(ctx context.Context, transaction *Transaction, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, transaction).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Transaction.ID, Columns.Transaction.CreatedAt)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteTransaction deletes Transaction from DB.
func (cr CommonRepo) DeleteTransaction(ctx context.Context, id int) (deleted bool, err error) {
	transaction := &Transaction{ID: id}

	res, err := cr.db.ModelContext(ctx, transaction).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}
