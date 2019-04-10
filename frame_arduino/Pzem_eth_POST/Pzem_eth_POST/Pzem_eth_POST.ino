#include <UIPEthernet.h> // Used for Ethernet
#include <SPI.h>
#include <SoftwareSerial.h> 
#include "PZEM004T.h"

PZEM004T pzem(3,2);
IPAddress pzip(192,168,1,1);

byte mac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xDE, 0x01 }; // RESERVED MAC ADDRESS
EthernetClient client;


// **** ETHERNET SETTING ****
//byte mac[] = { 0x90, 0xA2, 0xDA, 0x0D, 0x78, 0xEE  };                                       
IPAddress ip(10, 10, 10, 244); 
IPAddress mydns(10,10,10,1);
IPAddress mygw(10,10,10,1);
IPAddress subnet(255,255,255,0);

void setup() { 
	Serial.begin(9600);

  pzem.setAddress(pzip);

	if (!Ethernet.begin(mac) ) {
		Serial.println("Failed to configure Ethernet using DHCP"); 
    Ethernet.begin(mac, ip,mydns,mygw,subnet);
   
	}
  //Ethernet.begin(mac, ip);
  Serial.print("ip-");
  Serial.println( Ethernet.localIP());
  Serial.print("Subnet mask-");
  Serial.println( Ethernet.subnetMask());
  Serial.print("Gateway-");
  Serial.println( Ethernet.gatewayIP());
  Serial.print("DNS-");
  Serial.println( Ethernet.dnsServerIP());
	
}

void loop(){

  sendPOST();
  delay(7000); // WAIT   BEFORE SENDING AGAIN
}


void sendPOST() //client function to send/receive GET request data.
{
    String current_pzem = Pzem();
    Serial.println(current_pzem);

   if(current_pzem.length()>=0){
       if (client.connect("framecase.tula.su",8080)) { // REPLACE WITH YOUR SERVER ADDRESS
          Serial.println("connected");
          Serial.println("=================>");
          client.println("POST /v1/upload/voltage/?token={token} HTTP/1.1"); 
          client.println("Host: framecase.tula.su"); // SERVER ADDRESS HERE TOO
          client.println("Content-Type: text/plain;");
          client.print("Content-Length: "); 
          client.println(current_pzem.length()); 
          client.println(); 
          client.print(current_pzem);
          Serial.println("disconnecting.");
          client.stop(); //stop client
          
          } 
       else {
          Serial.println("connection failed"); //error message if no client connect
       }
        //Serial.println("len str voltage"); //error message if no client connect
        delay(100);
   }

}

String Pzem(){
  String volt;
  int ivoltage;
  float v = pzem.voltage(pzip);
  //Serial.print(v);Serial.print("V; ");
  if (v <= 0.0){
     volt="0.00";
  } 
  else {
    ivoltage = v*10;
    volt = String(ivoltage/10, DEC);
    volt += ".";
    volt += String(ivoltage%10, DEC);
    //Serial.print(currency);
    //Serial.println();
  }
  return volt;
 }

