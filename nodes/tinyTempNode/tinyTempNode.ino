//
// Jeego Node - [TinyTX] Temperature (http://nathan.chantrell.net/tinytx-wireless-sensor/)
//
// Original sketch: https://github.com/nathanchantrell/TinyTX/blob/master/TinyTX_DS18B20/TinyTX_DS18B20.ino
//
// Sensors:
//  - DS18B20 for temperature
//
// Arduino IDE Setup:
//   - Install OneWire lib into ~/Documents/Arduino/libraries/OneWire/
//       http://www.pjrc.com/teensy/arduino_libraries/OneWire.zip
//     Open OneWire.h and below the line:
//       #include “Arduino.h”       // for delayMicroseconds, digitalPinToBitMask, etc
//     add:
//       #include “pins_arduino.h”  // for digitalPinToBitMask, etc
//
//   - Install DallasTemperature lib into ~/Documents/Arduino/libraries/DallasTemperature/
//       http://download.milesburton.com/Arduino/MaximTemperature/DallasTemperature_LATEST.zip
//
//   - Install Jeelib into ~/Documents/Arduino/libraries/jeelib/
//       https://github.com/jcw/jeelib
//
// Programmer setup:
//   - Use an ISP USB programmer
//   - Setup an ISP adapter: http://www.evilmadscientist.com/2007/using-avr-microcontrollers-minimalist-target-boards/
//
// Program ATTiny84:
//   - Setup ATTiny support in Arduino thanks to: http://code.google.com/p/arduino-tiny
//   - Select Tools/Board => ATtiny84 @ 8Mhz (internal oscillator; BOD disabled)
//   - Select Tools/Programmer/USBasp
//
//   - Select Tools/Burn Bootloader. Note that this isn’t actually burning a bootloader to the ATtiny (it doesn’t use one),
//     it is just using this function to set the AVR fuses to configure the oscillator at 8MHz.
//
//   - Upload that sketch
//
// Build circuit:
//   - Try on a stripboard with: http://nathan.chantrell.net/downloads/arduino/tinytx/tinytx_stripboard_ds18b20.png
//
//   - If the sketch is stuck during call to rf12_sendWait then Burn Bootloader to fix the issue\
//

//----------------------------------------------------------------------------------------------------------------------
// TinyTX - An ATtiny84 and RFM12B Wireless Temperature Sensor Node
// By Nathan Chantrell. For hardware design see http://nathan.chantrell.net/tinytx
//
// Using the Dallas DS18B20 temperature sensor
//
// Licenced under the Creative Commons Attribution-ShareAlike 3.0 Unported (CC BY-SA 3.0) licence:
// http://creativecommons.org/licenses/by-sa/3.0/
//
// Requires Arduino IDE with arduino-tiny core: http://code.google.com/p/arduino-tiny/
// and small change to OneWire library, see: http://arduino.cc/forum/index.php/topic,91491.msg687523.html#msg687523
//----------------------------------------------------------------------------------------------------------------------

#include <OneWire.h>           // http://www.pjrc.com/teensy/arduino_libraries/OneWire.zip
#include <DallasTemperature.h> // http://download.milesburton.com/Arduino/MaximTemperature/DallasTemperature_LATEST.zip
#include <JeeLib.h>            // https://github.com/jcw/jeelib

ISR(WDT_vect) { Sleepy::watchdogEvent(); } // interrupt handler for JeeLabs Sleepy power saving

#define myNodeID 30       // RF12 node ID in the range 1-30
#define network 212       // RF12 Network group
#define freq RF12_868MHZ  // Frequency of RFM12B module

#define USE_ACK           // Enable ACKs, comment out to disable
#define RETRY_PERIOD 5    // How soon to retry (in seconds) if ACK didn't come in
#define RETRY_LIMIT 5     // Maximum number of times to retry
#define ACK_TIME 10       // Number of milliseconds to wait for an ack

#define ONE_WIRE_BUS 10   // DS18B20 Temperature sensor is connected on D10/ATtiny pin 13
#define ONE_WIRE_POWER 9  // DS18B20 Power pin is connected on D9/ATtiny pin 12

