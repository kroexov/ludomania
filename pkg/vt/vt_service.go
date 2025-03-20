package vt

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"apisrv/pkg/db"
	"apisrv/pkg/embedlog"

	"github.com/vmkteam/zenrpc/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	zenrpc.Service
	embedlog.Logger
	commonRepo db.CommonRepo
}

var (
	errInvalidLoginPassword = zenrpc.NewStringError(http.StatusBadRequest, "invalid login or password")
)

func NewAuthService(dbo db.DB, logger embedlog.Logger) *AuthService {
	return &AuthService{
		Logger:     logger,
		commonRepo: db.NewCommonRepo(dbo),
	}
}

// Login authenticates user.
//
//zenrpc:login User login
//zenrpc:password User password
//zenrpc:remember Remember for week
//zenrpc:return User authentication key
//zenrpc:400 Invalid login or password
//zenrpc:500 Internal Error
func (s AuthService) Login(ctx context.Context, login, password string, remember bool) (string, error) {
	if login == "" || password == "" {
		return "", errInvalidLoginPassword
	}
	dbu, err := s.commonRepo.EnabledUserByLogin(ctx, login)
	if err != nil {
		return "", InternalError(err)
	} else if dbu == nil {
		return "", errInvalidLoginPassword
	}

	if ok := s.checkHash(password, dbu.Password); !ok {
		return "", errInvalidLoginPassword
	}

	if ok, err := s.commonRepo.AuthenticateUser(ctx, dbu, s.generateAuthKey(dbu, remember)); err != nil || !ok {
		return "", InternalError(err)
	}

	return dbu.AuthKey, nil
}

// Logout current user from every session
//
//zenrpc:return Successful logout
//zenrpc:401 Invalid authentication credentials
//zenrpc:500 Internal Error
func (s AuthService) Logout(ctx context.Context) (bool, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return false, ErrUnauthorized
	}

	if ok, err := s.commonRepo.AuthenticateUser(ctx, user, ""); err != nil || !ok {
		return false, InternalError(err)
	}

	return true, nil
}

// Profile is a function that returns current user profile
//
//zenrpc:return UserProfile
//zenrpc:401 Invalid authentication credentials
func (s AuthService) Profile(ctx context.Context) (*UserProfile, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return nil, ErrUnauthorized
	}

	return NewUserProfile(user), nil
}

// ChangePassword changes current user password.
//
//zenrpc:password New user password
//zenrpc:return New user authentication key
//zenrpc:401 Invalid authentication credentials
//zenrpc:500 Internal Error
func (s AuthService) ChangePassword(ctx context.Context, password string) (string, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return "", ErrUnauthorized
	}

	p, err := passwordHash(password)
	if err != nil {
		return "", InternalError(err)
	}
	user.Password = p
	user.AuthKey = s.generateAuthKey(user, false)

	if ok, err := s.commonRepo.UpdateUserPassword(ctx, user); err != nil || !ok {
		return "", InternalError(err)
	}
	return user.AuthKey, nil
}

// VfsAuthToken get auth token for VFS requests
func (s AuthService) VfsAuthToken(ctx context.Context) (string, error) {
	user := UserFromContext(ctx)
	return user.AuthKey, nil
}

func (s AuthService) checkHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s AuthService) generateAuthKey(u *db.User, remember bool) string {
	if remember {
		// valid only on current week
		key := fmt.Sprintf("%s:%s:%d:%d", u.Login, u.Password, u.ID, time.Now().Weekday())
		return fmt.Sprintf("%x", md5.Sum([]byte(key)))
	}
	return s.generateRandom(32)
}

