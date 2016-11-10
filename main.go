package main

import (
    "net"
    "net/http"
    "os"
    "fmt"
    "io/ioutil"
    "time"
    "strings"
    "strconv"

    // Used for web sockets
    "github.com/gorilla/websocket"

    // used for GPIO
    "github.com/kidoman/embd"
    _ "github.com/kidoman/embd/host/rpi"

)

// Package variables.  Would remove these if this was
// more than just a PoC (ie if I really had time to code this)
var powerPin embd.DigitalPin
var monitorPin embd.DigitalPin
var globalChannel chan int 
var upgrader websocket.Upgrader
var counter int = 0
var running bool = false
var registered bool = false
var firstcallback bool = true
var gpioStatus int 
var debug bool = true// this is a PoC...

const(
    TAGNAME = "TOTALUNITS"
    DEBUG = "DEBUG"
)

func startFoam(conn *websocket.Conn) {

    // Now that we are on set the status to running
    fmt.Println("Foam started" )
    running = true

    // if not running in debug mode, turn on the power
    // and set up the call back
    if (!debug) {
      // Try to open GPIO 017 as the power pin
      var err error
      powerPin, err = embd.NewDigitalPin(17)
      if err != nil {
        fmt.Println( "Error opening pin %s\n", err)
        return
      }
      powerPin.SetDirection(embd.Out)
      monitorPin, err = embd.NewDigitalPin(23)
      if err != nil {
        fmt.Println( "Error opening pin %s\n", err)
        return
      }
        monitorPin.SetDirection(embd.In)
        err = powerPin.Write(embd.High)
        if err != nil {
            fmt.Println(err)
            return
       }

       if !registered {
           globalChannel = make( chan int )
           go monitorFlow(globalChannel, conn)
           registered = true
       }
//       embd.InitGPIO()
//       powerPin.SetDirection(embd.Out)
       monitorPin.SetDirection(embd.In)
       monitorPin.ActiveLow(false)
       monitorPin.PullUp()
       err = monitorPin.Watch(embd.EdgeFalling, gpioCallback)
       if err != nil {
           fmt.Println(err)
           return
       }

    }

}

func gpioCallback( pin embd.DigitalPin) {

//    if firstcallback {
//        firstcallback = false
//        
//        fmt.Println( "First callback" )
//        return
//    }

//    fmt.Println( "Callback for %d", counter )
    counter += 1
//    globalChannel <-counter
}

func stopFoam() {

    // if not running in debug mode, turn on the power
    if (!debug) {
        err := powerPin.Write(embd.Low)
        if err != nil {
            fmt.Println(err)
            return
       }
        monitorPin.StopWatching()
    }
    embd.InitGPIO()

    // set the running state to false.  this will shut everything down
    running = false
    fmt.Println("stopped running")
}

func monitorFlow( c chan int, conn *websocket.Conn ) {

    // Create a channel with a buffer of 100 messages for calling TS
    tsC := make(chan float64, 100)
    go callTS(tsC)

    meterStart := time.Now()
    var meterEnd time.Time
    // 2000 pulses per minute / 60 seconds
    var targetFlow float64 = 2000 / 60
    var count float64
    var duration float64
    var msg float64
    for (running ) {

	// if in debug mode, sleep every second then trigger a callback
        if debug {
           time.Sleep(1 * time.Second)
           counter += 1
           msg = float64(counter)
//           err := conn.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(msg)))
           err := conn.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(counter)))
           if err != nil {
             fmt.Println(err)
             stopFoam()
             break
           }

        }  else {

            // figure out how many pulses since the last time we ran
            count = float64(counter)
            msg = 0
            counter = 0
            meterEnd = time.Now()
            duration = float64(meterEnd.Unix() - meterStart.Unix())
 
fmt.Println( "Duration = ", duration)           
            // prep for next time
            meterStart = meterEnd

            // flow rate in gallons per second has to be
            // pulses / number of seconds / targetFlow
            
            if count == 0 {
               fmt.Println("count is zero")
            } else { 
               msg = count / duration / targetFlow
            }
            time.Sleep(time.Second)
      }
 
      // Now send it to time series
      tsC <- float64(msg)
            
    }

}

func callTS( c chan float64 ) {

    for (running || len(c) > 0 ) {

      msg := <- c
fmt.Println("Sending to time series", msg )
      sendDataPointToPTS( TAGNAME, msg )

    }
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        fmt.Println(err)
        // the better way to do it is:
        // log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().String()
    idx := strings.LastIndex(localAddr, ":")

    return localAddr[0:idx]
}

func readHTML() (string) {
    // Read the HTML file from disk and return as string
    indexFile, err := os.Open("html/index.html")
    if err != nil {
        fmt.Println(err)
    }

   index, err := ioutil.ReadAll(indexFile)
   if err != nil {
      fmt.Println(err)
   }

  // Add the local IP address to the WSURi in the HMTL page and return it.
  // Again, this is a PoC, so there is no URL
  // The index[:] is used to convert the byte array to a string for Sprintf
  wsHost := GetOutboundIP()
  fmt.Println("Binding to " + wsHost)
  return string(fmt.Sprintf(string(index[:]), wsHost))
}

// Handle all of the web socket code
func handleWebSocket(w http.ResponseWriter, r *http.Request) {

    fmt.Println("Entered websocket handler")
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
      fmt.Println(err)
      return
    }

    for {
      msgType, msg, err := conn.ReadMessage()
      if err != nil || msgType != websocket.TextMessage{
        fmt.Println(err)
        return
      }

    fmt.Println("recieved " + string(msg))

    switch {
      case string(msg) == "start" : startFoam(conn)
      case string(msg) == "stop" : stopFoam()
      default : fmt.Printf("Error: received " + string(msg))
    }
  }

}

func main() {

    // Initialization - not running and set up the web socket
    running = false

    // Check to see if in debug mode or not.  Assuming true for PoC
    // TODO move a lot of this to debug mode - non debug is for GPIO
    env_debug := os.Getenv( DEBUG )
    if ( len( env_debug ) == 0 || env_debug == "FALSE" ) {
        debug = false
    } else {
      fmt.Println( "DEBUG RUN")
    }
   
    // in case we have trouble finding the IP address
    fmt.Println("using IP address " + GetOutboundIP() )

    upgrader = websocket.Upgrader{
      CheckOrigin: func(r *http.Request) bool { return true },
      ReadBufferSize: 1024,
      WriteBufferSize: 1024,
    }

    // Handlers for URLs - the main page and the web socket client
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, readHTML())
    })
    http.HandleFunc("/websocket", handleWebSocket)


    // Start the web server
    http.ListenAndServe(":3000", nil)
}
