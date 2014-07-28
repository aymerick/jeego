package app

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// JEENODE_THLM_NODE
//

func Test_JeenodeTHLM_Sensors(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	expected := []Sensor{TEMP_SENSOR, HUMI_SENSOR, LIGHT_SENSOR, MOTION_SENSOR, LOWBAT_SENSOR}

	assert.True(t, reflect.DeepEqual(node.sensors(), expected))
}

func Test_JeenodeTHLM_AbsentSensors(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	expected := []Sensor{VCC_SENSOR}

	assert.True(t, reflect.DeepEqual(node.absentSensors(), expected))
}

func Test_JeenodeTHLM_HaveSensor(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	if !node.haveSensor(TEMP_SENSOR) ||
		!node.haveSensor(HUMI_SENSOR) ||
		!node.haveSensor(LIGHT_SENSOR) ||
		!node.haveSensor(MOTION_SENSOR) ||
		!node.haveSensor(LOWBAT_SENSOR) ||
		node.haveSensor(VCC_SENSOR) {
		t.Errorf("JEENODE_THLM_NODE - Incorrect sensor attribution")
	}
}

func Test_JeenodeTHLM_ExpectedDataLength(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	assert.Equal(t, node.expectedDataLength(), 4, "Incorrect expected data length")
}

//
//             213 => 1 1 0 1 0 1 0 1
//              40 => 0 0 1 0 1 0 0 0
//              57 => 0 0 1 1 1 0 0 1
//               3 => 0 0 0 0 0 0 1 1
//
//       temperature: 1 1 0 1 0 1 0 1
//                                0 0 => 213 / 10 = 21.3
//          humidity: 0 0 1 0 1 0
//                                  1 => 74
//             light: 0 0 1 1 1 0 0
//                                  1 => 156 * 100 / 255 = 61
//            motion:             1   => true
//       low battery:           0     => false
//        <not used>: 0 0 0 0 0
//
func Test_JeenodeTHLM_ParseData(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	data := []byte{213, 40, 57, 3}
	expected := map[Sensor]uint64{TEMP_SENSOR: 213, HUMI_SENSOR: 74, LIGHT_SENSOR: 156, MOTION_SENSOR: 1, LOWBAT_SENSOR: 0}

	result := node.parseData(data)
	assert.True(t, reflect.DeepEqual(result, expected))
}

//
// JEENODE_THL_NODE
//

func Test_JeenodeTHL_Sensors(t *testing.T) {
	node := &Node{Kind: JEENODE_THL_NODE}

	expected := []Sensor{TEMP_SENSOR, HUMI_SENSOR, LIGHT_SENSOR, LOWBAT_SENSOR}

	assert.True(t, reflect.DeepEqual(node.sensors(), expected))
}

func Test_JeenodeTHL_AbsentSensors(t *testing.T) {
	node := &Node{Kind: JEENODE_THL_NODE}

	expected := []Sensor{MOTION_SENSOR, VCC_SENSOR}

	assert.True(t, reflect.DeepEqual(node.absentSensors(), expected))
}

func Test_JeenodeTHL_HaveSensor(t *testing.T) {
	node := &Node{Kind: JEENODE_THL_NODE}

	if !node.haveSensor(TEMP_SENSOR) ||
		!node.haveSensor(HUMI_SENSOR) ||
		!node.haveSensor(LIGHT_SENSOR) ||
		node.haveSensor(MOTION_SENSOR) ||
		!node.haveSensor(LOWBAT_SENSOR) ||
		node.haveSensor(VCC_SENSOR) {
		t.Errorf("JEENODE_THL_NODE - Incorrect sensor attribution")
	}
}

func Test_JeenodeTHL_ExpectedDataLength(t *testing.T) {
	node := &Node{Kind: JEENODE_THL_NODE}

	assert.Equal(t, node.expectedDataLength(), 4, "Incorrect expected data length")
}

//
//             210 => 1 1 0 1 0 0 1 0
//             200 => 1 1 0 0 1 0 0 0
//             112 => 0 1 1 1 0 0 0 0
//              1  => 0 0 0 0 0 0 0 1
//
//       temperature: 1 1 0 1 0 0 1 0
//                                0 0 => 210 / 10 = 21.0
//          humidity: 1 1 0 0 1 0
//                                  0 => 50
//             light: 0 1 1 1 0 0 0
//                                  1 => 184 * 100 / 255 = 72
//       low battery:             0   => false
//        <not used>: 0 0 0 0 0 0
//
func Test_JeenodeTHL_ParseData(t *testing.T) {
	node := &Node{Kind: JEENODE_THL_NODE}

	data := []byte{210, 200, 112, 1}
	expected := map[Sensor]uint64{TEMP_SENSOR: 210, HUMI_SENSOR: 50, LIGHT_SENSOR: 184, LOWBAT_SENSOR: 0}

	result := node.parseData(data)
	assert.True(t, reflect.DeepEqual(result, expected))
}

//
// TINYTX_T_NODE
//

