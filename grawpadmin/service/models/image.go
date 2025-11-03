package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type ServiceImage struct {
	Uuid        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Tag         string    `json:"tag"`
	DockerID    string    `json:"docker_id"`
	IsAvailable bool      `json:"is_available"`
}

func (si *ServiceImage) Scan(value any) error {
	return json.Unmarshal([]byte(value.(string)), si)
}

func (si *ServiceImage) Value() (driver.Value, error) {
	b, err := json.Marshal(si)
	return string(b), err
}

// Attempt to add a new `ServiceImage` to the data cache.
//
// Fails if the image already exists or if there is an I/O
// error with the database.
func ServiceImageAdd(db *sql.DB, si ...ServiceImage) (int, error) {
	stmt, err := db.Prepare("INSERT INTO service_image(serviceimage) VALUES(?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count int = 0
	var buffer driver.Value
	for _, image := range si {
		if yes, err := ServiceImageExists(db, image); yes {
			return count, fmt.Errorf("Image %s:%s already exists", image.Name, image.Tag)
		} else if err != nil {
			return count, err
		}

		buffer, err = image.Value()
		if err != nil {
			return count, err
		}
		_, err = stmt.Exec(buffer)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// Attempt to delete a `ServiceImage` from the data cache.
//
// Fails if there is an I/O error with the database.
func ServiceImageDel(db *sql.DB, si ...ServiceImage) (int, error) {
	stmt, err := db.Prepare("delete from service_image where serviceimage->>'uuid' = ?")
	if err != nil {
		return 0, nil
	}
	defer stmt.Close()

	var count int = 0
	for _, image := range si {
		_, err = stmt.Exec(&image)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// The `ServiceImage` exists in the data cache.
//
// `ServiceImage` records are identified by their image
// `Name` and their image `Tag`.
func ServiceImageExists(db *sql.DB, si ServiceImage) (bool, error) {
	stmt, err := db.Prepare("SELECT COUNT(serviceimage) FROM service_image WHERE serviceimage->>'name' == ? AND serviceimage->>'tag' == ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(&si.Name, &si.Tag).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type ServiceImageFindOpts struct {
	Uuid     uuid.UUID
	Name     string
	DockerID string
	Tag      string
	Limit    uint
}

func ServiceImagesFind(db *sql.DB, opts ServiceImageFindOpts) ([]ServiceImage, error) {
	var buf strings.Builder
	var si ServiceImage
	var sis []ServiceImage
	var args []any
	buf.WriteString("SELECT serviceimage FROM service_image")

	var cond []string
	if opts.Uuid != uuid.Nil {
		cond = append(cond, "serviceimage->>'uuid' = ?")
		args = append(args, opts.Uuid)
	}
	if opts.Name != "" {
		cond = append(cond, "serviceimage->>'name' = ?")
		args = append(args, opts.Name)
	}
	if opts.DockerID != "" {
		cond = append(cond, "serviceimage->>'docker_id' = ?")
		args = append(args, opts.DockerID)
	}
	if opts.Tag != "" {
		cond = append(cond, "serviceimage->>'tag' = ?")
		args = append(args, opts.Tag)
	}

	if len(cond) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(cond, " AND "))
	}

	if opts.Limit != 0 {
		buf.WriteString(" LIMIT ?")
		args = append(args, opts.Limit)
	}

	stmt, err := db.Prepare(buf.String())
	if err != nil {
		return sis, err
	}
	defer stmt.Close()

	resp, err := stmt.Query(args...)
	if err != nil {
		return sis, err
	}

	for resp.Next() {
		err = resp.Scan(&si)
		if err != nil {
			return sis, err
		}
		sis = append(sis, si)
	}

	return sis, nil
}

// List all known `ServiceImage` records.
func ServiceImagesList(db *sql.DB) ([]ServiceImage, error) {
	var si ServiceImage
	var sis []ServiceImage
	var err error

	resp, err := db.Query("SELECT serviceimage FROM service_image")
	if err != nil {
		return sis, err
	}
	defer resp.Close()

	for resp.Next() {
		err = resp.Scan(&si)
		if err != nil {
			return sis, err
		}
		sis = append(sis, si)
	}
	return sis, nil
}

// Attempt to insert new `ServiceImage` records. If a
// record exists, updates the record instead.
func ServiceImagePut(db *sql.DB, si ...ServiceImage) (int, error) {
	var count int = 0
	for _, image := range si {
		// Add service image to database if it does not
		// exist yet.
		if yes, err := ServiceImageExists(db, image); !yes {
			_, err := ServiceImageAdd(db, image)
			if err != nil {
				return count, err
			}
			// Bail on any other error
		} else if err != nil {
			return count, err
			// Update an existing service image.
		} else {
			_, err := ServiceImageUpdate(db, image)
			if err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}

// Update one or many existing `ServiceImage` record.
func ServiceImageUpdate(db *sql.DB, si ...ServiceImage) (int, error) {
	stmt, err := db.Prepare("UPDATE service_image SET serviceimage = ? WHERE serviceimage->>'name' = ? AND serviceimage->>'tag' == ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count int = 0
	var buffer driver.Value
	for _, image := range si {
		buffer, err = image.Value()
		if err != nil {
			return count, err
		}
		_, err := stmt.Exec(buffer, image.Name, image.Tag)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// Initialize the `ServiceImage` table.
func ServiceImageTableInit(db *sql.DB) error {
	return createServiceModelTable(db, ServiceModelOpts{
		TableName: "service_image",
		ModelName: "serviceimage",
	})
}

type ServiceImageNewOptions struct {
	Uuid     uuid.UUID
	Name     string
	Tag      string
	DockerId string
}

// Create a new `ServiceImage` model.
func ServiceImageNew(opts ServiceImageNewOptions) (ServiceImage, error) {
	var si ServiceImage
	err := validateUuidOrGenerateNewUuid(&opts.Uuid)
	if err != nil {
		return si, err
	}

	si.Uuid = opts.Uuid
	si.Name = opts.Name
	si.Tag = opts.Tag
	si.DockerID = opts.DockerId
	si.IsAvailable = true
	return si, nil
}
