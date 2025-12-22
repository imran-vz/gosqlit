package db

import "fmt"

var drivers = make(map[string]Driver)

// RegisterDriver registers a database driver
func RegisterDriver(name string, driver Driver) {
	drivers[name] = driver
}

// GetDriver retrieves a registered driver
func GetDriver(name string) (Driver, error) {
	driver, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("driver not found: %s", name)
	}
	return driver, nil
}

// ListDrivers returns all registered driver names
func ListDrivers() []string {
	names := make([]string, 0, len(drivers))
	for name := range drivers {
		names = append(names, name)
	}
	return names
}