OneWire oneWire(ONE_WIRE_BUS); // Setup a oneWire instance

DallasTemperature sensors(&oneWire); // Pass our oneWire reference to Dallas Temperature


// Data Structure to be sent
typedef struct {
  int temp;    // Temperature reading
  int supplyV; // Supply voltage
} Payload;

Payload tinytx;


#ifdef USE_ACK

// Wait a few milliseconds for proper ACK
static byte waitForAck() {
  MilliTimer ackTimer;
  while (!ackTimer.poll(ACK_TIME)) {
    if (rf12_recvDone() && rf12_crc == 0 &&
      rf12_hdr == (RF12_HDR_DST | RF12_HDR_CTL | myNodeID))
      return 1;
    }
  return 0;
}

static void rfwrite(){
  // tx and wait for ack up to RETRY_LIMIT times
  for (byte i = 0; i <= RETRY_LIMIT; ++i) {
    // Wake up RF module
    rf12_sleep(-1);

    while (!rf12_canSend())
      rf12_recvDone();

    rf12_sendStart(RF12_HDR_ACK, &tinytx, sizeof tinytx);

    // Wait for RF to finish sending while in standby mode
    rf12_sendWait(2);

    // Wait for ACK
    byte acked = waitForAck();

    // Put RF module to sleep
    rf12_sleep(0);

    // Return if ACK received
    if (acked) { return; }

    // If no ack received wait and try again
    Sleepy::loseSomeTime(RETRY_PERIOD * 1000);
  }
}

#else

static void rfwrite(){
  // Wake up RF module
  rf12_sleep(-1);

  while (!rf12_canSend())
    rf12_recvDone();

  rf12_sendStart(0, &tinytx, sizeof tinytx);

  // Wait for RF to finish sending while in standby mode
  rf12_sendWait(2);

  // Put RF module to sleep
  rf12_sleep(0);

  return;
}

#endif


// Read current supply voltage
 long readVcc() {
   bitClear(PRR, PRADC); ADCSRA |= bit(ADEN); // Enable the ADC
   long result;
   // Read 1.1V reference against Vcc
   #if defined(__AVR_ATtiny84__)
    ADMUX = _BV(MUX5) | _BV(MUX0); // For ATtiny84
   #else
    ADMUX = _BV(REFS0) | _BV(MUX3) | _BV(MUX2) | _BV(MUX1);  // For ATmega328
   #endif
   delay(2); // Wait for Vref to settle
   ADCSRA |= _BV(ADSC); // Convert
   while (bit_is_set(ADCSRA,ADSC));
   result = ADCL;
   result |= ADCH<<8;
   result = 1126400L / result; // Back-calculate Vcc in mV
   ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC); // Disable the ADC to save power
   return result;
}

void setup() {
  // Initialize RFM12 with settings defined above
  rf12_initialize(myNodeID, freq, network);

  // Put the RFM12 to sleep
  rf12_sleep(0);

  // set power pin for DS18B20 to output
  pinMode(ONE_WIRE_POWER, OUTPUT);

  // only keep timer 0 going
  PRR = bit(PRTIM1);

  // Disable the ADC to save power
  ADCSRA &= ~ bit(ADEN); bitSet(PRR, PRADC);
}

void loop() {
  // turn DS18B20 sensor on
  digitalWrite(ONE_WIRE_POWER, HIGH);

  //Sleepy::loseSomeTime(5); // Allow 5ms for the sensor to be ready
  delay(5); // The above doesn't seem to work for everyone (why?)

  //start up temp sensor
  sensors.begin();

  // Get the temperature
  sensors.requestTemperatures();

  // Read first sensor and convert to integer, reversed at receiving end
  tinytx.temp = (sensors.getTempCByIndex(0)*100);

  // turn DS18B20 off
  digitalWrite(ONE_WIRE_POWER, LOW);

  // Get supply voltage
  tinytx.supplyV = readVcc();

  // Send data via RF
  rfwrite();

  // JeeLabs power save function: enter low power mode for 60 seconds (valid range 16-65000 ms)
  Sleepy::loseSomeTime(60000);
}
