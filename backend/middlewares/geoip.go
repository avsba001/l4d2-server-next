package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/service"
)

var (
	ipRegionService *service.Ip2Region
	allowedRegions  []string
)

// InitGeoIP initializes the ip2region service with v4 and v6 databases
// and parses the REGION_WHITE_LIST environment variable
func InitGeoIP(v4Path, v6Path string) bool {
	var err error

	// 1. Create v4 config
	v4Config, err := service.NewV4Config(service.VIndexCache, v4Path, 20)
	if err != nil {
		fmt.Printf("Failed to create v4 config: %v\n", err)
		return false
	}

	// 2. Create v6 config
	v6Config, err := service.NewV6Config(service.VIndexCache, v6Path, 20)
	if err != nil {
		fmt.Printf("Failed to create v6 config: %v\n", err)
		return false
	}

	// 3. Create service
	ipRegionService, err = service.NewIp2Region(v4Config, v6Config)
	if err != nil {
		fmt.Printf("Failed to create ip2region service: %v\n", err)
		return false
	}

	// 4. Parse Whitelist
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

	return true
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
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":  "Access denied: Region not allowed",
			"region": region,
		})
	}
}

// CloseGeoIP closes the service
func CloseGeoIP() {
	if ipRegionService != nil {
		ipRegionService.Close()
	}
}
