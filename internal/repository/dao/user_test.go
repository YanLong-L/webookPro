package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testcases := []struct {
		name    string
		ctx     context.Context
		user    User
		wantErr error
		sqlmock func(t *testing.T) *sql.DB
	}{
		{
			name: "插入数据库成功",
			ctx:  context.Background(),
			user: User{
				Email: sql.NullString{
					Valid:  true,
					String: "123@qq.com",
				},
			},
			wantErr: nil,
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(1, 1)
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
		},
		{
			name:    "邮箱冲突，插入数据库失败",
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrUserDuplicateEmail,
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				require.NoError(t, err)
				return mockDB
			},
		},
		{
			name:    "数据库错误",
			ctx:     context.Background(),
			user:    User{},
			wantErr: errors.New("数据库错误"),
			sqlmock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思就是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("数据库错误"))
				require.NoError(t, err)
				return mockDB
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: tc.sqlmock(t),
				// SELECT VERSION;
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 你 mock DB 不需要 ping
				DisableAutomaticPing: true,
				// 这个是什么呢？
				SkipDefaultTransaction: true,
			})
			d := NewGormUserDAO(db)
			u := tc.user
			err = d.Insert(tc.ctx, u)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
