package service_models

import (
	"database/sql"

	"github.com/google/uuid"
)

func validateUuidOrGenerateNewUuid(u *uuid.UUID) error {
	if *u == uuid.Nil {
		if newUuid, err := uuid.NewUUID(); err != nil {
			return err
		} else {
			*u = newUuid
		}
	}
	return nil
}

func InitDatabaseTables(db *sql.DB) error {
	return ServiceImageTableInit(db)
}
