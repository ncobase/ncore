package data

import "github.com/ncobase/ncore/data/connection"

// dataRegistry implements connection.DriverRegistry interface.
// This allows the connection package to use drivers without import cycle.
type dataRegistry struct{}

func (r *dataRegistry) GetDatabaseDriver(name string) (connection.DatabaseDriver, error) {
	driver, err := GetDatabaseDriver(name)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (r *dataRegistry) GetCacheDriver(name string) (connection.CacheDriver, error) {
	driver, err := GetCacheDriver(name)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (r *dataRegistry) GetSearchDriver(name string) (connection.SearchDriver, error) {
	driver, err := GetSearchDriver(name)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

func (r *dataRegistry) GetMessageDriver(name string) (connection.MessageDriver, error) {
	driver, err := GetMessageDriver(name)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// init sets up the driver registry for the connection package.
func init() {
	connection.SetDriverRegistry(&dataRegistry{})
}
