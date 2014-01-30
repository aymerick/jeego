//
// Jeego Node - [Jeenode] Temperature Humidity Light
//
// Sensors:
//  - DHT22 for temperature and humidity
//  - LDR for light
//
// References:
//  - http://jeelabs.net/projects/hardware/wiki/JeeNode
//  - http://jeelabs.org/2011/12/13/developing-a-low-power-sketch/
//  - https://github.com/nathanchantrell/TinyTX/blob/master/TinyTX_DHT22/TinyTX_DHT22.ino
//  - https://github.com/mharizanov/TinySensor/blob/master/Funky_DHT22/Funky_DHT22.ino
//
// + Jeenode design by JC Wippler: http://jeelabs.net/projects/hardware/wiki/JeeNode
//

#include <JeeLib.h>
#include <DHT22.h>
#include <avr/sleep.h>

// has to be defined because we're using the watchdog for low-power waiting
ISR(WDT_vect) { Sleepy::watchdogEvent(); }


#define DEBUG 0
#define NOOP 0

// Node kind
#define NODE_KIND 2

// DHT22 Power wire is plugged into jeenode DI03 (arduino: digital 6)
#define DHT22_POWER_PIN 6

// DHT22 Data wire is plugged into jeenode DI02 (arduino: digital 5)
#define DHT22_DATA_PIN 5

// LDR wire is plugged into jeenode AIO04
#define LDR_PORT 4

// how often to report, in minutes
#define REPORT_PERIOD 5

// set the sync mode to 2 if the fuses are still the Arduino default
// mode 3 (full powerdown) can only be used with 258 CK startup fuses
#define RADIO_SYNC_MODE 2

// number of milliseconds to wait for an ack
#define ACK_TIME 10

// how soon to retry if ACK didn't come in
#define ACK_RETRY_PERIOD 10

// maximum number of times to retry
#define ACK_RETRY_LIMIT 5


static byte myNodeID;

// serialized payload
struct {
  byte kind     :7;  // Node kind
  byte reserved :1;  // Reserved for future use. Must be zero.
  byte light;        // Light sensor: 0..255
  byte lowbat   :1;  // Supply voltage dropped under 3.1V: 0..1
  byte humi     :7;  // Humidity: 0..100
  int  temp     :10; // Temperature: -512..+512 (tenths)
} payload;

// sensors
DHT22 sensorDHT22(DHT22_DATA_PIN);
Port  sensorLDR(LDR_PORT);


//
// Helpers
//

static void serialFlush () {
#if ARDUINO >= 100
  Serial.flush();
#endif

  // make sure tx buf is empty before going back to sleep
  delay(2);
}

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
static void sendPayload() {
  for (byte i = 0; i < ACK_RETRY_LIMIT; i++) {
    // power up RF
    rf12_sleep(RF12_WAKEUP);

    // send payload
    rf12_sendNow(RF12_HDR_ACK, &payload, sizeof payload);
    rf12_sendWait(RADIO_SYNC_MODE);

    // wait for ack
    byte acked = waitForAck();

    // power down RF
    rf12_sleep(RF12_SLEEP);

    if (acked) {
#if DEBUG
      Serial.print(" ack");
      Serial.println((int) i);
      serialFlush();
#endif
      return;
    }

    // If no ack received wait and try again
    delay(ACK_RETRY_PERIOD * 100);
  }

#if DEBUG
  Serial.println(" no ack!");
  serialFlush();
#endif
}

// send payload and optionally report on serial port
static void doReport() {
#if DEBUG
  Serial.print("jeeThlNode ");
  Serial.print((int) payload.light);
  Serial.print(' ');
  Serial.print((int) payload.lowbat);
  Serial.print(' ');
  Serial.print((int) payload.humi);
  Serial.print(' ');
  Serial.print((int) payload.temp);
  Serial.println();
  serialFlush();
#endif

#if !NOOP
  sendPayload();
#endif
}


//
// Sensors
//


// read LDR data
void readLDR() {
  // enable AIO pull-up
  sensorLDR.digiWrite2(1);

  payload.light = ~ sensorLDR.anaRead() >> 2;

  // disable pull-up
  sensorLDR.digiWrite2(0);
}

// read Battery status
void readLowBat() {
  payload.lowbat = rf12_lowbat();
}

// read DHT22 data
void readDHT22() {
  DHT22_ERROR_t errorCode;

  // power on sensor
  digitalWrite(DHT22_POWER_PIN, HIGH);

  // wait for sensor warm-up
  Sleepy::loseSomeTime(2000);

  // read data
  errorCode = sensorDHT22.readData();

  // power off sensor
  digitalWrite(DHT22_POWER_PIN, LOW);

  // handle data
  switch(errorCode)
  {
    case DHT_ERROR_NONE:
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
      break;
#if DEBUG
    case DHT_ERROR_CHECKSUM:
      Serial.print("check sum error ");
      Serial.print(sensorDHT22.getTemperatureC());
      Serial.print("C ");
      Serial.print(sensorDHT22.getHumidity());
      Serial.println("%");
      break;
    case DHT_BUS_HUNG:
      Serial.println("BUS Hung");
      break;
    case DHT_ERROR_NOT_PRESENT:
      Serial.println("Not present");
      break;
    case DHT_ERROR_ACK_TOO_LONG:
      Serial.println("ACK timeout");
      break;
    case DHT_ERROR_SYNC_TIMEOUT:
      Serial.println("Sync timeout");
      break;
    case DHT_ERROR_DATA_TIMEOUT:
      Serial.println("Data timeout");
      break;
    case DHT_ERROR_TOOQUICK:
      Serial.println("Polled too quick");
      break;
#endif
  }
}


//
// Main
//

void setup() {
#if DEBUG
  Serial.begin(57600);
  Serial.print("\n[jeeThlNode]");
  myNodeID = rf12_config();
  serialFlush();
#else
  // don't report info on the serial port
  myNodeID = rf12_config(0);
#endif

  // power down RF
  rf12_sleep(RF12_SLEEP);

  // set output mode for DHT22 power pin
  pinMode(DHT22_POWER_PIN, OUTPUT);

  // init payload
  payload.reserved = 0;
  payload.kind = NODE_KIND;
}

void loop() {
  // read sensors
  readLDR();
  readLowBat();
  readDHT22();

  // report
  doReport();

  // sleep
  for (byte i = 0; i < REPORT_PERIOD; i++) {
    // max value is 60 seconds
    Sleepy::loseSomeTime(60000);
  }
}
