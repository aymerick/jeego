package main

import (
    "net/http"
    "io/ioutil"
    "log"
    "fmt"
)

func pushToDomoticz(config *Config, node INode) {
    if config.DomoticzHost != "" {
        idx := node.domoticzIdx()
        value := node.domoticzValue()

        if (idx != "") && (value != "") {
            // Example:
            //   $ curl -s -i -H "Accept: application/json" "http://pikan.local:8080/json.htm?type=command&param=udevice&idx=2&nvalue=0&svalue=12.3;45;0"
            url := fmt.Sprintf("http://%s:%d/json.htm?type=command&param=udevice&idx=%s&nvalue=0&svalue=%s", config.DomoticzHost, config.DomoticzPort, idx, value)

            // url := fmt.Sprintf("http://%s:%d/json.htm?type=command&param=udevice&nvalue=0&svalue=%s&...", config.DomoticzHost, config.DomoticzPort, value, ...)
            //   hid: HardwareID
            //   did: DeviceID
            //   dunit: Unit
            //   dtype: Type
            //   dsubtype: SubType

            // cf. http://code.google.com/p/usb-sensors-linux/wiki/Domoticz
            //
            // # Sensor parameters
            // HID="1"
            // DID="4000"
            // DUNIT="4"
            // DTYPE="82"
            // DSUBTYPE="1"
            // NVALUE="0"
            // SVALUE="$TEMP;$HUM;9"
            //
            // # Send data
            // curl -s -i -H "Accept: application/json" "http://$SERVER/json.htm?type=command&param=udevice&hid=$HID&did=$DID&dunit=$DUNIT&dtype=$DTYPE&dsubtype=$DSUBTYPE&nvalue=$NVALUE&svalue=$SVALUE"
            //
            // With these sensor type the data will be reported as type "Temp + Humidity" and subtype "Oregon THGN122/123, THGN132, THGR122/228/238/268".

            if config.Debug {
                log.Printf("[%s] Pushing to domoticz: %s", node.name(), url)
            }

            resp, err := http.Get(url)
            if err != nil {
                log.Printf("[%s] Failed to push value to domoticz", node.name())
            } else {
                respText, err := ioutil.ReadAll(resp.Body)
                resp.Body.Close()
                if err != nil {
                    log.Printf("[%s] Failed to get domoticz response", node.name())
                } else {
                    if config.Debug {
                        log.Printf("[%s] Domoticz response: %s", node.name(), respText)
                    }
                }
            }
        }
    }
}
