package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Vehicles struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Make string `json:"make"`
	Model string `json:"model"`
	Year int64 `json:"year"`
	Vin string `json:"vin"`
	LicensePlate string `json:"license_plate"`
	Mileage int64 `json:"mileage"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type Maintenance struct {
	ID string `json:"id"`
	VehicleId string `json:"vehicle_id"`
	ServiceType string `json:"service_type"`
	Date string `json:"date"`
	Mileage int64 `json:"mileage"`
	Cost float64 `json:"cost"`
	Provider string `json:"provider"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "fleet.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS vehicles(id TEXT PRIMARY KEY, name TEXT NOT NULL, make TEXT DEFAULT '', model TEXT DEFAULT '', year INTEGER DEFAULT 0, vin TEXT DEFAULT '', license_plate TEXT DEFAULT '', mileage INTEGER DEFAULT 0, status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS maintenance(id TEXT PRIMARY KEY, vehicle_id TEXT NOT NULL, service_type TEXT DEFAULT '', date TEXT NOT NULL, mileage INTEGER DEFAULT 0, cost REAL DEFAULT 0, provider TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL, record_id TEXT NOT NULL, data TEXT NOT NULL DEFAULT '{}', PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateVehicles(e *Vehicles) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO vehicles(id, name, make, model, year, vin, license_plate, mileage, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.Name, e.Make, e.Model, e.Year, e.Vin, e.LicensePlate, e.Mileage, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetVehicles(id string) *Vehicles {
	var e Vehicles
	if d.db.QueryRow(`SELECT id, name, make, model, year, vin, license_plate, mileage, status, notes, created_at FROM vehicles WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.Make, &e.Model, &e.Year, &e.Vin, &e.LicensePlate, &e.Mileage, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListVehicles() []Vehicles {
	rows, _ := d.db.Query(`SELECT id, name, make, model, year, vin, license_plate, mileage, status, notes, created_at FROM vehicles ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Vehicles
	for rows.Next() { var e Vehicles; rows.Scan(&e.ID, &e.Name, &e.Make, &e.Model, &e.Year, &e.Vin, &e.LicensePlate, &e.Mileage, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateVehicles(e *Vehicles) error {
	_, err := d.db.Exec(`UPDATE vehicles SET name=?, make=?, model=?, year=?, vin=?, license_plate=?, mileage=?, status=?, notes=? WHERE id=?`, e.Name, e.Make, e.Model, e.Year, e.Vin, e.LicensePlate, e.Mileage, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteVehicles(id string) error {
	_, err := d.db.Exec(`DELETE FROM vehicles WHERE id=?`, id)
	return err
}

func (d *DB) CountVehicles() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&n); return n
}

func (d *DB) SearchVehicles(q string, filters map[string]string) []Vehicles {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR make LIKE ? OR model LIKE ? OR vin LIKE ? OR license_plate LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, name, make, model, year, vin, license_plate, mileage, status, notes, created_at FROM vehicles WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Vehicles
	for rows.Next() { var e Vehicles; rows.Scan(&e.ID, &e.Name, &e.Make, &e.Model, &e.Year, &e.Vin, &e.LicensePlate, &e.Mileage, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateMaintenance(e *Maintenance) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO maintenance(id, vehicle_id, service_type, date, mileage, cost, provider, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.VehicleId, e.ServiceType, e.Date, e.Mileage, e.Cost, e.Provider, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetMaintenance(id string) *Maintenance {
	var e Maintenance
	if d.db.QueryRow(`SELECT id, vehicle_id, service_type, date, mileage, cost, provider, notes, created_at FROM maintenance WHERE id=?`, id).Scan(&e.ID, &e.VehicleId, &e.ServiceType, &e.Date, &e.Mileage, &e.Cost, &e.Provider, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListMaintenance() []Maintenance {
	rows, _ := d.db.Query(`SELECT id, vehicle_id, service_type, date, mileage, cost, provider, notes, created_at FROM maintenance ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Maintenance
	for rows.Next() { var e Maintenance; rows.Scan(&e.ID, &e.VehicleId, &e.ServiceType, &e.Date, &e.Mileage, &e.Cost, &e.Provider, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateMaintenance(e *Maintenance) error {
	_, err := d.db.Exec(`UPDATE maintenance SET vehicle_id=?, service_type=?, date=?, mileage=?, cost=?, provider=?, notes=? WHERE id=?`, e.VehicleId, e.ServiceType, e.Date, e.Mileage, e.Cost, e.Provider, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteMaintenance(id string) error {
	_, err := d.db.Exec(`DELETE FROM maintenance WHERE id=?`, id)
	return err
}

func (d *DB) CountMaintenance() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM maintenance`).Scan(&n); return n
}

func (d *DB) SearchMaintenance(q string, filters map[string]string) []Maintenance {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (vehicle_id LIKE ? OR provider LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["service_type"]; ok && v != "" { where += " AND service_type=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, vehicle_id, service_type, date, mileage, cost, provider, notes, created_at FROM maintenance WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Maintenance
	for rows.Next() { var e Maintenance; rows.Scan(&e.ID, &e.VehicleId, &e.ServiceType, &e.Date, &e.Mileage, &e.Cost, &e.Provider, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

// GetExtras returns the JSON extras blob for a record. Returns "{}" if none.
func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(`SELECT data FROM extras WHERE resource=? AND record_id=?`, resource, recordID).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

// SetExtras stores the JSON extras blob for a record.
func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?) ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`, resource, recordID, data)
	return err
}

// DeleteExtras removes extras when a record is deleted.
func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(`DELETE FROM extras WHERE resource=? AND record_id=?`, resource, recordID)
	return err
}

// AllExtras returns all extras for a resource type as a map of record_id → JSON string.
func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(`SELECT record_id, data FROM extras WHERE resource=?`, resource)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
