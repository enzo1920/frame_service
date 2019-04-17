#include <UIPEthernet.h> // Used for Ethernet
#include <OneWire.h>
#include <SPI.h>



byte mac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xEE, 0x01 }

EthernetClient client;
char server[] = "framecase.tula.su"; // имя вашего сервера
int buff=0;

// Relay state and pin
String relay1State = "Off";
const int relay = 7;


IPAddress ip(10, 10, 10, 244); 
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

	sensors.requestTemperatures();

	if (client.connect(server, 8080)) 
	{
		client.print("GET /v1/get/relay/reset/?token=5d6f3ecb1cb3d69b HTTP/1.1");
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
		delay(100); 
	}
	else {
		client.stop();
		delay(1000);
		client.connect(server, 8080);
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