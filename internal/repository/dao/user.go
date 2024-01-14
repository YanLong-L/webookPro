package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var ErrUserDuplicateEmail error = errors.New("邮箱冲突")

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx *gin.Context, phone string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewGormUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

func (ud *GormUserDAO) Insert(ctx context.Context, u User) error {
	// 获取当前时间，补充 Utime和 Ctime
	now := time.Now().UnixMilli() // 毫秒数时间戳
	u.Utime = now
	u.Ctime = now
	// 插入数据库
	err := ud.db.WithContext(ctx).Create(&u).Error
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		const uniqueIndexErrNo uint16 = 1062
		if mysqlErr.Number == uniqueIndexErrNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (ud *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return User{}, err
	}
	return u, err
}

func (ud *GormUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		return User{}, err
	}
	return u, err
}

func (ud *GormUserDAO) FindByPhone(ctx *gin.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if err != nil {
		return User{}, err
	}
	return u, err
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string
	Ctime    int64
	Utime    int64
	Phone    sql.NullString `gorm:"unique"` // 和 email一样是唯一索引，但是null值不冲突
}
