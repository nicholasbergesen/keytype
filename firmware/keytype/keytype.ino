// keytype firmware: a USB keyboard that types whatever the host sends it
// over the board's USB serial port.
//
// Runs on any ATmega32U4 board with native USB (DFRobot Beetle, Arduino
// Leonardo, SparkFun Pro Micro and clones). Flash it as "Arduino Leonardo".
//
// Protocol: one command per line ('\n'), one reply line per command.
//   P            ping            -> "OK keytype 1"
//   T<hex>       type bytes      -> "OK"   (ASCII, '\n' is Enter, '\t' is Tab)
//   K<hex>       press key codes together, then release all -> "OK"
//   D<ms>        delay after each key, 0-1000 ms (default 5) -> "OK"
//   R            release all keys -> "OK"
// Errors reply "ERR <reason>". Key codes are the Arduino Keyboard library's
// (0x80-0x87 modifiers, 0xB0+ special keys, ASCII for the rest).

#include <Keyboard.h>

const char VERSION[] = "keytype 1";
const size_t LINE_MAX = 130;   // 'T' plus 64 bytes of hex, with room to spare
const unsigned int HOLD_MS = 20;  // how long a key combination is held

char line[LINE_MAX + 1];
size_t lineLen = 0;
bool overflow = false;
unsigned int keyDelay = 5;

void setup() {
  Serial.begin(115200);
  Keyboard.begin();
}

void loop() {
  while (Serial.available() > 0) {
    char c = Serial.read();
    if (c == '\r') {
      continue;
    }
    if (c == '\n') {
      if (overflow) {
        Serial.println("ERR line too long");
      } else {
        line[lineLen] = '\0';
        handle(line, lineLen);
      }
      lineLen = 0;
      overflow = false;
      continue;
    }
    if (lineLen < LINE_MAX) {
      line[lineLen++] = c;
    } else {
      overflow = true;
    }
  }
}

int hexValue(char c) {
  if (c >= '0' && c <= '9') return c - '0';
  if (c >= 'a' && c <= 'f') return c - 'a' + 10;
  if (c >= 'A' && c <= 'F') return c - 'A' + 10;
  return -1;
}

// decodeHex turns the n hex digits in s into bytes, in place, and returns how
// many bytes it wrote, or -1 if s is not valid hex.
int decodeHex(char *s, size_t n) {
  if (n % 2 != 0) return -1;
  for (size_t i = 0; i < n; i += 2) {
    int hi = hexValue(s[i]);
    int lo = hexValue(s[i + 1]);
    if (hi < 0 || lo < 0) return -1;
    s[i / 2] = (char)((hi << 4) | lo);
  }
  return n / 2;
}

void handle(char *cmd, size_t n) {
  if (n == 0) {
    Serial.println("ERR empty command");
    return;
  }
  char *arg = cmd + 1;
  size_t argLen = n - 1;

  switch (cmd[0]) {
    case 'P':
      Serial.print("OK ");
      Serial.println(VERSION);
      return;

    case 'D': {
      if (argLen == 0 || argLen > 4) {
        Serial.println("ERR bad delay");
        return;
      }
      unsigned int ms = 0;
      for (size_t i = 0; i < argLen; i++) {
        if (arg[i] < '0' || arg[i] > '9') {
          Serial.println("ERR bad delay");
          return;
        }
        ms = ms * 10 + (arg[i] - '0');
      }
      if (ms > 1000) {
        Serial.println("ERR bad delay");
        return;
      }
      keyDelay = ms;
      Serial.println("OK");
      return;
    }

    case 'T': {
      int count = decodeHex(arg, argLen);
      if (count < 0) {
        Serial.println("ERR bad hex");
        return;
      }
      for (int i = 0; i < count; i++) {
        Keyboard.write((uint8_t)arg[i]);
        delay(keyDelay);
      }
      Serial.println("OK");
      return;
    }

    case 'K': {
      int count = decodeHex(arg, argLen);
      if (count <= 0) {
        Serial.println("ERR bad keys");
        return;
      }
      for (int i = 0; i < count; i++) {
        Keyboard.press((uint8_t)arg[i]);
      }
      delay(HOLD_MS);
      Keyboard.releaseAll();
      delay(keyDelay);
      Serial.println("OK");
      return;
    }

    case 'R':
      Keyboard.releaseAll();
      Serial.println("OK");
      return;

    default:
      Serial.println("ERR unknown command");
  }
}
