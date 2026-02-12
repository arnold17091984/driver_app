package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/kento/driver/backend/internal/handler"
	"github.com/kento/driver/backend/internal/middleware"
)

func buildRouter(
	jwtSecret string,
	authH *handler.AuthHandler,
	vehicleH *handler.VehicleHandler,
	dispatchH *handler.DispatchHandler,
	reservationH *handler.ReservationHandler,
	conflictH *handler.ConflictHandler,
	attendanceH *handler.AttendanceHandler,
	locationH *handler.LocationHandler,
	adminH *handler.AdminHandler,
	notifH *handler.NotificationHandler,
	routeH *handler.RouteHandler,
	bookingH *handler.BookingHandler,
) chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(middleware.NewRateLimiter(20, 40).Limit) // 20 req/s per IP, burst 40

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static file serving for uploads
	fileServer := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", fileServer))

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth endpoints
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/refresh", authH.Refresh)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtSecret))

			// Auth
			r.Get("/auth/me", authH.Me)
			r.Post("/auth/logout", authH.Logout)

			// Notifications
			r.Put("/notifications/fcm-token", notifH.UpdateFCMToken)

			// Vehicles (P1 - all authenticated)
			r.Get("/vehicles", vehicleH.List)
			r.Get("/vehicles/{id}", vehicleH.Get)
			r.Get("/vehicles/{id}/location/history", vehicleH.LocationHistory)
			r.Get("/vehicles/{id}/timeline", bookingH.GetVehicleTimeline)
			r.Get("/vehicles/available", vehicleH.ListAvailable)

			// Routes (Google Routes API proxy)
			r.Post("/routes/compute", routeH.ComputeRoute)

			// Dispatches (read: all; write: dispatcher+)
			r.Get("/dispatches", dispatchH.List)
			r.Get("/dispatches/{id}", dispatchH.Get)
			r.Get("/dispatches/{id}/eta", dispatchH.GetETASnapshots)

			// Reservations (read: all)
			r.Get("/reservations", reservationH.List)
			r.Get("/reservations/{id}", reservationH.Get)
			r.Get("/reservations/availability", reservationH.CheckAvailability)

			// Attendance (read: all)
			r.Get("/attendance/history", attendanceH.GetHistory)

			// Dispatcher+ routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin", "dispatcher"))

				// Dispatch management
				r.Post("/dispatches", dispatchH.Create)
				r.Post("/dispatches/quick-board", dispatchH.QuickBoard)
				r.Post("/dispatches/calculate-eta", dispatchH.CalculateETAs)
				r.Post("/dispatches/{id}/assign", dispatchH.Assign)
				r.Post("/dispatches/{id}/alight", dispatchH.Alight)
				r.Post("/dispatches/{id}/cancel", dispatchH.Cancel)

				// Reservation management
				r.Post("/reservations", reservationH.Create)
				r.Put("/reservations/{id}", reservationH.Update)
				r.Post("/reservations/{id}/cancel", reservationH.Cancel)

				// Unified booking
				r.Post("/bookings", bookingH.CreateBooking)

				// Conflict management (P7)
				r.Get("/conflicts", conflictH.ListPending)
				r.Get("/conflicts/{id}", conflictH.Get)
				r.Post("/conflicts/{id}/reassign", conflictH.Reassign)
				r.Post("/conflicts/{id}/change-time", conflictH.ChangeTime)
				r.Post("/conflicts/{id}/cancel", conflictH.Cancel)

				// Vehicle maintenance toggle (P10)
				r.Patch("/vehicles/{id}/maintenance", vehicleH.ToggleMaintenance)
			})

			// Admin only routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))

				r.Get("/admin/users", adminH.ListUsers)
				r.Put("/admin/users/{id}/role", adminH.UpdateRole)
				r.Put("/admin/users/{id}/priority", adminH.UpdatePriority)
				r.Get("/admin/audit-logs", adminH.ListAuditLogs)
				r.Get("/admin/audit-logs/{id}", adminH.GetAuditLog)

				// Vehicle CRUD (admin only)
				r.Post("/vehicles", vehicleH.Create)
				r.Put("/vehicles/{id}", vehicleH.Update)
				r.Delete("/vehicles/{id}", vehicleH.Delete)
				r.Post("/vehicles/{id}/photo", vehicleH.UploadPhoto)

				// Force assign (P8 - admin only)
				r.Post("/conflicts/{id}/force-assign", conflictH.ForceAssign)
			})

			// Driver only routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("driver"))

				r.Post("/locations/report", locationH.Report)
				r.Post("/attendance/clock-in", attendanceH.ClockIn)
				r.Post("/attendance/clock-out", attendanceH.ClockOut)
				r.Get("/attendance/status", attendanceH.GetStatus)
				r.Put("/driver/status", attendanceH.UpdateDriverStatus)
				r.Get("/driver/trips/current", dispatchH.CurrentTrip)
				r.Post("/driver/trips/{id}/accept", dispatchH.AcceptTrip)
				r.Post("/driver/trips/{id}/en-route", dispatchH.EnRouteTrip)
				r.Post("/driver/trips/{id}/arrived", dispatchH.ArriveTrip)
				r.Post("/driver/board", dispatchH.DriverBoard)
				r.Post("/driver/trips/{id}/alight", dispatchH.Alight)
				r.Post("/driver/trips/{id}/complete", dispatchH.CompleteTrip)

				// Driver reservation management
				r.Get("/driver/reservations/pending", bookingH.PendingReservations)
				r.Post("/driver/reservations/{id}/accept", bookingH.AcceptReservation)
				r.Post("/driver/reservations/{id}/decline", bookingH.DeclineReservation)
			})
		})
	})

	return r
}
