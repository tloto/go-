package model

import (
	"time"

	"git.g7n3.com/hamster/gautu/msg"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"primary_key;type:varchar(36)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

//type Model map[string]string
type models map[string]interface{}

var (
	m     models = make(models)
	CFlag bool
)

func init() {
	CFlag = true
}

func (b *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	if CFlag {
		scope.SetColumn("ID", uuid.NewV1())
	}
	return nil
}

func Gorm() *gorm.DB {
	return db
}

func GetModels() *models {
	return &m
}

func AddModel(n string, v interface{}) {
	m[n] = v
}

func Save(v interface{}) *gorm.DB {
	db := Gorm().Save(v)
	msg.Errors(db.GetErrors())
	return db
}

//
//func SetMigrate(b bool) {
//	isMigrate = b
//}
//
//func IsMigrate() bool {
//	return isMigrate
//}
func IsNull(id2 uuid.UUID) bool {
	return uuid.Equal(uuid.Nil, id2)
}
