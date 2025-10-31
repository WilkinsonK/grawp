package service_models

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

func ServiceImageTableInit(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE service_image (serviceimage jsonb)")
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

type ServiceImageNewOptions struct {
	Uuid     uuid.UUID
	Name     string
	Tag      string
	DockerId string
}

func ServiceImageNew(opts ServiceImageNewOptions) (ServiceImage, error) {
	validateUuidOrGenerateNewUuid(&opts.Uuid)

	var si ServiceImage
	si.Uuid = opts.Uuid
	si.Name = opts.Name
	si.Tag = opts.Tag
	si.DockerID = opts.DockerId
	si.IsAvailable = true
	return si, nil
}
