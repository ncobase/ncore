package helper

import (
	"fmt"
	"ncore/pkg/config"
	"strconv"
)

// GetHost constructs the URL based on the given tenant and config, with an optional port.
func GetHost(conf *config.Config, tenant string, ports ...int) string {
	port := getPort(conf, ports...)
	return buildURL(conf.Protocol, tenant, port)
}

// getPort retrieves the port number from the config or the optional ports parameter.
func getPort(conf *config.Config, ports ...int) string {
	if len(ports) > 0 {
		return strconv.Itoa(ports[0])
	} else if conf.Port != 0 {
		return strconv.Itoa(conf.Port)
	}
	return ""
}

// buildURL constructs the URL string based on the protocol, tenant, and optional port.
func buildURL(protocol, tenant, port string) string {
	if port != "" {
		return fmt.Sprintf("%v://%v:%v", protocol, tenant, port)
	}
	return fmt.Sprintf("%v://%v", protocol, tenant)
}