func Test_TinyTxT_Sensors(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	expected := []Sensor{TEMP_SENSOR, VCC_SENSOR}

	assert.True(t, reflect.DeepEqual(node.sensors(), expected))
}

func Test_TinyTxT_AbsentSensors(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	expected := []Sensor{HUMI_SENSOR, LIGHT_SENSOR, MOTION_SENSOR, LOWBAT_SENSOR}

	assert.True(t, reflect.DeepEqual(node.absentSensors(), expected))
}

func Test_TinyTxT_HaveSensor(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	if !node.haveSensor(TEMP_SENSOR) ||
		node.haveSensor(HUMI_SENSOR) ||
		node.haveSensor(LIGHT_SENSOR) ||
		node.haveSensor(MOTION_SENSOR) ||
		node.haveSensor(LOWBAT_SENSOR) ||
		!node.haveSensor(VCC_SENSOR) {
		t.Errorf("TINYTX_T_NODE - Incorrect sensor attribution")
	}
}

func Test_TinyTxT_ExpectedDataLength(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	assert.Equal(t, node.expectedDataLength(), 3, "Incorrect expected data length")
}

//
//
//              18 => 0 0 0 1 0 0 1 0
//             113 => 0 1 1 1 0 0 0 1
//              49 => 0 0 1 1 0 0 0 1
//
//       temperature: 0 0 0 1 0 0 1 0
//                                0 1 => 274 / 10 = 27.4
//              vcc:  0 1 1 1 0 0
//                        1 1 0 0 0 1 => 3164 mv
//        <not used>: 0 0
//
func Test_TinyTxT_ParseData(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	data := []byte{18, 113, 49}
	expected := map[Sensor]uint64{TEMP_SENSOR: 274, VCC_SENSOR: 3164}

	result := node.parseData(data)
	assert.True(t, reflect.DeepEqual(result, expected))
}

//
// TINYTX_TH_NODE
//

func Test_TinyTxTH_HaveSensor(t *testing.T) {
	node := &Node{Kind: TINYTX_TH_NODE}

	if !node.haveSensor(TEMP_SENSOR) ||
		!node.haveSensor(HUMI_SENSOR) ||
		node.haveSensor(LIGHT_SENSOR) ||
		node.haveSensor(MOTION_SENSOR) ||
		node.haveSensor(LOWBAT_SENSOR) ||
		!node.haveSensor(VCC_SENSOR) {
		t.Errorf("TINYTX_TH_NODE - Incorrect sensor attribution")
	}
}

func Test_TinyTxTH_ExpectedDataLength(t *testing.T) {
	node := &Node{Kind: TINYTX_TH_NODE}

	assert.Equal(t, node.expectedDataLength(), 4, "Incorrect expected data length")
}

//
//              18 => 0 0 0 1 0 0 1 0
//             201 => 1 1 0 0 1 0 0 1
//             184 => 1 0 1 1 1 0 0 0
//              24 => 0 0 0 1 1 0 0 0
//
//       temperature: 0 0 0 1 0 0 1 0
//                                0 1  => 274 / 10 = 27.4
//          humidity: 1 1 0 0 1 0
//                                  0  => 50
//               vcc: 1 0 1 1 1 0 0
//                          1 1 0 0 0 => 3164 mv
//        <not used>: 0 0 0
//
func Test_TinyTxTH_ParseData(t *testing.T) {
	node := &Node{Kind: TINYTX_TH_NODE}

	data := []byte{18, 201, 184, 24}
	expected := map[Sensor]uint64{TEMP_SENSOR: 274, HUMI_SENSOR: 50, VCC_SENSOR: 3164}

	result := node.parseData(data)
	assert.True(t, reflect.DeepEqual(result, expected))
}

//
// TINYTX_TL_NODE
//

func Test_TinyTxTL_HaveSensor(t *testing.T) {
	node := &Node{Kind: TINYTX_TL_NODE}

	if !node.haveSensor(TEMP_SENSOR) ||
		node.haveSensor(HUMI_SENSOR) ||
		!node.haveSensor(LIGHT_SENSOR) ||
		node.haveSensor(MOTION_SENSOR) ||
		node.haveSensor(LOWBAT_SENSOR) ||
		!node.haveSensor(VCC_SENSOR) {
		t.Errorf("TINYTX_TL_NODE - Incorrect sensor attribution")
	}
}

func Test_TinyTxTL_ExpectedDataLength(t *testing.T) {
	node := &Node{Kind: TINYTX_TL_NODE}

	if node.expectedDataLength() != 4 {
		t.Errorf("TINYTX_TL_NODE - Incorrect expected data length")
	}
}

