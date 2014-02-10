//
// Jeego Node - [TinyTX] Temperature Humidity
//
// Sensors:
//  - DHT22 for temperature and humidity
//
// + TinyTX design by Nathan Chantrell: http://nathan.chantrell.net/tinytx-wireless-sensor/
// + Original sketch: https://github.com/nathanchantrell/TinyTX/blob/master/TinyTX_DHT22/TinyTX_DHT22.ino
//

#include <DHT22.h>  // https://github.com/nathanchantrell/Arduino-DHT22
#include <JeeLib.h> // https://github.com/jcw/jeelib

// has to be defined because we're using the watchdog for low-power waiting
ISR(WDT_vect) { Sleepy::watchdogEvent(); }


// Node kind
#define NODE_KIND 4

// RF12 node ID in the range 1-30
#define myNodeID 28

// RF12 Network group
#define network 212

// frequency of RFM12B module
#define freq RF12_868MHZ

// DHT22 Temperature sensor is connected on D10/ATtiny pin 13
#define DHT22_BUS_PIN 10

// DHT22 Power pin is connected on D9/ATtiny pin 12
#define DHT22_POWER_PIN 9

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
  byte humi     :7;  // Humidity: 0..100
} payload;


// sensors
DHT22 sensorDHT22(DHT22_BUS_PIN);


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

// read DHT22 data
boolean readDHT22() {
  DHT22_ERROR_t errorCode;

  // power on sensor
  digitalWrite(DHT22_POWER_PIN, HIGH);

  // wait for sensor warm-up
  Sleepy::loseSomeTime(2000);

  // read data
  errorCode = sensorDHT22.readData();

  // power off sensor
  digitalWrite(DHT22_POWER_PIN, LOW);

  if (errorCode == DHT_ERROR_NONE) {
    short int temp;

    // temperature value is send in payload on 10bits, so we keep only temperatures between -51.2 and 51.2
    temp = sensorDHT22.getTemperatureCInt();
    if (temp > 512) {
      temp = 512;
    }
    if (temp < -512) {
      temp = -512;
    }

    payload.temp = temp;
    payload.humi = sensorDHT22.getHumidityInt() / 10;

    return true;
  }

  return false;
}


//
// Main
//

void setup() {
  // initialize RFM12
  rf12_initialize(myNodeID, freq, network);

  // power down RF
  rf12_sleep(0);

  // set power pin for DHT22 to output
  pinMode(DHT22_POWER_PIN, OUTPUT);

  // only keep timer 0 going
  PRR = bit(PRTIM1);

  // Disable the ADC to save power
  ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC);

  // init payload
  payload.reserved = 0;
  payload.kind = NODE_KIND;
}

void loop() {
  // read sensor
  if (readDHT22()) {
    // get supply voltage
    payload.vcc = readVcc();

    // send data
    sendPayload();
  }

  // sleep
  for (byte i = 0; i < REPORT_PERIOD; i++) {
    // max value is 60 seconds
    Sleepy::loseSomeTime(60000);
  }
}
