package repository

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/models"
)

type DeviceRepository interface {
	Create(device *models.Device) error
	GetByID(id string) (*models.Device, error)
	GetAll() ([]*models.Device, error)
	Update(device *models.Device) error
	Delete(id string) error
	// For sync: get device by IP
	GetByIP(ip string) (*models.Device, error)
	UpdateLastSeen(id string) error
}

type deviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(device *models.Device) error {
	query := `INSERT INTO devices (id, local_ip, hostname, last_seen, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, device.ID, device.LocalIP, device.Hostname, device.LastSeen, device.CreatedAt)
	return err
}

func (r *deviceRepository) GetByID(id string) (*models.Device, error) {
	var d models.Device
	query := `SELECT id, local_ip, hostname, last_seen, created_at FROM devices WHERE id = ?`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&d.ID, &d.LocalIP, &d.Hostname, &d.LastSeen, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deviceRepository) GetAll() ([]*models.Device, error) {
	rows, err := r.db.Query(`SELECT id, local_ip, hostname, last_seen, created_at FROM devices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []*models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.LocalIP, &d.Hostname, &d.LastSeen, &d.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}
	return devices, rows.Err()
}

func (r *deviceRepository) Update(device *models.Device) error {
	query := `UPDATE devices SET local_ip = ?, hostname = ?, last_seen = ? WHERE id = ?`
	_, err := r.db.Exec(query, device.LocalIP, device.Hostname, device.LastSeen, device.ID)
	return err
}

func (r *deviceRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM devices WHERE id = ?`, id)
	return err
}

func (r *deviceRepository) GetByIP(ip string) (*models.Device, error) {
	var d models.Device
	query := `SELECT id, local_ip, hostname, last_seen, created_at FROM devices WHERE local_ip = ?`
	row := r.db.QueryRow(query, ip)
	err := row.Scan(&d.ID, &d.LocalIP, &d.Hostname, &d.LastSeen, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deviceRepository) UpdateLastSeen(id string) error {
	query := `UPDATE devices SET last_seen = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
