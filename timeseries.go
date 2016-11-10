//*****************************
//
// Utility code to help with writing messages to Predix Time series
// Leverages a package from Alteros to do the actual time series writes
//*****************************

package main

import (

    "os"
    "fmt"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/clientcredentials"
    "github.com/Altoros/go-predix-timeseries/dataquality"
    "github.com/Altoros/go-predix-timeseries/measurement"
    "github.com/Altoros/go-predix-timeseries/api"
)

//TODO refactor this to camelCase
type EnvVariables struct {
  ts_uaa_host string
  ts_uaa_token_url string
  ts_uaa_client_id string
  ts_uaa_client_secret string
  ts_zone_id string
  ts_ingest string
}

const (
    TS_UAA_HOST ="TS_UAA_HOST"
    TS_UAA_TOKEN_URL="TS_UAA_TOKEN_URL"
    TS_UAA_CLIENT_ID="TS_UAA_CLIENT_ID"
    TS_UAA_CLIENT_SECRET="TS_UAA_CLIENT_SECRET"
    TS_ZONE_ID="TS_ZONE_ID"
    TS_INGEST="TS_INGEST"

    DEFAULT_ASSET_ID="ddf64089-713b-4510-9449-b8fe1db315f2"
)

// Create a placeholder for all of the environment variables
var env EnvVariables
var oauth2Token *oauth2.Token
var initialized bool = false

// Look up everthing that needs to be looked up
func doInit() {

    // Really boring, just read everything from OS
    env.ts_uaa_host = os.Getenv( TS_UAA_HOST)
    if ( len( env.ts_uaa_host) == 0 ) {
        fmt.Println( "Error reading " + TS_UAA_HOST)
    }

    env.ts_uaa_token_url = os.Getenv( TS_UAA_TOKEN_URL)
    if ( len( env.ts_uaa_token_url) == 0 ) {
        fmt.Println( "Error reading " + TS_UAA_TOKEN_URL)
    }

    env.ts_uaa_client_id = os.Getenv( TS_UAA_CLIENT_ID)
    if ( len( env.ts_uaa_client_id) == 0 ) {
        fmt.Println( "Error reading " + TS_UAA_CLIENT_ID)
    }

    env.ts_uaa_client_secret = os.Getenv( TS_UAA_CLIENT_SECRET)
    if ( len( env.ts_uaa_client_secret) == 0 ) {
        fmt.Println( "Error reading " + TS_UAA_CLIENT_SECRET)
    }

    env.ts_zone_id = os.Getenv( TS_ZONE_ID)
    if ( len( env.ts_zone_id) == 0 ) {
        fmt.Println( "Error reading " + TS_ZONE_ID)
    }

    env.ts_ingest = os.Getenv( TS_INGEST)
    if ( len( env.ts_ingest ) == 0 ) {
        fmt.Println( "Error reading " + TS_INGEST)
    }

    // If it all worked we are initialized
    initialized = true
}

func getAuthToken() string {

    if oauth2Token.Valid() {
      return oauth2Token.AccessToken
    }

    scopes := []string { "timeseries.zones." + env.ts_zone_id + ".user",
                         "timeseries.zones." + env.ts_zone_id + ".ingest",
                         "timeseries.zones." + env.ts_zone_id + ".query"}
    credentials := &clientcredentials.Config{
      ClientID: env.ts_uaa_client_id,
      ClientSecret: env.ts_uaa_client_secret,
      TokenURL: env.ts_uaa_host + env.ts_uaa_token_url,
      Scopes: scopes }

    var err error
    oauth2Token, err = credentials.Token(nil)
    if err != nil {
        fmt.Println( err )
        return ""
    }
    fmt.Println(oauth2Token.AccessToken)
    return oauth2Token.AccessToken

}

// Will add functions for the rest later if needed
// but the simplest help is just to send a data point
// assumes all of the
func sendDataPointToPTS( tagName string, data float64 ) {

    // Gotta be initializaed
    if !initialized {
      doInit()
    }

    // Send the message
    fmt.Println( fmt.Sprintf("Ready to send %f to time series as %s", data, DEFAULT_ASSET_ID + "." + tagName ))

    // TODO find a way not to reconnect every time.  May not be the generic
    // case but is a waste in this context.

    // The following is from https://github.com/Altoros/go-predix-timeseries/blob/master/doc.go
    // The following example shows a data ingestion request:
    //  api := api.Ingest("wss://ingest_url", accessToken, predixZoneId)
	  //  m := api.IngestMessage()
	  //  t, _ := m.AddTag("test_tag")
	  //  t.AddDatapoint(measurement.Int(123), dataquality.Good)
	  //  t.SetAttribute("key", "value")
	  //  m.Send()
    api := api.Ingest(env.ts_ingest, getAuthToken(), env.ts_zone_id)
    tsIngest := api.IngestMessage()
    tag, _ := tsIngest.AddTag(DEFAULT_ASSET_ID + "." + tagName)
    tag.AddDatapoint(measurement.Float(data), dataquality.Uncertain)
    tsIngest.Send()
}
