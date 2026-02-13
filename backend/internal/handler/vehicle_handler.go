package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

// allowedMIMESignatures maps file extensions to their magic byte signatures.
var allowedMIMESignatures = map[string][][]byte{
	".jpg":  {{0xFF, 0xD8, 0xFF}},
	".jpeg": {{0xFF, 0xD8, 0xFF}},
	".png":  {{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}},
	".webp": {{'R', 'I', 'F', 'F'}}, // RIFF header; full check includes "WEBP" at offset 8
}

type VehicleHandler struct {
	vehicleSvc  vehicleService
	locationSvc locationService
	uploadDir   string
}

func NewVehicleHandler(vehicleSvc vehicleService, locationSvc locationService, uploadDir string) *VehicleHandler {
	return &VehicleHandler{vehicleSvc: vehicleSvc, locationSvc: locationSvc, uploadDir: uploadDir}
}

func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.vehicleSvc.ListWithStatus(r.Context())
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	apperror.WriteSuccess(w, vehicles)
}

func (h *VehicleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	vehicle, err := h.vehicleSvc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if vehicle == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}
	apperror.WriteSuccess(w, vehicle)
}

func (h *VehicleHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.vehicleSvc.ListAvailable(r.Context())
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	apperror.WriteSuccess(w, vehicles)
}

func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.Name == "" || req.LicensePlate == "" || req.DriverID == "" {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	vehicle, err := h.vehicleSvc.Create(r.Context(), claims.UserID, req.Name, req.LicensePlate, req.DriverID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	apperror.WriteSuccess(w, vehicle)
}

func (h *VehicleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.UpdateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.Name == "" || req.LicensePlate == "" || req.DriverID == "" {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.vehicleSvc.Update(r.Context(), claims.UserID, id, req.Name, req.LicensePlate, req.DriverID); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.vehicleSvc.Delete(r.Context(), claims.UserID, id); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Check vehicle exists
	vehicle, err := h.vehicleSvc.GetByID(r.Context(), id)
	if err != nil || vehicle == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	// 10 MB max
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	// Validate magic bytes match declared extension
	if !validateMagicBytes(file, ext) {
		apperror.WriteErrorMsg(w, 400, "INVALID_FILE", "file content does not match declared type")
		return
	}
	// Seek back to start after reading magic bytes
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Ensure upload directory exists
	vehicleDir := filepath.Join(h.uploadDir, "vehicles")
	if err := os.MkdirAll(vehicleDir, 0755); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	// Delete old photo if exists (safe: uses only basename)
	if vehicle.PhotoURL != nil {
		h.safeDeleteOldPhoto(*vehicle.PhotoURL)
	}

	// Save new file
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(vehicleDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	// Update DB
	photoURL := fmt.Sprintf("/uploads/vehicles/%s", filename)
	if err := h.vehicleSvc.UpdatePhotoURL(r.Context(), id, &photoURL); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, map[string]string{"photo_url": photoURL})
}

// safeDeleteOldPhoto removes an old photo file using only the base filename,
// preventing path traversal attacks.
func (h *VehicleHandler) safeDeleteOldPhoto(photoURL string) {
	// Extract just the filename from the URL path
	base := filepath.Base(photoURL)
	if base == "." || base == "/" || base == ".." {
		return
	}

	fullPath := filepath.Join(h.uploadDir, "vehicles", base)

	// Verify the resolved path is within the upload directory
	absUpload, err := filepath.Abs(filepath.Join(h.uploadDir, "vehicles"))
	if err != nil {
		return
	}
	absTarget, err := filepath.Abs(fullPath)
	if err != nil {
		return
	}
	if !strings.HasPrefix(absTarget, absUpload+string(filepath.Separator)) && absTarget != absUpload {
		return
	}

	os.Remove(fullPath)
}

// validateMagicBytes reads the first bytes of a file and checks them against
// the expected magic bytes for the given extension.
func validateMagicBytes(r io.Reader, ext string) bool {
	sigs, ok := allowedMIMESignatures[ext]
	if !ok {
		return false
	}

	// Read enough bytes to check the longest signature
	maxLen := 0
	for _, sig := range sigs {
		if len(sig) > maxLen {
			maxLen = len(sig)
		}
	}

	buf := make([]byte, maxLen)
	n, err := io.ReadFull(r, buf)
	if err != nil && n < maxLen {
		return false
	}

	for _, sig := range sigs {
		if n >= len(sig) {
			match := true
			for i, b := range sig {
				if buf[i] != b {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

func (h *VehicleHandler) ToggleMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.ToggleMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.vehicleSvc.ToggleMaintenance(r.Context(), claims.UserID, id, req.IsMaintenance); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) LocationHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)

	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}

	locations, err := h.locationSvc.GetHistory(r.Context(), id, from, to)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, locations)
}
