package utility

import (
	"fmt"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
)

var (
	ipRegionService *service.Ip2Region
)

// InitGeoIP initializes the ip2region service with v4 and v6 databases
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

	fmt.Println("GeoIP service initialized.")
	return true
}

// GetLocation returns the location string for a given IP
// Format: [ ISP ] Country-Province-City
func GetLocation(ip string) string {
	if ipRegionService == nil {
		return ""
	}

	// Clean up IP (remove port if present)
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	// Skip check for local IPs
	if ip == "::1" || ip == "127.0.0.1" || ip == "localhost" {
		return "Localhost"
	}

	region, err := ipRegionService.SearchByStr(ip)
	if err != nil {
		return ""
	}

	// Format: Country|Region|Province|City|ISP
	// Example: 中国|0|上海|上海市|电信
	// Raw Log Example: 中国|黑龙江省|双鸭山市|电信|CN
	//
	// The parts seem to be:
	// parts[0]: Country (中国)
	// parts[1]: Province (黑龙江省) - previously thought to be Region/0
	// parts[2]: City (双鸭山市) - previously thought to be Province
	// parts[3]: ISP (电信) - previously thought to be City
	// parts[4]: Extra info (CN) - previously thought to be ISP

	parts := strings.Split(region, "|")
	if len(parts) >= 5 {
		var validParts []string

		// Country (parts[0])
		if parts[0] != "0" {
			validParts = append(validParts, parts[0])
		}

		// Province (parts[1])
		if parts[1] != "0" && parts[1] != parts[0] {
			validParts = append(validParts, parts[1])
		}

		// City (parts[2])
		if parts[2] != "0" && parts[2] != parts[1] {
			validParts = append(validParts, parts[2])
		}

		locationStr := strings.Join(validParts, "-")

		// ISP (parts[3])
		if parts[3] != "0" {
			return fmt.Sprintf("[ %s ] %s", parts[3], locationStr)
		}

		return locationStr
	}

	return region
}

// GetIPRegionService returns the initialized service instance
func GetIPRegionService() *service.Ip2Region {
	return ipRegionService
}

// CloseGeoIP closes the service
func CloseGeoIP() {
	if ipRegionService != nil {
		ipRegionService.Close()
	}
}
