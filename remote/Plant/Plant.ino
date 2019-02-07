#include "DHT.h"

// Input
#define DHT_IN 2       // The DHT pin to use
#define DHT_TYPE DHT11 // The type of DHT sensor
#define LIGHT_IN A0    // The pin to read the soil moisture on
#define SOIL_IN A2     // The pin to read the soil moisture on

// Output
#define LED_PIN 10      // The LED pin
#define SOIL_POWER 8    // The pin to turn on the soil sensor
#define PUMP_1_PIN 9    // The pump pin
#define PUMP_2_PIN 11   // The pump pin

// Read times (in seconds)
#define DHT_READ 2
#define SOIL_READ 900
int pinOnTimes[] = {-1, -1};

// Internal variables
DHT dht(DHT_IN, DHT_TYPE);    // Allows reading the DHT11 sensor
bool ledOn = false;           // Whether the LED is on or not
int loopCount = 0;            // The current loop count (always increases)
String incomingCommand = "";  // The incoming command
bool commandComplete = false; // Whether the command is complete and ready to be processed
int pins[] = {PUMP_1_PIN, PUMP_2_PIN};    // The associated pump pins
bool pinOn = false;           // Whether to check any of the pins
int photoresistor = 0;        // The last reading from the photoresistor
int soilValue = 0;            // The last reading from the soil sensor
int soilTime = 0;             // The cycles remaining until we read the soil sensor again
int numberOfPins = (sizeof(pins)/sizeof(pins[0])); // The number of Pins to set

void setup()
{
  dht.begin();
  incomingCommand.reserve(200);

  // Tell the monitor what we are exposing
  Serial.begin(9600);
  sendDetails();

  // Initialise the pins
  pinMode(LED_PIN, OUTPUT);
  digitalWrite(LED_PIN, LOW);
  int i = 0;
  for (i = 0; i < numberOfPins; i++) {
    pinMode(pins[i], OUTPUT);
    digitalWrite(pins[i], HIGH);
  }

  pinMode(SOIL_POWER, OUTPUT);
  digitalWrite(SOIL_POWER, LOW);
}

void loop()
{
  // Flash the LED on a 1s on/1s off freq
  delay(1000);
  ledOn = !ledOn;
  if (ledOn)
  {
    digitalWrite(LED_PIN, HIGH);
  }
  else
  {
    digitalWrite(LED_PIN, LOW);
  }

  loopCount++;

  if (pinOn)
  {
    checkPinTimes();
  }

  if (commandComplete)
  {
    if (incomingCommand[0] == 'C')
    {
      Serial.println(incomingCommand);
      int pinToChange = (int)incomingCommand[2] - 48;
      bool setPinOn = incomingCommand[3] == '+';
      if (incomingCommand.length() > 4)
      {
        pinOnTimes[pinToChange] = incomingCommand.substring(4).toInt();
      }
      Serial.print("A:");
      Serial.print(pinToChange);
      if (setPinOn)
      {
        digitalWrite(pins[pinToChange], LOW);
        Serial.println("+");
      }
      else
      {
        digitalWrite(pins[pinToChange], HIGH);
        Serial.println("-");
      }
      pinOn = true;
    } else if (incomingCommand[0] = 'I') {
      sendDetails();
    }

    // Clear the last command so we can receive another one
    incomingCommand = "";
    commandComplete = false;
  }

  if (soilTime <= 0)
  {
    Serial.println("R:soil");
    soilTime = SOIL_READ;
    readSoil();
  }
  else
  {
    soilTime--;
  }

  // Only read at specified intervals
  if (loopCount % DHT_READ == 0)
  {
    unsigned long timeSinceStart = millis();
    float humidity = dht.readHumidity();
    float temp = dht.readTemperature();

    // Check if any reads failed and exit early (to try again).
    if (isnan(humidity) || isnan(temp))
    {
      Serial.println("E:DHT");
      loopCount--; // Need to decrement otherwise it will wait the full five seconds
      return;
    }

    // Compute heat index in Celsius (isFahreheit = false)
    float hic = dht.computeHeatIndex(temp, humidity, false);

    // Read the light level
    photoresistor = analogRead(LIGHT_IN);

    // Send the data to the monitor
    Serial.print("D:");
    Serial.print(timeSinceStart);
    Serial.print(",");
    Serial.print(humidity);
    Serial.print(",");
    Serial.print(temp);
    Serial.print(",");
    Serial.print(hic);
    Serial.print(",");
    Serial.print(photoresistor);
    Serial.print(",");
    Serial.print(soilValue);
    Serial.println();
  }
}

void sendDetails()
{
  Serial.println("=====");      // Clear any pending output
  Serial.println("O:time,humidity,tempC,heatIndC,light,soil");
  Serial.println("I:Pump 1,Pump 2");
}

void checkPinTimes()
{
    Serial.print("Pin:");
    Serial.println(pinOn);
    Serial.println(numberOfPins);
  int i;
  pinOn = false;
  for (i = 0; i < numberOfPins; i++)
  {
    Serial.print("Loop:");
    Serial.print(i);
    Serial.print(",");
    Serial.print(pinOnTimes[i]);
    Serial.print(",");
    Serial.println(pinOn);
    if (pinOnTimes[i] > 0)
    {
      pinOn = true;
      pinOnTimes[i]--;
    }
    else if (pinOnTimes[i] == 0)
    {
      digitalWrite(pins[i], HIGH);
      pinOnTimes[i]--;
      Serial.print("A:");
      Serial.print(i);
      Serial.println("-");
    }
  }
}

void serialEvent()
{
  while (Serial.available() && !commandComplete)
  {
    char inChar = (char)Serial.read();
    incomingCommand += inChar;
    // If the incoming character is a newline, set a flag so the main loop can process it
    if (inChar == '\n')
    {
      commandComplete = true;
    }
  }
}

int readSoil()
{
  digitalWrite(SOIL_POWER, HIGH);
  delay(10);
  soilValue = analogRead(SOIL_IN);
  digitalWrite(SOIL_POWER, LOW);
}
