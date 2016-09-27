/*
    Host Resources - Network

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"net"
)

// GetIPOfInt will iterate over all addresses for the given network interface, but will return only
// the first one it finds. TODO(mierdin): This has obvious drawbacks, particularly with IPv6. Need to figure out a better way.
func GetIPOfInt(ifname string) net.IP {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range interfaces {
		if iface.Name == ifname {

			addrs, err := iface.Addrs()
			if err != nil {
				panic(err)
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						return ipnet.IP
					}

				}
			}
		}
	}
	return nil
}
