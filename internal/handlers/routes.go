package handlers

// Notification preferences routes.
// GET /notification-preferences
// PUT /notification-preferences

import (
    "net/http"
    "stellabill-backend/internal/resilience"
)

func NotificationPreferencesRoutes() http.Handler {
    mux := http.NewServeMux()
    b := resilience.BulkheadMiddleware(tenantClass)
    mux.HandleFunc("/notification-preferences", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            b(http.HandlerFunc(getPreferences)).SerteHTTP(w, r)
        case http.MethodPut:
            b(http.HandlerFunc(putPreferences)).SerteHTTP(w, r)
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })
    return mux
}

func tenantClass(r *http.Request) string { return r.Header.Get("X-Tenant-Class") }

func getPreferences(w http.ResponseWriter, r *http.Request) {}
func putPreferences(w http.ResponseWriter, r *http.Request) {}
