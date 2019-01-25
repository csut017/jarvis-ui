#include "DHT.h"

// Input
#define DHT_PIN     8           // The pin to use
#define DHT_TYPE    DHT11       // The type of sensor

// Output
#define LED_PIN     4           // The LED pin
#define PUMP_1_PIN  2           // The pump pin

// Read times (in seconds)
#define DHT_READ    5

// Internal variables
DHT dht(DHT_PIN, DHT_TYPE);     // Allows reading the DHT11 sensor
bool ledOn = false;             // Whether the LED is on or not
int loopCount = 0;              // The current loop count (always increases)
String incomingCommand = "";    // The incoming command
bool commandComplete = false;   // Whether the command is complete and ready to be processed

void setup() {
  dht.begin();
  incomingCommand.reserve(200);

  // Tell the monitor what we are exposing
  Serial.begin(9600);
  Serial.println("O:Time since start (s),Humidity (%),Temperature (C),Heat Index (C)");
  Serial.println("I:Water pump");

  // Initialise the pins
  pinMode(LED_PIN, OUTPUT);
  digitalWrite(LED_PIN, LOW);
  pinMode(PUMP_1_PIN, OUTPUT);
  digitalWrite(PUMP_1_PIN, HIGH);
}

void loop() {
  // Flash the LED on a 1s on/1s off freq
  delay(1000);
  ledOn = !ledOn;
  if (ledOn) {
    digitalWrite(LED_PIN, HIGH);
  } else {
    digitalWrite(LED_PIN, LOW);
  }

  loopCount++;
  
  // Only read at specified intervals
  if (loopCount % DHT_READ == 0) {
    unsigned long timeSinceStart = millis();
    float humidity = dht.readHumidity();
    float temp = dht.readTemperature();
  
    // Check if any reads failed and exit early (to try again).
    if (isnan(humidity) || isnan(temp)) {
      loopCount--;        // Need to decrement otherwise it will wait the full five seconds
      return;
    }
  
    // Compute heat index in Celsius (isFahreheit = false)
    float hic = dht.computeHeatIndex(temp, humidity, false);

    // Send the data to the monitor
    Serial.print("D:");
    Serial.print(timeSinceStart);
    Serial.print(",");
    Serial.print(humidity);
    Serial.print(",");
    Serial.print(temp);
    Serial.print(",");
    Serial.print(hic);
    Serial.println();
  }

  if (commandComplete) {
    Serial.print("C:");
    Serial.println(incomingCommand);
    // Clear the last command so we can receive another one
    incomingCommand = "";
    commandComplete = false;
  }
}

void serialEvent() {
  while (Serial.available() && !commandComplete) {
    char inChar = (char)Serial.read();
    incomingCommand += inChar;
    // If the incoming character is a newline, set a flag so the main loop can process it
    if (inChar == '\n') {
      commandComplete = true;
    }
  }
}
