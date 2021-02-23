package databases

import (
	"fmt"
	"time"

	viper "github.com/spf13/viper"
	postgres "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
)

type (
	// Base struct
	Base struct {
		ID           uint      `gorm:"primary_key" json:"id"`                              //
		ModifiedDate time.Time `gorm:"column:modified_date;not null" json:"modified_date"` // Өөрчилсөн огноо
		CreatedDate  time.Time `gorm:"column:created_date;not null" json:"created_date"`   // Үүсгэсэн огноо
	}
)

// InitDB initialize databases and tables
func InitDB() *gorm.DB {
	DBUser := viper.GetString("database.user")
	DBPassword := viper.GetString("database.password")
	DBDatabase := viper.GetString("database.name")
	DBHost := viper.GetString("database.host")
	DBPort := viper.GetString("database.port")
	DBTimezone := viper.GetString("database.timezone")
	ConnectionString := fmt.Sprintf(
		"host=%v port=%v user=%v dbname=%v password=%v sslmode=disable TimeZone=%v",
		DBHost,
		DBPort,
		DBUser,
		DBDatabase,
		DBPassword,
		DBTimezone,
	)

	fmt.Println("ConnectionString", ConnectionString)

	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold: time.Second,   // Slow SQL threshold
	// 		LogLevel:      logger.Silent, // Log level
	// 		Colorful:      false,         // Disable color
	// 	},
	// )

	db, err := gorm.Open(postgres.Open(ConnectionString), &gorm.Config{
		// Logger:                                   newLogger,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		panic(err.Error())
	}
	db.AutoMigrate(
		&SystemUser{},
		&Company{},
		&AwsCredentials{},
		&ConfirmUser{},
	)
	return db
}

// func (u *MedSystemUser) BeforeCreate(tx *gorm.DB) (err error) {
// 	u.UUID = uuid.New()

// 	if !u.IsValid() {
// 		err = errors.New("can't save invalid data")
// 	}
// 	return
// }
