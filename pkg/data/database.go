package data

import (
	"os"
	"sync"

	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
	dbMu       sync.Mutex
)

// GetDatabaseURL returns database URL from environment or default SQLite path.
func GetDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "skills.db"
}

// InitDB initializes the GORM database connection and migrates models.
func InitDB(databaseURL ...string) (*gorm.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	url := GetDatabaseURL()
	if len(databaseURL) > 0 && databaseURL[0] != "" {
		url = databaseURL[0]
	}

	db, err := gorm.Open(sqlite.Open(url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.RegisteredApp{},
		&model.Skill{},
		&model.SkillVersion{},
		&model.SkillMetadata{},
		&model.SkillResource{},
		&model.SkillExample{},
		&model.SkillEmbedding{},
	)
	if err != nil {
		return nil, err
	}

	dbInstance = db
	return db, nil
}

// GetDB returns the initialized GORM DB instance.
func GetDB() *gorm.DB {
	dbMu.Lock()
	defer dbMu.Unlock()

	if dbInstance == nil {
		_, _ = InitDB()
	}
	return dbInstance
}

// ResetEngine resets the cached database instance (useful in unit testing).
func ResetEngine() {
	dbMu.Lock()
	defer dbMu.Unlock()
	if dbInstance != nil {
		sqlDB, err := dbInstance.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		dbInstance = nil
	}
}
