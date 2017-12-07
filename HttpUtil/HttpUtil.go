//HttpUtil contains functions which relate to the http.Client struct.
//Author: Edwin Heerschap
package HttpUtil

import(
  "net"
  "net/http"
  "net/url"
  "io/ioutil"
  "log"
  "golang.org/x/net/proxy"
  "time"
  "fmt"
)

//SetupProxyClient creates a new Http.Client object with a modified transport struct
//to use the passed proxy. Proxy is a direct proxy.
func SetupProxyClient(proxyUrl string, timeout int) *http.Client {

  // Parse proxy URL string to a URL type
  proxyUrlObj, err := url.Parse(proxyUrl)
  if err != nil {
    log.Fatal("Error parsing Tor Proxy URL:", proxyUrl, ".", err)
  }

  // Create a dialer for the Transport struct
  proxyDialer, err := proxy.FromURL(proxyUrlObj, proxy.Direct)
  if err != nil {
    log.Fatal("Error setting Tor proxy.", err)
  }

  // Set up a custom HTTP transport to use the proxy and create the client
  transport := &http.Transport{Dial: proxyDialer.Dial}
  client := &http.Client{Transport: transport, Timeout: time.Second * 60}

  return client
}

//Should only be called if using the Tor service.
//This will cause the Tor service to reset its routing.
//
//*IMPORTANT* This has been designed to work on Ubuntu 16.04 running the
//Tor service with the controlport opened and config file set correctly.
//*NOTE* http.Client services must be recreated to use new Tor circuit after
//this function is called.
func SwitchTorEndpoint(torControl string, torPass string){

  //Creating connection to Tor Control port
  conn, err := net.Dial("tcp", torControl)

  if err != nil {
    log.Fatal("Failed to switch Tor endpoint")
  }

  defer conn.Close()

  //Authenticating with Tor Control TCP server.
  conn.Write([]byte("authenticate \"" + torPass + "\"\n"))
  response := make([]byte, 1024)
  conn.Read(response)

  //Requesting new Tor circuit
  conn.Write([]byte("signal newnym\n"))
  response = make([]byte, 1024)
  conn.Read(response)

  //Closing connection
  conn.Write([]byte("quit\n"))
  response = make([]byte, 1024)
  conn.Read(response)

  client := SetupProxyClient("socks5://127.0.0.1:9050", 60)
  resp, success, httpcode := SendHttpRequest(client, "http://checkip.amazonaws.com/")
      if !success || ( httpcode != http.StatusOK ){
        fmt.Println("Failed to check ip")
      }

  fmt.Println(string(resp))

}

//SendHttpRequest sends an http request using the passed http client, url.
//The response is read into a byte array using the ioutil.
//Error handling is performed and logged to the console.
//
//Returns a byte array, local success and http status
func SendHttpRequest(c *http.Client, url string) ([]byte, bool, int) {

  resp, err := c.Get(url)

  //Ensuring body closes
  defer resp.Body.Close()

  if err != nil {
    log.Fatal("Failed to send http request request: ", err)
    return nil, false, 0
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal("Failed to read body through ioutil")
    return nil, false, resp.StatusCode
  }

  return respBody, true, resp.StatusCode

}
