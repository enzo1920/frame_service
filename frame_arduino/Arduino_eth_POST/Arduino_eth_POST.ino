#include <ArduinoJson.h>
#include <UIPEthernet.h> // Used for Ethernet
#include <OneWire.h>
#include <SPI.h>


byte mac[] = { 0x00, 0xAA, 0xBB, 0xCC, 0xDE, 0x01 }; // RESERVED MAC ADDRESS
EthernetClient client;

OneWire ds(7); // на пине 7 (нужен резистор 2.2 КОм)


char server[] = "cloud.framecase.ru";
// **** ETHERNET SETTING ****
//byte mac[] = { 0x90, 0xA2, 0xDA, 0x0D, 0x78, 0xEE  };                                       
IPAddress ip(10, 10, 10, 244); 
IPAddress mydns(10,10,10,1);
IPAddress mygw(10,10,10,1);
IPAddress subnet(255,255,255,0);


//String PostData = "";
char  post[] = "POST /v1/upload/temp/?token=5d6f3ecb1cb3d69b HTTP/1.1";

void setup() { 
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
  
  //delay(1000); // GIVE THE SENSOR SOME TIME TO START

 
  //str_temp =  DS18S20_read_temp(); 
  //Serial.println(str_temp);

  //data = "";
}

void loop(){

  /*currentMillis = millis();
  if(currentMillis - previousMillis > interval) { // READ ONLY ONCE PER INTERVAL
    previousMillis = currentMillis;
    str_temp =  DS18S20_read_temp(); 
  }*/
  sendPOST();
  delay(20000); // 50 sec WAIT   BEFORE SENDING AGAIN
}


void sendPOST() //client function to send/receive GET request data.
{
   //PostData = "";
   String PostData=  DS18S20_read_temp();
   //Serial.println(PostData);

   if(PostData.length()>0){

       if (client.connect(server,80)) {           
          //Serial.println("connected");
          //Serial.println("=================>");
          client.println(post);
          client.println("Host: cloud.framecase.ru");
          client.println("User-Agent: Arduino/1.0");
          //client.println("Connection: close");
          client.println("Content-Type: text/plain");
          client.print("Content-Length: ");
          client.println(PostData.length());
          client.println();
          client.println(PostData);
          //Serial.println("disconnecting.");
          client.stop(); //stop client
    
          
          } 
       //else {
        //  Serial.println("connection failed"); //error message if no client connect
          //Serial.println();
      // }
        //delay(100);
   }

}

char*  DS18S20_read_temp(){
   byte i;
   byte present = 0;
   byte type_s;
   byte data[12];
   byte addr[8];
   char result[8];
   float celsius, fahrenheit;
   if (!ds.search(addr)) {
         //Serial.println("No more addresses.");
         ds.reset_search();
         delay(250);
         return;
    }

   if (OneWire::crc8(addr, 7) != addr[7]) {
        //Serial.println("CRC is not valid!");
        return;
   }
   // первый байт определяет чип
   switch (addr[0]) {
          case 0x10:
                   //Serial.println("Chip = DS18S20"); // или более старый DS1820
                   type_s = 1;
                   break;
          case 0x28:
                   //Serial.println("Chip = DS18B20");
                   type_s = 0;
                   break;
          case 0x22:
                   //Serial.println("Chip = DS1822");
                   type_s = 0;
                   break;
          default:
                   //Serial.println("Device is not a DS18x20 family device.");
                   return;
          }
    ds.reset();
    ds.select(addr);
    ds.write(0x44); // начинаем преобразование, используя ds.write(0x44,1) с "паразитным" питанием
    delay(2000); // 750 может быть достаточно, а может быть и не хватит
    // мы могли бы использовать тут ds.depower(), но reset позаботится об этом
    present = ds.reset();
    ds.select(addr);
    ds.write(0xBE);

    for ( i = 0; i < 9; i++) { // нам необходимо 9 байт
          data[i] = ds.read();
          //Serial.print(data[i], HEX);
          //Serial.print(" ");
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
            // при маленьких значениях, малые биты не определены, давайте их обнулим
            if (cfg == 0x00) raw = raw & ~7; // разрешение 9 бит, 93.75 мс
            else if (cfg == 0x20) raw = raw & ~3; // разрешение 10 бит, 187.5 мс
            else if (cfg == 0x40) raw = raw & ~1; // разрешение 11 бит, 375 мс
            //// разрешение по умолчанию равно 12 бит, время преобразования - 750 мс
      }
      celsius = (float)raw / 16.0;
      dtostrf(celsius, 6, 2, result);
      //cels_str = String(celsius,2);
      return result;
}


