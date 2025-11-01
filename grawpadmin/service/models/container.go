package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ServiceContainer struct {
	Uuid        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	DockerId    string    `json:"docker_id"`
	IsAvailable bool      `json:"is_available"`
}

func (sc *ServiceContainer) Scan(value any) error {
	return json.Unmarshal([]byte(value.(string)), sc)
}

func (sc *ServiceContainer) Value() (driver.Value, error) {
	b, err := json.Marshal(sc)
	return string(b), err
}

func ServiceContainerAdd(db *sql.DB, sc ...ServiceContainer) (int, error) {
	stmt, err := db.Prepare("INSERT INTO service_container(servicecontainer) VALUES(?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count int = 0
	var buffer driver.Value
	for _, model := range sc {
		if yes, err := ServiceContainerExists(db, model); yes {
			return count, fmt.Errorf("Container %s already exists", model.Name)
		} else if err != nil {
			return count, err
		}

		buffer, err = model.Value()
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

func ServiceContainerDel(db *sql.DB, sc ...ServiceContainer) (int, error) {
	stmt, err := db.Prepare("delete from service_container where servicecontainer->>'uuid' = ?")
	if err != nil {
		return 0, nil
	}
	defer stmt.Close()

	var count int = 0
	for _, model := range sc {
		_, err = stmt.Exec(&model)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

func ServiceContainerExists(db *sql.DB, sc ServiceContainer) (bool, error) {
	stmt, err := db.Prepare("SELECT COUNT(servicecontainer) FROM service_container WHERE servicecontainer->>'name' == ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(&sc.Name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type ServiceContainerFindOpts struct {
	Uuid     uuid.UUID
	Name     string
	DockerID string
	Limit    uint
}

func ServiceContainerFind(db *sql.DB, opts ServiceContainerFindOpts) ([]ServiceContainer, error) {
	var buf strings.Builder
	var sc ServiceContainer
	var scs []ServiceContainer
	var args []any
	buf.WriteString("SELECT servicecontainer FROM service_container")

	var cond []string
	if opts.Uuid != uuid.Nil {
		cond = append(cond, "servicecontainer->>'uuid' = ?")
		args = append(args, opts.Uuid)
	}
	if opts.Name != "" {
		cond = append(cond, "servicecontainer->>'name' = ?")
		args = append(args, opts.Name)
	}
	if opts.DockerID != "" {
		cond = append(cond, "servicecontainer->>'docker_id' = ?")
		args = append(args, opts.DockerID)
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
		return scs, err
	}
	defer stmt.Close()

	resp, err := stmt.Query(args...)
	if err != nil {
		return scs, err
	}

	for resp.Next() {
		err = resp.Scan(&sc)
		if err != nil {
			return scs, err
		}
		scs = append(scs, sc)
	}

	return scs, nil
}

func ServiceContainerPut(db *sql.DB, sc ...ServiceContainer) (int, error) {
	var count int = 0
	for _, model := range sc {
		// Add service image to database if it does not
		// exist yet.
		if yes, err := ServiceContainerExists(db, model); !yes {
			_, err := ServiceContainerAdd(db, model)
			if err != nil {
				return count, err
			}
			// Bail on any other error
		} else if err != nil {
			return count, err
			// Update an existing service image.
		} else {
			_, err := ServiceContainerUpdate(db, model)
			if err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}

func ServiceContainerUpdate(db *sql.DB, sc ...ServiceContainer) (int, error) {
	stmt, err := db.Prepare("UPDATE service_container SET servicecontainer = ? WHERE servicecontainer->>'name' = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var count int = 0
	var buffer driver.Value
	for _, model := range sc {
		buffer, err = model.Value()
		if err != nil {
			return count, err
		}
		_, err := stmt.Exec(buffer, model.Name)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

func ServiceContainerTableInit(db *sql.DB) error {
	return createServiceModelTable(db, ServiceModelOpts{
		TableName: "service_container",
		ModelName: "servicecontainer",
	})
}

type ServiceContainerNewOpts struct {
	Uuid     uuid.UUID
	Name     string
	DockerId string
}

// Create a new `ServiceContainer` model.
func ServiceContainerNew(opts ServiceContainerNewOpts) (ServiceContainer, error) {
	var sc ServiceContainer
	err := validateUuidOrGenerateNewUuid(&opts.Uuid)
	if err != nil {
		return sc, err
	}

	sc.Uuid = opts.Uuid
	sc.Name = opts.Name
	sc.DockerId = opts.DockerId
	sc.IsAvailable = true
	return sc, nil
}
