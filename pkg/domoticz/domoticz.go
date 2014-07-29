package domoticz

import (
	"fmt"
	"io/ioutil"
	"net/http"

	log "code.google.com/p/log4go"
)

// Domoticz server
type Domoticz struct {
	Host       string
	Port       int
	HardwareId string
}

// Push data to domoticz server
//
// Device can be created dynamically by using those parameters instead of 'idx':
//  hid: HardwareID
//  did: DeviceID
//  dunit: Unit
//  dtype: Type
//  dsubtype: SubType
//
// Example Temp+Humi:
//  hid: 1      => jeego virtual device already created with domoticz UI)
//  did: 4000   => "ID": "0FA0" (created dynamically)
//  dunit: 4    => ??
//  dtype: 82   => pTypeTEMP_HUM 0x52 (temperature+humidity)
//  dsubtype: 1 => sTypeTH1 0x1  //THGN122/123,THGN132,THGR122/228/238/268
//
//  http://pikan.local:8080/json.htm?type=command&param=udevice&hid=1&did=4000&dunit=4&dtype=82&dsubtype=1&nvalue=0&svalue=12.3;99;0
//
// Example Temp:
//  ...
//  dtype: 80   => pTypeTEMP 0x50 (temperature)
//  dsubtype: 1 => sTypeTEMP1 0x1  //THR128/138,THC138
func (server *Domoticz) Push(params string) {
	if server.Host != "" {
		if params != "" {
			url := fmt.Sprintf("http://%s:%d/json.htm?type=command&param=udevice&%s", server.Host, server.Port, params)

			log.Debug(fmt.Sprintf("Pushing to domoticz: %s", url))

			resp, err := http.Get(url)
			if err != nil {
				log.Warn("Failed to push value to domoticz")
			} else {
				respText, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					log.Warn("Failed to get domoticz response")
				} else {
					log.Debug(fmt.Sprintf("Domoticz response: %s", respText))
				}
			}
		}
	}
}
