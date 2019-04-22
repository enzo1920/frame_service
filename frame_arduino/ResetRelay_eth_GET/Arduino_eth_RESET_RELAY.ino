#include <UIPEthernet.h> // Used for Ethernet
#include <OneWire.h>
#include <SPI.h>



byte mac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xEE, 0x01 };

EthernetClient client;
char serverName[] = "server"; // server
int buff=0;

// Relay state and pin
String relay1State = "Off";
const int relay = 4;


IPAddress ip(10, 10, 10, 227); 
IPAddress mydns(10,10,10,1);
IPAddress mygw(10,10,10,1);
IPAddress subnet(255,255,255,0);

void setup()
{
  Serial.begin(9600);
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
  
  // Relay module prepared 
  pinMode(relay, OUTPUT);
  digitalWrite(relay, HIGH);
}

void loop()
{
	sendGET();
  //2 min = 2 min * 60 sec/min * 1000 msec/sec = 120,000msec
  delay(120000); 
}

void sendGET() //client function to send/receive GET request data.
{
 if (client.connect(serverName,8080)>=0) { // REPLACE WITH YOUR SERVER ADDRESS
    client.print("GET /v1/get/relay/reset/?token={token} HTTP/1.1");
    client.println("Host: reseter"); // SERVER ADDRESS HERE TOO
    client.println("Content-Type: text/plain;");
    client.print("Content-Length: "); 
    client.println("Connection: close");
    client.println();
    client.println();
    delay(200);
    while (client.available())
    {
      char c = client.read();
      if (c=='1'){
        buff=1;
      }
      if (c=='0'){
        buff=0;
      }
    }
    client.stop();
    client.flush();
  }
  else {
    client.stop();
    delay(1000);
    client.connect(serverName, 8080);
  }
  //work with relay
  if ( buff==1)
  {
    digitalWrite(relay, LOW);
    relay1State = "On";

    delay(7000);
    digitalWrite(relay, HIGH);
    relay1State = "Off";
  }
  else
  {
    digitalWrite(relay, HIGH);
    relay1State = "Off";
  }
  delay(500);
  
  }
