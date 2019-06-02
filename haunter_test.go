// Copyright 2019 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package haunter

import "testing"

func TestRandProxy(t *testing.T) {
	data := Proxies{
		Proxy{
			// Container{
			// 	Proxy: Proxy{
			// 		IP:          "127.0.0.1",
			// 		PortNum:     "80",
			// 		CountryCode: "UK",
			// 		CountryName: "United Kingdom",
			// 		RegionName:  "England",
			// 		CityName:    "London",
			// 		Status:      "online",
			// 		PanelUser:   "",
			// 		PanelPass:   "",
			// 	},
			// },
			ProxyIP:       "127.0.0.2",
			ProxyPort:     "81",
			ProxyCountry:  "US",
			ProxyArea:     "New York",
			ProxyLocation: "New York",
			ProxyStatus:   "offline",
			Username:      "",
			Password:      "",
		},
	}

	selection := data.RandProxy().ProxyCountry
	// FIXME: vet complains about this comparison...
	// if selection != "US" || selection != "UK" {
	if selection != "US" {
		t.Errorf("RandProxy(%q).CountryCode == %q, want \"US\" or \"UK\"", data, selection)
	}
}
