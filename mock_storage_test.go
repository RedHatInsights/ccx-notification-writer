/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main_test

// Mock storage implementation that is used in unit tests.
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/ccx-notification-writer/packages/mock_storage_test.html

import (
	"time"

	main "github.com/RedHatInsights/ccx-notification-writer"
)

// MockStorage structure represents a mocked implementation of Storage
// interface to be used by tested consumer
type MockStorage struct {
	databaseInitMigrationCalled     int
	databaseInitializationCalled    int
	databaseCleanupCalled           int
	databaseDropTablesCalled        int
	databaseDropIndexesCalled       int
	getLatestKafkaOffsetCalled      int
	printNewReportsForCleanupCalled int
	cleanupNewReportsCalled         int
	printOldReportsForCleanupCalled int
	cleanupOldReportsCalled         int
	printReadErrorsForCleanupCalled int
	cleanupReadReportsCalled        int
	closeCalled                     int
	writeReportCalled               int
}

// NewMockStorage constructs new mock storage instance
func NewMockStorage() MockStorage {
	return MockStorage{
		databaseInitMigrationCalled:     0,
		databaseInitializationCalled:    0,
		databaseCleanupCalled:           0,
		databaseDropTablesCalled:        0,
		databaseDropIndexesCalled:       0,
		getLatestKafkaOffsetCalled:      0,
		printNewReportsForCleanupCalled: 0,
		cleanupNewReportsCalled:         0,
		printOldReportsForCleanupCalled: 0,
		cleanupOldReportsCalled:         0,
		printReadErrorsForCleanupCalled: 0,
		cleanupReadReportsCalled:        0,
		closeCalled:                     0,
		writeReportCalled:               0,
	}
}

// Close is a mocked reimplementation of the real Close method.
func (storage *MockStorage) Close() error {
	storage.closeCalled++

	// return no error
	return nil
}

// WriteReportForCluster is a mocked reimplementation of the real
// WriteReportForCluster method.
func (storage *MockStorage) WriteReportForCluster(orgID main.OrgID, accountNumber main.AccountNumber, clusterName main.ClusterName, report main.ClusterReport, collectedAtTime time.Time, kafkaOffset main.KafkaOffset) error {
	storage.writeReportCalled++

	// return no error
	return nil
}

// DatabaseInitialization is a mocked reimplementation of the real
// DatabaseInitialization method.
func (storage *MockStorage) DatabaseInitialization() error {
	storage.databaseInitializationCalled++

	// return no error
	return nil
}

// DatabaseInitMigration is a mocked reimplementation of the real
// DatabaseInitMigration method.
func (storage *MockStorage) DatabaseInitMigration() error {
	storage.databaseInitMigrationCalled++

	// return no error
	return nil
}

// DatabaseCleanup is a mocked reimplementation of the real DatabaseCleanup
// method.
func (storage *MockStorage) DatabaseCleanup() error {
	storage.databaseCleanupCalled++

	// return no error
	return nil
}

// DatabaseDropTables is a mocked reimplementation of the real
// DatabaseDropTables method.
func (storage *MockStorage) DatabaseDropTables() error {
	storage.databaseDropTablesCalled++

	// return no error
	return nil
}

// DatabaseDropIndexes is a mocked reimplementation of the real
// DatabaseDropIndexes method.
func (storage *MockStorage) DatabaseDropIndexes() error {
	storage.databaseDropIndexesCalled++

	// return no error
	return nil
}

// GetLatestKafkaOffset is a mocked reimplementation of the real
// GetLatestKafkaOffset method.
func (storage *MockStorage) GetLatestKafkaOffset() (main.KafkaOffset, error) {
	storage.getLatestKafkaOffsetCalled++

	// return some offset + no error
	return 1, nil
}

// PrintNewReportsForCleanup is a mocked reimplementation of the real
// PrintNewReportsForCleanup method.
func (storage *MockStorage) PrintNewReportsForCleanup(maxAge string) error {
	storage.printNewReportsForCleanupCalled++

	// return no error
	return nil
}

// CleanupNewReports is a mocked reimplementation of the real CleanupNewReports
// method.
func (storage *MockStorage) CleanupNewReports(maxAge string) (int, error) {
	storage.cleanupNewReportsCalled++

	// return number of cleaned records + no error
	return 1, nil
}

// PrintOldReportsForCleanup is a mocked reimplementation of the real
// PrintOldReportsForCleanup method.
func (storage *MockStorage) PrintOldReportsForCleanup(maxAge string) error {
	storage.printOldReportsForCleanupCalled++

	// return no error
	return nil
}

// CleanupOldReports is a mocked reimplementation of the real CleanupOldReports
// method.
func (storage *MockStorage) CleanupOldReports(maxAge string) (int, error) {
	storage.cleanupOldReportsCalled++

	// return number of cleaned records + no error
	return 1, nil
}

// PrintReadErrorsForCleanup is a mocked reimplementation of the real
// PrintReadErrorsForCleanup method.
func (storage *MockStorage) PrintReadErrorsForCleanup(maxAge string) error {
	storage.printReadErrorsForCleanupCalled++

	// return no error
	return nil
}

// CleanupReadErrors is a mocked reimplementation of the real CleanupReadErrors
// method.
func (storage *MockStorage) CleanupReadErrors(maxAge string) (int, error) {
	storage.cleanupOldReportsCalled++

	// return number of cleaned records + no error
	return 1, nil
}
