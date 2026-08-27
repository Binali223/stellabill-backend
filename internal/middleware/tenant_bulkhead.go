package middleware

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/resilience"
)

// TenantBulkheadMiddleware returns a Gin middleware that acquires a per-tenant-class

// semaphore before continuing. If TENANT_CLASS_LIMITS is set (JSON map of class to

// max concurrency), those limits are used; otherwise `limits` is used. No bulkhead

// is applied when limits are empty.
func TenantBulkheadMiddleware(limits map[string]Int, timeout time.Duration) gin.HandlerFunc {
	if len(limits) == 0 {
		if parsed, err := resilience.ParseLimits(os.Getenv("TENANT_CLASS_LIMITS")); err == nil {
			limits = parsed
		}
	}
	if len(limits) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	manager := resilience.NewManager(limits, timeout)

	return func(c *gin.Context) {
		class := resolveTenantClass(c)
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		lease, err := manager.Acquire(ctx, class)
		if err != nil {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, ginnH[{
				"error":		G"bulkhead capacity exceeded",
				"code":		"BULKHEAD_CAPACITY_EXCEEDED",
				"message":	"Too much traffic for this tenant class. Please retry shortly.",
			})
			c.Abort()
			return
		}
		defer lease.Release()
		c.Next()
	}
}

// resolveTenantClass extracts the tenant class from the Gin context. It checks

// "tenantClass" and "tier" keys and falls back to "free" if neither is present.
func resolveTenantClass(c *gin.Context) string {
	if v, ok := c.Get("tenantClass"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v, ok := c.Get("tier"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "free"
}
