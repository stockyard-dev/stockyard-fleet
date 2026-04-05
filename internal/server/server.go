package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/stockyard-dev/stockyard-fleet/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}
	s.mux.HandleFunc("GET /api/vehicles", s.listVehicles)
	s.mux.HandleFunc("POST /api/vehicles", s.createVehicles)
	s.mux.HandleFunc("GET /api/vehicles/export.csv", s.exportVehicles)
	s.mux.HandleFunc("GET /api/vehicles/{id}", s.getVehicles)
	s.mux.HandleFunc("PUT /api/vehicles/{id}", s.updateVehicles)
	s.mux.HandleFunc("DELETE /api/vehicles/{id}", s.delVehicles)
	s.mux.HandleFunc("GET /api/maintenance", s.listMaintenance)
	s.mux.HandleFunc("POST /api/maintenance", s.createMaintenance)
	s.mux.HandleFunc("GET /api/maintenance/export.csv", s.exportMaintenance)
	s.mux.HandleFunc("GET /api/maintenance/{id}", s.getMaintenance)
	s.mux.HandleFunc("PUT /api/maintenance/{id}", s.updateMaintenance)
	s.mux.HandleFunc("DELETE /api/maintenance/{id}", s.delMaintenance)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/fleet/"})})
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listVehicles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"vehicles": oe(s.db.SearchVehicles(q, filters))}); return }
	wj(w, 200, map[string]any{"vehicles": oe(s.db.ListVehicles())})
}

func (s *Server) createVehicles(w http.ResponseWriter, r *http.Request) {
	if s.limits.MaxItems > 0 { if s.db.CountVehicles() >= s.limits.MaxItems { we(w, 402, "Free tier limit reached. Upgrade at https://stockyard.dev/fleet/"); return } }
	var e store.Vehicles
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateVehicles(&e)
	wj(w, 201, s.db.GetVehicles(e.ID))
}

func (s *Server) getVehicles(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetVehicles(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateVehicles(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetVehicles(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Vehicles
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	if patch.Make == "" { patch.Make = existing.Make }
	if patch.Model == "" { patch.Model = existing.Model }
	if patch.Vin == "" { patch.Vin = existing.Vin }
	if patch.LicensePlate == "" { patch.LicensePlate = existing.LicensePlate }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateVehicles(&patch)
	wj(w, 200, s.db.GetVehicles(patch.ID))
}

func (s *Server) delVehicles(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteVehicles(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportVehicles(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListVehicles()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=vehicles.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "name", "make", "model", "year", "vin", "license_plate", "mileage", "status", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Name), fmt.Sprintf("%v", e.Make), fmt.Sprintf("%v", e.Model), fmt.Sprintf("%v", e.Year), fmt.Sprintf("%v", e.Vin), fmt.Sprintf("%v", e.LicensePlate), fmt.Sprintf("%v", e.Mileage), fmt.Sprintf("%v", e.Status), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listMaintenance(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("service_type"); v != "" { filters["service_type"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"maintenance": oe(s.db.SearchMaintenance(q, filters))}); return }
	wj(w, 200, map[string]any{"maintenance": oe(s.db.ListMaintenance())})
}

func (s *Server) createMaintenance(w http.ResponseWriter, r *http.Request) {
	var e store.Maintenance
	json.NewDecoder(r.Body).Decode(&e)
	if e.VehicleId == "" { we(w, 400, "vehicle_id required"); return }
	if e.Date == "" { we(w, 400, "date required"); return }
	s.db.CreateMaintenance(&e)
	wj(w, 201, s.db.GetMaintenance(e.ID))
}

func (s *Server) getMaintenance(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetMaintenance(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateMaintenance(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetMaintenance(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Maintenance
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.VehicleId == "" { patch.VehicleId = existing.VehicleId }
	if patch.ServiceType == "" { patch.ServiceType = existing.ServiceType }
	if patch.Date == "" { patch.Date = existing.Date }
	if patch.Provider == "" { patch.Provider = existing.Provider }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateMaintenance(&patch)
	wj(w, 200, s.db.GetMaintenance(patch.ID))
}

func (s *Server) delMaintenance(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteMaintenance(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportMaintenance(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListMaintenance()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=maintenance.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "vehicle_id", "service_type", "date", "mileage", "cost", "provider", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.VehicleId), fmt.Sprintf("%v", e.ServiceType), fmt.Sprintf("%v", e.Date), fmt.Sprintf("%v", e.Mileage), fmt.Sprintf("%v", e.Cost), fmt.Sprintf("%v", e.Provider), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["vehicles_total"] = s.db.CountVehicles()
	m["maintenance_total"] = s.db.CountMaintenance()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "fleet"}
	m["vehicles"] = s.db.CountVehicles()
	m["maintenance"] = s.db.CountMaintenance()
	wj(w, 200, m)
}
