package middlewares

import (
	"fmt"
	"l4d2-manager-next/utility"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	allowedRegions []string
)

// InitGeoIPMiddleware parses the REGION_WHITE_LIST environment variable
// Note: GeoIP service initialization is now handled in utility.InitGeoIP
func InitGeoIPMiddleware() {
	// Parse Whitelist
	whitelistEnv := os.Getenv("REGION_WHITE_LIST")
	allowedRegions = []string{} // Reset
	if whitelistEnv != "" {
		parts := strings.Split(whitelistEnv, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				allowedRegions = append(allowedRegions, trimmed)
			}
		}
	}

	// If whitelist is active, ensure Reserved/Internal IPs are allowed
	if len(allowedRegions) > 0 {
		allowedRegions = append(allowedRegions, "Reserved")
		fmt.Printf("GeoIP middleware initialized. Whitelist: %v\n", allowedRegions)
	} else {
		fmt.Println("GeoIP middleware initialized. No whitelist (allowing all).")
	}
}

// BlockForeignIPs returns a middleware that blocks IPs not in REGION_WHITE_LIST
func BlockForeignIPs() gin.HandlerFunc {
	// If no whitelist is defined (and InitGeoIP has run), allow everything
	if len(allowedRegions) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		ipRegionService := utility.GetIPRegionService()
		if ipRegionService == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()

		// Skip check for local IPs (Always allow localhost/internal loopback)
		if ip == "::1" || ip == "127.0.0.1" || ip == "localhost" {
			c.Next()
			return
		}

		// Query Region
		region, err := ipRegionService.SearchByStr(ip)
		if err != nil {
			fmt.Printf("GeoIP search failed for IP %s: %v\n", ip, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "GeoIP lookup failed",
			})
			return
		}

		// Check against whitelist
		allowed := false
		for _, allowStr := range allowedRegions {
			if strings.Contains(region, allowStr) {
				allowed = true
				break
			}
		}

		if allowed {
			c.Next()
			return
		}

		// Log the blocked attempt
		fmt.Printf("Blocked IP: %s, Region: %s\n", ip, region)
		c.AbortWithStatus(http.StatusForbidden)
	}
}