//
//             245 => 1 1 1 1 0 1 0 1
//             196 => 1 1 0 0 0 1 0 0
//              79 => 0 1 0 0 1 1 1 1
//              49 => 0 0 1 1 0 0 0 1
//
//       temperature: 1 1 1 1 0 1 0 1
//                                0 0 => 245 / 10 = 24.3
//             light: 1 1 0 0 0 1
//                                1 1 => 241 * 100 / 255 = 94
//               vcc: 0 1 0 0 1 1
//                        1 1 0 0 0 1 => 3155 mv
//        <not used>: 0 0
//
func Test_TinyTxTL_ParseData(t *testing.T) {
	node := &Node{Kind: TINYTX_TL_NODE}

	data := []byte{245, 196, 79, 49}
	expected := map[Sensor]uint64{TEMP_SENSOR: 245, LIGHT_SENSOR: 241, VCC_SENSOR: 3155}

	result := node.parseData(data)
	assert.True(t, reflect.DeepEqual(result, expected))
}

//
// Sensors
//

func Test_ComputeTemperatureValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	var value uint64
	var expected float64

	value = 0
	expected = 0

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 213
	expected = 21.3

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 512
	expected = 51.2

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 1024
	expected = 0

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 614
	expected = -41.0

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 1012
	expected = -1.2

	if result := node.computeTemperatureValue(value); result != expected {
		t.Errorf("computeTemperatureValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_ComputeHumidityValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	var value uint64
	var expected uint8

	value = 0
	expected = 0

	if result := node.computeHumidityValue(value); result != expected {
		t.Errorf("computeHumidityValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 74
	expected = 74

	if result := node.computeHumidityValue(value); result != expected {
		t.Errorf("computeHumidityValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 100
	expected = 100

	if result := node.computeHumidityValue(value); result != expected {
		t.Errorf("computeHumidityValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_ComputeLightValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	var value uint64
	var expected uint8

	value = 0
	expected = 0

	if result := node.computeLightValue(value); result != expected {
		t.Errorf("computeLightValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 255
	expected = 100

	if result := node.computeLightValue(value); result != expected {
		t.Errorf("computeLightValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 128
	expected = 50

	if result := node.computeLightValue(value); result != expected {
		t.Errorf("computeLightValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 156
	expected = 61

	if result := node.computeLightValue(value); result != expected {
		t.Errorf("computeLightValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_ComputeMotionValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	var value uint64
	var expected bool

	value = 0
	expected = false

	if result := node.computeMotionValue(value); result != expected {
		t.Errorf("computeMotionValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 1
	expected = true

	if result := node.computeMotionValue(value); result != expected {
		t.Errorf("computeMotionValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_ComputeLowbatValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	var value uint64
	var expected bool

	value = 0
	expected = false

	if result := node.computeLowbatValue(value); result != expected {
		t.Errorf("computeLowbatValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 1
	expected = true

	if result := node.computeLowbatValue(value); result != expected {
		t.Errorf("computeLowbatValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_ComputeVccValue(t *testing.T) {
	node := &Node{Kind: TINYTX_T_NODE}

	var value uint64
	var expected uint

	value = 3142
	expected = 3142

	if result := node.computeVccValue(value); result != expected {
		t.Errorf("computeVccValue(%v) = %v | expected: %v", value, result, expected)
	}

	value = 3000
	expected = 3000

	if result := node.computeVccValue(value); result != expected {
		t.Errorf("computeVccValue(%v) = %v | expected: %v", value, result, expected)
	}
}

func Test_SetSensorRawValue(t *testing.T) {
	node := &Node{Kind: JEENODE_THLM_NODE}

	assert.Equal(t, node.Temperature, float64(0))
	assert.Equal(t, node.Humidity, uint8(0))
	assert.Equal(t, node.Light, uint8(0))
	assert.Equal(t, node.Motion, false)
	assert.Equal(t, node.LowBattery, false)
	assert.Equal(t, node.Vcc, uint(0))

	node.setSensorRawValue(TEMP_SENSOR, uint64(213))
	node.setSensorRawValue(HUMI_SENSOR, uint64(74))
	node.setSensorRawValue(LIGHT_SENSOR, uint64(156))
	node.setSensorRawValue(MOTION_SENSOR, uint64(1))
	node.setSensorRawValue(LOWBAT_SENSOR, uint64(1))
	node.setSensorRawValue(VCC_SENSOR, uint64(3142))

	assert.Equal(t, node.Temperature, float64(21.3))
	assert.Equal(t, node.Humidity, uint8(74))
	assert.Equal(t, node.Light, uint8(61))
	assert.Equal(t, node.Motion, true)
	assert.Equal(t, node.LowBattery, true)
	assert.Equal(t, node.Vcc, uint(3142))
}

func Test_TextData(t *testing.T) {
	node := &Node{
		Kind:        JEENODE_THLM_NODE,
		Temperature: float64(21.3),
		Humidity:    uint8(74),
		Light:       uint8(61),
		Motion:      true,
		LowBattery:  false,
	}

	assert.Equal(t, node.TextData(), "temperature: 21.3 | humidity: 74 | light: 61 | motion: true | lowbat: false")

	node = &Node{
		Kind:        TINYTX_T_NODE,
		Temperature: float64(21.3),
		Vcc:         uint(3142),
	}

	assert.Equal(t, node.TextData(), "temperature: 21.3 | vcc: 3142")
}
