// Twitter client sketch for ENC28J60 based Ethernet Shield. Uses 
// arduino-tweet.appspot.com as a OAuth gateway.
// Step by step instructions:
// 
//  1. Get a oauth token:
//     http://arduino-tweet.appspot.com/oauth/twitter/login
//  2. Put the token value in the TOKEN define below
//  3. Run the sketch!
//
//  WARNING: Don't send more than 1 tweet per minute!
//  WARNING: This example uses insecure HTTP and not HTTPS.
//  The API key will be sent over the wire in plain text.
//  NOTE: Twitter rejects tweets with identical content as dupes (returns 403)

#include <EtherCard.h>
#include <OneWire.h>
#include <SPI.h>

// OAUTH key from http://arduino-tweet.appspot.com/


// ethernet interface mac address, must be unique on the LAN
byte mymac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xDE, 0x01 };
OneWire ds(7); // на пине 7 (нужен резистор 2.2 КОм)

const char website[] PROGMEM = "cloud.framecase.ru";

static byte session;

byte Ethernet::buffer[700];
static uint32_t timer;
 

Stash stash;

static void sendToFramecase () {
  Serial.println("Sending temp...");
  byte sd = stash.create();
  
  char PostData[8]; 
  DS18S20_read_temp(PostData);
  Serial.println(PostData);
  stash.println(PostData);
  stash.save();
  int stash_size = stash.size();

  // Compose the http POST request, taking the headers below and appending
  // previously created stash in the sd holder.
  Stash::prepare(PSTR("POST http://$F/v1/upload/temp/?token=5d6f3ecb1cb3d69b HTTP/1.0" "\r\n"
    "Host: $F" "\r\n"
    "Content-Length: $D" "\r\n"
    "\r\n"
    "$H"),
  website, website, stash_size, sd);

  // send the packet - this also releases all stash buffers once done
  // Save the session ID so we can watch for it in the main loop.
  session = ether.tcpSend();
  PostData[0] = (char)0;//characters are terminated by a zero byte, only the first byte needs to be zeroed
}

void setup () {
  Serial.begin(57600);
  Serial.println("\n[Framecase Client]");

  if (ether.begin(sizeof Ethernet::buffer, mymac) == 0) 
    Serial.println(F("Failed to access Ethernet controller"));
  if (!ether.dhcpSetup())
    Serial.println(F("DHCP failed"));

  ether.printIp("IP:  ", ether.myip);
  ether.printIp("GW:  ", ether.gwip);  
  ether.printIp("DNS: ", ether.dnsip);  

  if (!ether.dnsLookup(website))
    Serial.println(F("DNS failed"));

  ether.printIp("SRV: ", ether.hisip);

 
}

void DS18S20_read_temp(char * result){
   byte i;
   byte present = 0;
   byte type_s = 0;
   //byte type_s;
   byte data[12];
   byte addr[8];
   //char result[8];
   //float celsius, fahrenheit;
   if (!ds.search(addr)) {

         ds.reset_search();
         delay(250);
         return;
    }

   if (OneWire::crc8(addr, 7) != addr[7]) {

        return;
   }

    ds.reset();
    ds.select(addr);
    ds.write(0x44); // начинаем преобразование, используя ds.write(0x44,1) с "паразитным" питанием
    delay(750); // 750 

    present = ds.reset();
    ds.select(addr);
    ds.write(0xBE);

    for ( i = 0; i < 9; i++) { // нам необходимо 9 байт
          data[i] = ds.read();
        }
    
    int16_t raw = (data[1] << 8) | data[0];
    if (type_s) {
            raw = raw << 3; // разрешение 9 бит по умолчанию
            if (data[7] == 0x10) {
                  raw = (raw & 0xFFF0) + 12 - data[6];
            }
     } 
     else {
            byte cfg = (data[4] & 0x60);
            if (cfg == 0x00) raw = raw & ~7; // разрешение 9 бит, 93.75 мс
            else if (cfg == 0x20) raw = raw & ~3; // разрешение 10 бит, 187.5 мс
            else if (cfg == 0x40) raw = raw & ~1; // разрешение 11 бит, 375 мс
          
      }
      //celsius = (float)raw / 16.0;
      dtostrf((float)raw / 16.0, 6, 2, result);
      //return result;
}





void loop () {
  if (millis() > timer) {
    timer = millis() + 50000;
    sendToFramecase();
    Serial.println("send post to cloud.framecase.ru ");
  }

  ether.packetLoop(ether.packetReceive());

  const char* reply = ether.tcpReply(session);
  if (reply != 0) {
    Serial.println("Got a response!");
    Serial.println(reply);
  }
}

