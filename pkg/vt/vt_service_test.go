package vt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"apisrv/pkg/db"
	"apisrv/pkg/embedlog"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDB_AuthService(t *testing.T) {
	Convey("Test AuthService", t, func() {
		ctx := context.Background()
		srv := NewAuthService(testDb, embedlog.Logger{})
		So(srv, ShouldNotBeNil)

		Convey("Positive testing", func() {

			Convey("Login method with remember password", func() {
				authKey, err := srv.Login(ctx, "admin", "12345", true)
				So(err, ShouldBeNil)
				authKey2, err := srv.Login(ctx, "admin", "12345", true)
				So(err, ShouldBeNil)
				So(authKey, ShouldEqual, authKey2)
			})

			Convey("Login without remember password", func() {
				authKey, err := srv.Login(ctx, "admin", "12345", false)
				So(err, ShouldBeNil)
				So(authKey, ShouldHaveLength, 32)

				u, err := srv.commonRepo.EnabledUserByAuthKey(ctx, authKey)
				So(err, ShouldBeNil)
				ctx = context.WithValue(ctx, userKey, u)

				Convey("Get profile", func() {
					user, err := srv.Profile(ctx)
					So(err, ShouldBeNil)
					So(user, ShouldNotBeNil)
				})

				Convey("Logout", func() {
					ok, err := srv.Logout(ctx)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
				})
			})
		})

		Convey("Negative testing", func() {

			Convey("Login not exists", func() {
				_, err := srv.Login(ctx, "vova", "12345", false)
				So(err, ShouldBeError)
			})

			Convey("Wrong password", func() {
				_, err := srv.Login(ctx, "admin", "admin", false)
				So(err, ShouldBeError)
			})

			Convey("Empty login/password", func() {
				_, err := srv.Login(ctx, "", "", false)
				So(err, ShouldBeError)
			})

			Convey("Profile without user in context", func() {
				ctx := context.Background()
				u, err := srv.Profile(ctx)
				So(err, ShouldBeError)
				So(u, ShouldBeNil)
			})

			Convey("Logout without user in context", func() {
				ctx := context.Background()
				ok, err := srv.Logout(ctx)
				So(err, ShouldBeError)
				So(ok, ShouldBeFalse)
			})

		})
	})
}

func TestDB_UserService(t *testing.T) {
	Convey("Test UserService", t, func() {
		ctx := context.Background()
		srv := NewUserService(testDb, embedlog.Logger{})
		So(srv, ShouldNotBeNil)

		Convey("Positive testing", func() {

			Convey("Test CRUD", func() {
				login := fmt.Sprintf("ivan_%d", time.Now().Unix())
				password := fmt.Sprintf("pwd_%v", login)

				inUser := User{
					Login:     login,
					Password:  password,
					StatusID:  db.StatusEnabled,
					CreatedAt: time.Now(),
				}

				// Add
				outUser, err := srv.Add(ctx, inUser)
				So(err, ShouldBeNil)
				So(outUser, ShouldNotBeNil)

				So(outUser.ID, ShouldBeGreaterThan, 0)
				So(outUser.Login, ShouldEqual, inUser.Login)
				So(outUser.Password, ShouldNotEqual, "")

				// GetByID
				u, err := srv.GetByID(ctx, outUser.ID)
				So(err, ShouldBeNil)
				So(u, ShouldNotBeNil)

				So(u.ID, ShouldEqual, outUser.ID)
				So(u.Login, ShouldEqual, outUser.Login)
				So(u.Password, ShouldEqual, outUser.Password)

				// Update
				u.Login = "test"

				ok, err := srv.Update(ctx, *u)
				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)

				updated, err := srv.GetByID(ctx, outUser.ID)
				So(err, ShouldBeNil)
				So(updated.Login, ShouldEqual, u.Login)

				// Delete
				ok, err = srv.Delete(ctx, outUser.ID)
				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)
			})
		})

		Convey("Negative testing", func() {

			Convey("Create user with empty login", func() {
				user := User{
					Login:    "",
					Password: "password",
					StatusID: db.StatusEnabled,
				}
				u, err := srv.Add(ctx, user)
				So(err, ShouldNotBeNil)
				So(u, ShouldBeNil)
			})

			Convey("Create user with empty password", func() {
				user := User{
					Login:    "vasya",
					Password: "",
					StatusID: db.StatusEnabled,
				}
				u, err := srv.Add(ctx, user)
				So(err, ShouldNotBeNil)
				So(u, ShouldBeNil)
			})

			Convey("Create user with duplicate login", func() {
				user := User{
					Login:     "unique",
					Password:  "unique2",
					StatusID:  db.StatusEnabled,
					CreatedAt: time.Now(),
				}
				u, err := srv.Add(ctx, user)
				So(err, ShouldBeNil)
				So(u, ShouldNotBeNil)

				u2, err := srv.Add(ctx, user)
				So(err, ShouldNotBeNil)
				So(u2, ShouldBeNil)
			})
		})

	})
}
