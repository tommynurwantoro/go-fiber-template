package service

import (
	"app/internal/adapter/database"
	"errors"
	"fmt"
	"runtime"

	"github.com/tommynurwantoro/golog"
)

type HealthCheckService interface {
	GormCheck() error
	MemoryHeapCheck() error
}

type HealthCheckServiceImpl struct {
	DB database.DatabaseAdapter `inject:"database"`
}

func (s *HealthCheckServiceImpl) GormCheck() error {
	sqlDB, errDB := s.DB.GetDB().DB()
	if errDB != nil {
		golog.Error("failed to access the database connection pool: %v", errDB)
		return errDB
	}

	if err := sqlDB.Ping(); err != nil {
		golog.Error("failed to ping the database: %v", err)
		return err
	}

	return nil
}

// MemoryHeapCheck checks if heap memory usage exceeds a threshold
func (s *HealthCheckServiceImpl) MemoryHeapCheck() error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats) // Collect memory statistics

	heapAlloc := memStats.HeapAlloc            // Heap memory currently allocated
	heapThreshold := uint64(300 * 1024 * 1024) // Example threshold: 300 MB

	golog.Info(fmt.Sprintf("Heap Memory Allocation: %v bytes", heapAlloc))

	// If the heap allocation exceeds the threshold, return an error
	if heapAlloc > heapThreshold {
		golog.Error(fmt.Sprintf("Heap memory usage exceeds threshold: %v bytes", heapAlloc), nil)
		return errors.New("heap memory usage too high")
	}

	return nil
}
