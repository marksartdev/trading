package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/marksartdev/trading/internal/broker"
)

// OHLCV entity.
type OHLCV struct {
	ID       int64         `gorm:"primarykey"`
	Ticker   string        `gorm:"not null"`
	Time     time.Time     `gorm:"not null"`
	Interval time.Duration `gorm:"not null"`
	Open     float64       `gorm:"not null"`
	High     float64       `gorm:"not null"`
	Low      float64       `gorm:"not null"`
	Close    float64       `gorm:"not null"`
	Volume   int32         `gorm:"not null"`
}

const limit = 300

// Statistic repository.
type statisticRepo struct {
	db *gorm.DB
}

// NewStatisticRepo creates new statistic repository.
func NewStatisticRepo(db *gorm.DB) broker.StatisticRepo {
	return statisticRepo{db: db}
}

// Add adds OHLCV to repository.
func (s statisticRepo) Add(ohlcv broker.OHLCV) error {
	entity := OHLCV{
		ID:       ohlcv.ID,
		Ticker:   ohlcv.Ticker,
		Time:     ohlcv.Time,
		Interval: ohlcv.Interval,
		Open:     ohlcv.Open,
		High:     ohlcv.High,
		Low:      ohlcv.Low,
		Close:    ohlcv.Close,
		Volume:   ohlcv.Volume,
	}

	return s.db.Create(&entity).Error
}

// Get returns statistic for last n sec from repository.
func (s statisticRepo) Get(ticker string) ([]broker.OHLCV, error) {
	var entities []OHLCV

	err := s.db.Where(OHLCV{Ticker: ticker}).Order("time DESC").Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	statistic := make([]broker.OHLCV, len(entities))
	for i := range statistic {
		statistic[i] = broker.OHLCV{
			ID:       entities[i].ID,
			Ticker:   entities[i].Ticker,
			Time:     entities[i].Time,
			Interval: entities[i].Interval,
			Open:     entities[i].Open,
			High:     entities[i].High,
			Low:      entities[i].Low,
			Close:    entities[i].Close,
			Volume:   entities[i].Volume,
		}
	}

	return statistic, nil
}
