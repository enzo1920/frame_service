// Using as client mode to send data to your website.

// Simple demo for feeding some random data to Pachube.
// 2011-07-08 <jc@wippler.nl> http://opensource.org/licenses/mit-license.php

// Handle returning code and reset ethernet module if needed
// 2013-10-22 hneiraf@gmail.com

// Modifing so that it works on my setup for www.thingspeak.com.
// Arduino pro-mini 5V/16MHz, ETH modul on SPI with CS on pin 10.
// Also added a few changes found on various forums. Do not know what the 
// res variable is for, tweaked it so it works faster for my application
// 2015-11-09 dani.lomajhenic@gmail.com

#include <EtherCard.h>

// change these settings to match your own setup
//#define FEED "000"
#define APIKEY "beef1337beef1337" // put your key here
#define ethCSpin 10 // put your CS/SS pin here.

// ethernet interface mac address, must be unique on the LAN
static byte mymac[] = { 0x74,0x69,0x69,0x2D,0x30,0x31 };
const char website[] PROGMEM = "moty22.co.uk";  //Change to your domain name 
byte Ethernet::buffer[700];
uint32_t timer;
Stash stash;
byte session;
String d1;
//timing variable
int res = 100; // was 0
int analog=0;


void setup () {
  pinMode(3,INPUT_PULLUP);
  Serial.begin(9600);
  Serial.println("\n[ThingSpeak example]");

  //Initialize Ethernet
  initialize_ethernet();
}


void loop () { 
  //if correct answer is not received then re-initialize ethernet module
  if (res > 220){
    initialize_ethernet(); 
  }
  
  res = res + 1;
  
  ether.packetLoop(ether.packetReceive());
  
  //200 res = 30 seconds (150ms each res)
  if (res == 200) {

    analog = analogRead(A0);
    if(digitalRead(3)) {d1 = "OFF";} else {d1 = "ON";}  

    // generate 3 values as payload - by using a separate stash,
    // we can determine the size of the generated message ahead of time
    
    byte sd = stash.create();

    stash.print("field1=");
    stash.print(highByte(analog));
    stash.print("&field2=");
    stash.print(lowByte(analog));
    stash.print("&field3=");
    stash.print(d1);
    
    stash.save();

    // generate the header with payload - note that the stash size is used,
    // and that a "stash descriptor" is passed in as argument using "$H"
    Stash::prepare(PSTR("POST /a11.php HTTP/1.1" "\r\n"
      "Host: $F" "\r\n"
      "Connection: close" "\r\n"
      "X-THINGSPEAKAPIKEY: $F" "\r\n"
      "Content-Type: application/x-www-form-urlencoded" "\r\n"
      "Content-Length: $D" "\r\n"
      "\r\n"
      "$H"),
      website, PSTR(APIKEY), stash.size(), sd);

    // send the packet - this also releases all stash buffers once done
    session = ether.tcpSend(); 

 // added from: http://jeelabs.net/boards/7/topics/2241
// int freeCount = stash.freeCount();
//    if (freeCount <= 3) {   Stash::initMap(56); } 
  }
  
   const char* reply = ether.tcpReply(session);
   
   if (reply != 0) {
     res = 0;
     Serial.println(F(" >>>REPLY recieved...."));
     Serial.println(reply);
     delay(300);
   }
   delay(150);
}

void initialize_ethernet(void){  
  for(;;){ // keep trying until you succeed 

    if (ether.begin(sizeof Ethernet::buffer, mymac, ethCSpin) == 0){ 
      Serial.println( "Failed to access Ethernet controller");
      continue;
    }
    
    if (!ether.dhcpSetup()){
      Serial.println("DHCP failed");
      continue;
    }

    ether.printIp("IP:  ", ether.myip);
    ether.printIp("GW:  ", ether.gwip);  
    ether.printIp("DNS: ", ether.dnsip);  

    if (!ether.dnsLookup(website))
      Serial.println("DNS failed");

    ether.printIp("SRV: ", ether.hisip);

    //reset init value
    res = 180;
    break;
  }
}
