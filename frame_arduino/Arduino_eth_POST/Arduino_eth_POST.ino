//#include <ArduinoJson.h>
#include <UIPEthernet.h> // Used for Ethernet
#include <OneWire.h>
#include <SPI.h>


byte mac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xDE, 0x01 }; // RESERVED MAC ADDRESS
EthernetClient client;

OneWire ds(7); // на пине 7 (нужен резистор 2.2 КОм)

char  post[] = "POST /v1/upload/temp/?token=5d6f3ecb1cb3d69b HTTP/1.1";
char server[] = "cloud.framecase.ru";
// **** ETHERNET SETTING ****                                     
IPAddress ip(10, 10, 10, 244); 
IPAddress mydns(10,10,10,1);
IPAddress mygw(10,10,10,1);
IPAddress subnet(255,255,255,0);





void setup() { 
  Serial.begin(9600);
  Serial.println("\n[memCheck]");
  Serial.println(freeRam());


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
  delay(50000); // 50 sec WAIT   BEFORE SENDING AGAIN
  Serial.println("send post to cloud.framecase.ru "); 
  sendPOST();
  Serial.println("\n[memCheck2:]");
  Serial.println(freeRam());

}


void sendPOST() //client function to send/receive GET request data.
{

   char PostData[8];  
   DS18S20_read_temp(PostData);
   Serial.println(PostData);
   if (client.connect(server,80)) { 
       if(sizeof(PostData)>0){
          client.println(post);
          client.println("Host: cloud.framecase.ru");
          client.println("User-Agent: Arduino/1.0");
          client.println("Content-Type: text/plain");
          client.print("Content-Length: ");
          client.println(sizeof(PostData));
          client.println();
          client.println(PostData);
          client.stop(); //stop client
          } 
   }

}

char*  DS18S20_read_temp(char * result){
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
      return result;
}

int freeRam () {
  extern int __heap_start, *__brkval; 
  int v; 
  return (int) &v - (__brkval == 0 ? (int) &__heap_start : (int) __brkval); 
}


