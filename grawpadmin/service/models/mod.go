package models

import (
	"bytes"
	"database/sql"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type ServiceModelOpts struct {
	TableName string
	ModelName string
}

func createServiceModelTable(db *sql.DB, opts ServiceModelOpts) error {
	templ, err := template.New("create-service-model").Parse("CREATE TABLE {{.TableName}} ({{.ModelName}} jsonb)")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, opts); err != nil {
		return err
	}

	_, err = db.Exec(buf.String())
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

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
	if err := ServiceImageTableInit(db); err != nil {
		return err
	}
	if err := ServiceContainerTableInit(db); err != nil {
		return err
	}
	return nil
}
