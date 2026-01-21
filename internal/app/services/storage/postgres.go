package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	common "ws_ingestor/internal/app/common/exception_handler"
	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/constants"
	"ws_ingestor/internal/app/dto"
	"ws_ingestor/internal/app/models"
	"ws_ingestor/internal/utils"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Store struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewPostgres(dbURL string) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, common.NewCustomError(common.ErrDBConnect, "Failed to connect to Postgres", err)
	}
	// Connection pooling
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	store := &Store{
		db:     db,
		logger: logger.GetLogger(),
	}
	if err := store.createTables(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) createTables() error {

	tableName := constants.MARKET_DATA_TABLE_NAME
	if tableName == "" {
		tableName = "market_data"
	}
	query := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			timestamp BIGINT NOT NULL,
			exchange VARCHAR(100),
			data JSONB
		)`
	if _, err := s.db.Exec(query); err != nil {
		return common.NewCustomError(common.ErrDBConnect, fmt.Sprintf("Failed to create table %s", tableName), err)
	} else {
		s.logger.Info(fmt.Sprintf("Ensured table %s exists", tableName))
	}

	clientsTable := constants.CLIENTS_CONFIGS_TABLE_NAME
	if clientsTable == "" {
		clientsTable = "clients_configs"
	}
	query = `CREATE TABLE IF NOT EXISTS ` + clientsTable + ` (
			id VARCHAR(255) PRIMARY KEY,
			config JSONB
		)`
	if _, err := s.db.Exec(query); err != nil {
		return common.NewCustomError(common.ErrDBConnect, fmt.Sprintf("Failed to create table %s", clientsTable), err)
	} else {
		s.logger.Info(fmt.Sprintf("Ensured table %s exists", clientsTable))
	}

	apiKeysTable := constants.API_KEYS_TABLE_NAME
	if apiKeysTable == "" {
		apiKeysTable = "api_keys"
	}
	query = `CREATE TABLE IF NOT EXISTS ` + apiKeysTable + ` (
			id SERIAL PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			key_hash VARCHAR(255) UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT true,
			last_used_at TIMESTAMP
		)`
	if _, err := s.db.Exec(query); err != nil {
		return common.NewCustomError(common.ErrDBConnect, fmt.Sprintf("Failed to create table %s", apiKeysTable), err)
	} else {
		s.logger.Info(fmt.Sprintf("Ensured table %s exists", apiKeysTable))
	}

	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) InsertBatch(ctx context.Context, batch []models.MarketData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to begin transaction: %v", err))
		return err
	}
	defer tx.Rollback() // Ensure rollback on error

	tableName := constants.MARKET_DATA_TABLE_NAME
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO `+tableName+` (name, timestamp, exchange, data) VALUES ($1,$2,$3,$4)`)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to prepare statement for table %s: %v", tableName, err))
		return err
	}
	defer stmt.Close()

	for _, record := range batch {
		dataBytes, _ := json.Marshal(record.Data)
		if record.Timestamp == 0 {
			continue // Skip entries with zero timestamp
		}
		if _, err := stmt.ExecContext(ctx, record.Name, record.Timestamp, record.Exchange, dataBytes); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to insert %s: %v", record.Name, err))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to commit transaction: %v", err))
		return err
	}
	return nil
}

func (s *Store) ValidateApiKey(ctx context.Context, apiKey string) (ClientID string, err error) {
	hash := utils.HashAPIKey(apiKey)

	var clientID string
	err1 := s.db.QueryRowContext(ctx, `
		SELECT client_id
		FROM api_keys
		WHERE key_hash = $1
		  AND is_active = true
	`, hash).Scan(&clientID)

	if err1 == sql.ErrNoRows {
		return "", errors.New("invalid api key")
	}
	if err1 != nil {
		return "", err1
	}

	//async update last_used_at
	go s.db.Exec(`
		UPDATE api_keys SET last_used_at = now()
		WHERE key_hash = $1
	`, hash)

	return clientID, nil
}

func (s *Store) GetClientConfig(ctx context.Context, clientID string) (*dto.ClientConfig, error) {
	tableName := constants.CLIENTS_CONFIGS_TABLE_NAME
	var configJSON []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT config
		FROM `+tableName+`
		WHERE id = $1
	`, clientID).Scan(&configJSON)
	if err == sql.ErrNoRows {
		return nil, nil // No config, use defaults
	}
	if err != nil {
		return nil, err
	}
	var config dto.ClientConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, err
	}
	fmt.Println("GetClientConfig =  =================================== ", config)
	return &config, nil
}