func (s AuthService) generateRandom(length int) string {
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func passwordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

type UserService struct {
	zenrpc.Service
	embedlog.Logger
	commonRepo db.CommonRepo
}

func NewUserService(dbo db.DB, logger embedlog.Logger) *UserService {
	return &UserService{
		Logger:     logger,
		commonRepo: db.NewCommonRepo(dbo),
	}
}

func (s UserService) dbSort(ops *ViewOps) db.OpFunc {
	v := s.commonRepo.DefaultUserSort()
	if ops == nil {
		return v
	}

	switch ops.SortColumn {
	case db.Columns.User.ID, db.Columns.User.CreatedAt, db.Columns.User.Login, db.Columns.User.LastActivityAt, db.Columns.User.StatusID:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}

	return v
}

// Count Users according to conditions in search params
//
//zenrpc:search UserSearch
//zenrpc:return int
//zenrpc:500 Internal Error
func (s UserService) Count(ctx context.Context, search *UserSearch) (int, error) {
	count, err := s.commonRepo.CountUsers(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get а list of Users according to conditions in search params
//
//zenrpc:search UserSearch
//zenrpc:viewOps ViewOps
//zenrpc:return []UserSummary
//zenrpc:500 Internal Error
func (s UserService) Get(ctx context.Context, search *UserSearch, viewOps *ViewOps) ([]UserSummary, error) {
	list, err := s.commonRepo.UsersByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.commonRepo.FullUser())
	if err != nil {
		return nil, InternalError(err)
	}
	users := make([]UserSummary, 0, len(list))
	for i := 0; i < len(list); i++ {
		if user := NewUserSummary(&list[i]); user != nil {
			users = append(users, *NewUserSummary(&list[i]))
		}
	}
	return users, nil
}

// GetByID returns a User by its ID.
//
//zenrpc:id int
//zenrpc:return User
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s UserService) GetByID(ctx context.Context, id int) (*User, error) {
	db, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewUser(db), nil
}

func (s UserService) byID(ctx context.Context, id int) (*db.User, error) {
	db, err := s.commonRepo.UserByID(ctx, id, s.commonRepo.FullUser())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}

// Add a User from the query
//
//zenrpc:user User
//zenrpc:return User
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s UserService) Add(ctx context.Context, user User) (*User, error) {
	if ve := s.isValid(ctx, user, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	p, err := passwordHash(user.Password)
	if err != nil {
		return nil, InternalError(err)
	}

	u := user.ToDB()
	u.Password = p

	dbc, err := s.commonRepo.AddUser(ctx, u)
	if err != nil {
		return nil, InternalError(err)
	}
	return NewUser(dbc), nil
}

// Update updates the User data identified by id from the query
//
//zenrpc:users User
//zenrpc:return User
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s UserService) Update(ctx context.Context, user User) (bool, error) {
	orig, err := s.byID(ctx, user.ID)
	if err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, user, true); ve.HasErrors() {
		return false, ve.Error()
	}

	cur := user.ToDB()
	cur.Password = orig.Password
	cur.AuthKey = orig.AuthKey

	if user.Password != "" {
		p, err := passwordHash(user.Password)
		if err != nil {
			return false, InternalError(err)
		}
		cur.Password = p
		cur.AuthKey = ""
	}

	ok, err := s.commonRepo.UpdateUser(ctx, cur)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete deletes the User by its ID.
//
//zenrpc:id int
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s UserService) Delete(ctx context.Context, id int) (bool, error) {
	if _, err := s.byID(ctx, id); err != nil {
		return false, err
	}

	ok, err := s.commonRepo.DeleteUser(ctx, id)
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Validate Verifies that User data is valid.
//
//zenrpc:user User
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s UserService) Validate(ctx context.Context, user User) ([]FieldError, error) {
	isUpdate := user.ID != 0
	if isUpdate {
		_, err := s.byID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, user, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s UserService) isValid(ctx context.Context, user User, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, user); v.HasInternalError() {
		return v
	}

	//check login unique
	item, err := s.commonRepo.OneUser(ctx, &db.UserSearch{Login: &user.Login, NotID: &user.ID})
	if err != nil {
		v.SetInternalError(err)
	} else if item != nil {
		v.Append("login", FieldErrorUnique)
	}

	// check empty password for add
	if !isUpdate && user.Password == "" {
		v.Append("password", FieldErrorRequired)
	}

	return v
}
