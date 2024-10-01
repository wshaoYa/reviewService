package data

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reviewService/internal/conf"
	"reviewService/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewReviewRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	q   *query.Query
	log *log.Helper
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}

	query.SetDefault(db)
	return &Data{
		q:   query.Q,
		log: log.NewHelper(logger),
	}, cleanup, nil
}

// NewDB 新建DB
func NewDB(c *conf.Data) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(c.Database.GetSource()), &gorm.Config{})
}
