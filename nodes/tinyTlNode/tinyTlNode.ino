//
// Jeego Node - [TinyTX] Temperature and Light
//
// Sensors:
//  - DS18B20 for temperature
//  - LDR for light
//
// + TinyTX design by Nathan Chantrell: http://nathan.chantrell.net/tinytx-wireless-sensor/
// + Original sketch: https://github.com/nathanchantrell/TinyTX/blob/master/TinyTX_DS18B20/TinyTX_DS18B20.ino
//

#include <JeeLib.h>            // https://github.com/jcw/jeelib
#include <OneWire.h>           // http://www.pjrc.com/teensy/arduino_libraries/OneWire.zip
#include <DallasTemperature.h> // http://download.milesburton.com/Arduino/MaximTemperature/DallasTemperature_LATEST.zip

// has to be defined because we're using the watchdog for low-power waiting
ISR(WDT_vect) { Sleepy::watchdogEvent(); }


// Node kind
#define NODE_KIND 5

// RF12 node ID in the range 1-30
#define myNodeID 27

// RF12 Network group
#define network 212

// frequency of RFM12B module
#define freq RF12_868MHZ

// DS18B20 Temperature sensor is connected on D10/ATtiny pin 13
#define ONE_WIRE_BUS_PIN 10

// DS18B20 Power pin is connected on D9/ATtiny pin 12
#define ONE_WIRE_POWER_PIN 9

// LDR is connected on A3
#define LDR_PIN 3
// #define LDR_PIN 2

// LDR Power pin is connected on D3
#define LDR_POWER_PIN 3
//#define LDR_POWER_PIN 7

// how often to report, in minutes
#define REPORT_PERIOD 5

// set the sync mode to 2 if the fuses are still the Arduino default
// mode 3 (full powerdown) can only be used with 258 CK startup fuses
#define RADIO_SYNC_MODE 2

// number of milliseconds to wait for an ack
#define ACK_TIME 10

// how soon to retry if ACK didn't come in
#define ACK_RETRY_PERIOD 10

// Maximum number of times to retry
#define ACK_RETRY_LIMIT 5  


// serialized payload
struct {
  byte kind     :7;  // Node kind
  byte reserved :1;  // Reserved for future use. Must be zero.
  int  vcc      :12; // Supply voltage: < 4096 mv
  int  temp     :10; // Temperature: -512..+512 (tenths)
  int  light    :8;  // Light sensor: 0..255
} payload;


// sensors
OneWire oneWire(ONE_WIRE_BUS_PIN);
DallasTemperature sensors(&oneWire);


// wait a few milliseconds for proper ACK to me, return true if indeed received
static byte waitForAck() {
  MilliTimer ackTimer;

  while (!ackTimer.poll(ACK_TIME)) {
    // see http://talk.jeelabs.net/topic/811#post-4712
    if (rf12_recvDone() && (rf12_crc == 0) && (rf12_hdr == (RF12_HDR_DST | RF12_HDR_CTL | myNodeID)))
      return 1;
  }

  return 0;
}

// send payload and wait for master node ack
static void sendPayload(){
  for (byte i = 0; i <= ACK_RETRY_LIMIT; i++) {
    // power up RF
    rf12_sleep(RF12_WAKEUP);

    // send payload
    rf12_sendNow(RF12_HDR_ACK, &payload, sizeof payload);
    rf12_sendWait(RADIO_SYNC_MODE);

    // wait for ack
    byte acked = waitForAck();

    // power down RF
    rf12_sleep(RF12_SLEEP);

    // return if ACK received
    if (acked) {
      return;
    }

    // if no ack received wait and try again
    delay(ACK_RETRY_PERIOD * 100);
  }
}

// Read current supply voltage in millivolts
long readVcc() {
  long result;

  // Enable the ADC
  bitClear(PRR, PRADC); ADCSRA |= bit(ADEN);

  // Read 1.1V reference against Vcc
  // set the reference to Vcc and the measurement to the internal 1.1V reference
#if defined(__AVR_ATtiny84__)
  // For ATtiny84
  ADMUX = _BV(MUX5) | _BV(MUX0);
#else
  // For ATmega328
  ADMUX = _BV(REFS0) | _BV(MUX3) | _BV(MUX2) | _BV(MUX1);
#endif

  // Wait for Vref to settle
  delay(2);

  // Start conversion
  ADCSRA |= _BV(ADSC);

  // measuring
  while (bit_is_set(ADCSRA,ADSC));

  result = ADCL;
  result |= ADCH<<8;

  // Calculate Vcc (in mV); 1125300 = 1.1*1023*1000
  result = 1126400L / result;

  // Disable the ADC to save power
  ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC);

  return result;
}

// read OneWire data
void readOneWire() {
  // power on sensor
  digitalWrite(ONE_WIRE_POWER_PIN, HIGH);

  // allow 5ms for the sensor to be ready
  delay(5);

  // start up temp sensor
  sensors.begin();

  // get the temperature
  sensors.requestTemperatures();

  // temperature value is send in payload on 10bits, so we keep only temperatures between -51.2 and 51.2
  short int temp = (sensors.getTempCByIndex(0)*10);

  if (temp > 512) {
    temp = 512;
  }
  if (temp < -512) {
    temp = -512;
  }

  payload.temp = temp;

  // power off sensor
  digitalWrite(ONE_WIRE_POWER_PIN, LOW);
}

// read LDR data
void readLDR() {
  digitalWrite(LDR_POWER_PIN, HIGH);
  delay(10);

  // Enable the ADC
  bitClear(PRR, PRADC); ADCSRA |= bit(ADEN);

  payload.light = ~ analogRead(LDR_PIN) >> 2;

  // Disable the ADC to save power
  ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC);

  digitalWrite(LDR_POWER_PIN, LOW);
}


//
// Main
//

void setup() {
  // initialize RFM12
  rf12_initialize(myNodeID, freq, network);

  // power down RF
  rf12_sleep(0);

  // set power pin for DS18B20 to output
  pinMode(ONE_WIRE_POWER_PIN, OUTPUT);

  // set power pin for LDR to output
  pinMode(LDR_POWER_PIN, OUTPUT);

  // only keep timer 0 going
  PRR = bit(PRTIM1);

  // Disable the ADC to save power
  ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC);

  // init payload
  payload.reserved = 0;
  payload.kind = NODE_KIND;
}

void loop() {
  // read sensors
  readOneWire();
  readLDR();

  // get supply voltage
  payload.vcc = readVcc();

  // send data
  sendPayload();

  // sleep
  for (byte i = 0; i < REPORT_PERIOD; i++) {
    // max value is 60 seconds
    Sleepy::loseSomeTime(60000);
  }
}
