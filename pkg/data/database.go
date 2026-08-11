// Copyright 2026 Ryan McGuinness
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package data

import (
	"os"
	"strings"
	"sync"

	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/driver/postgres"
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
	return "castor.db"
}

// InitDB initializes the GORM database connection and migrates models.
func InitDB(databaseURL ...string) (*gorm.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	url := GetDatabaseURL()
	if len(databaseURL) > 0 && databaseURL[0] != "" {
		url = databaseURL[0]
	}

	isPostgres := strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") ||
		strings.Contains(url, "host=") || strings.Contains(url, "user=") || strings.Contains(url, "port=")

	var dialector gorm.Dialector
	if isPostgres {
		dialector = postgres.Open(url)
	} else {
		dialector = sqlite.Open(url)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if db.Migrator().HasTable(&model.Skill{}) {
		// Clean up duplicate legacy rows keeping only newest record per (app_id, category, name)
		if isPostgres {
			_ = db.Exec(`
				DELETE FROM skills a USING skills b 
				WHERE a.created_at < b.created_at 
				  AND a.app_id = b.app_id 
				  AND COALESCE(a.category, 'general') = COALESCE(b.category, 'general') 
				  AND a.name = b.name;
			`)
		}
	}

	err = db.AutoMigrate(
		&model.RegisteredApp{},
		&model.AppMember{},
		&model.AppAPIKey{},
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

	// For PostgreSQL and AlloyDB, initialize vector/alloydb_scann/google_ml extensions and HNSW indexes
	if isPostgres {
		_ = db.Exec("CREATE EXTENSION IF NOT EXISTS vector;")
		_ = db.Exec("CREATE EXTENSION IF NOT EXISTS alloydb_scann;")
		_ = db.Exec("CREATE EXTENSION IF NOT EXISTS google_ml CASCADE;")
		_ = db.Exec("ALTER TABLE skill_embeddings ADD COLUMN IF NOT EXISTS embedding_768 vector(768);")
		_ = db.Exec("ALTER TABLE skill_embeddings ADD COLUMN IF NOT EXISTS embedding_1408 vector(1408);")
		_ = db.Exec("ALTER TABLE skill_embeddings ADD COLUMN IF NOT EXISTS embedding_3072 vector(3072);")
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_emb_hnsw_768 ON skill_embeddings USING hnsw (embedding_768 vector_cosine_ops);")
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_emb_hnsw_1408 ON skill_embeddings USING hnsw (embedding_1408 vector_cosine_ops);")
		_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_emb_hnsw_3072 ON skill_embeddings USING hnsw (embedding_3072 vector_cosine_ops);")

		// Automatically backfill existing records from canonical embedding_json into typed vector columns
		_ = db.Exec("UPDATE skill_embeddings SET embedding_1408 = embedding_json::vector WHERE dimension = 1408 AND embedding_1408 IS NULL;")
		_ = db.Exec("UPDATE skill_embeddings SET embedding_768 = embedding_json::vector WHERE dimension = 768 AND embedding_768 IS NULL;")
		_ = db.Exec("UPDATE skill_embeddings SET embedding_3072 = embedding_json::vector WHERE dimension = 3072 AND embedding_3072 IS NULL;")
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
