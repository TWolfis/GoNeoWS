package GoNeoWS

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// URL points to the API endpoint
const URL = "https://api.nasa.gov/neo/rest/v1/feed?"

type GoNeoWS struct {
	StartDate       string
	EndDate         string
	APIKey          string
	Query           string
	Request         *http.Request
	GoNeoWSResponse []GoNeoWSResponse
}

// GoNeoWSResponse holds the response received from the call to the NeoWS API
// Struct generated through the use of https://mholt.github.io/json-to-go/
type GoNeoWSResponse struct {
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	ID                 string  `json:"id"`
	NeoReferenceID     string  `json:"neo_reference_id"`
	Name               string  `json:"name"`
	NasaJplURL         string  `json:"nasa_jpl_url"`
	AbsoluteMagnitudeH float64 `json:"absolute_magnitude_h"`
	EstimatedDiameter  struct {
		Kilometers struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"kilometers"`
		Meters struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"meters"`
		Miles struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"miles"`
		Feet struct {
			EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
			EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
		} `json:"feet"`
	} `json:"estimated_diameter"`
	IsPotentiallyHazardousAsteroid bool `json:"is_potentially_hazardous_asteroid"`
	CloseApproachData              []struct {
		CloseApproachDate      string `json:"close_approach_date"`
		CloseApproachDateFull  string `json:"close_approach_date_full"`
		EpochDateCloseApproach int64  `json:"epoch_date_close_approach"`
		RelativeVelocity       struct {
			KilometersPerSecond string `json:"kilometers_per_second"`
			KilometersPerHour   string `json:"kilometers_per_hour"`
			MilesPerHour        string `json:"miles_per_hour"`
		} `json:"relative_velocity"`
		MissDistance struct {
			Astronomical string `json:"astronomical"`
			Lunar        string `json:"lunar"`
			Kilometers   string `json:"kilometers"`
			Miles        string `json:"miles"`
		} `json:"miss_distance"`
		OrbitingBody string `json:"orbiting_body"`
	} `json:"close_approach_data"`
	IsSentryObject bool `json:"is_sentry_object"`
}

// composeRequest() serves to create the query to be sent to the API endpoint
func (gs *GoNeoWS) composeRequest() {

	//create request to be sent to endpoint
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln("[Error] Failed to create a new request", err)
	}

	// query holds the options added to the GET request
	query := req.URL.Query()

	// add API key
	if gs.APIKey != "" {
		query.Add("api_key", gs.APIKey)
	} else {
		log.Println("[Informative] APIKey not set using DEMO_KEY")
		query.Add("api_key", "DEMO_KEY")
	}

	// verify StartDate
	_, err = time.Parse("2006-01-02", gs.StartDate)
	if err != nil {
		log.Fatalln("[Error] Failed to parse StartDate", err)
	}

	//add start date to query
	query.Add("start_date", gs.StartDate)

	// verify EndDate
	_, err = time.Parse("2006-01-02", gs.EndDate)
	if err != nil {

		// A Failure of parsing the EndDate is not a problem since by default the EndDate is set to 7 days from the StartDate
		log.Println("[Informative] Error when parsing EndDate", err)
		log.Println("[Informative] EndDate has not been set so this will default to 7 days from the Start date")
	} else {
		query.Add("end_date", gs.EndDate)
	}

	//encode request including the added options
	req.URL.RawQuery = query.Encode()

	gs.Request = req
	gs.Query = req.URL.String()

}

func (gs *GoNeoWS) MakeRequest(debug bool) {
	// compose request to the NeoWS endpoint
	gs.composeRequest()

	if debug {
		log.Println("[Informative] Making request to", gs.Query)
	}
	//create client and make request to the NeoWS endpoint
	resp, err := http.Get(gs.Query)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	reader, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[Error] while trying to read the response", err)
	}

	// unwrap response
	gs.unwrap(reader)
}

func (gs *GoNeoWS) unwrap(body []byte) {

	// unwrap initial JSON response to a temporary variable
	// temp is a mapping of json.RawMessage from where we want to extract the data stored in the key "near_earth_objects"
	var temp map[string]json.RawMessage
	err := json.Unmarshal(body, &temp)
	if err != nil {
		log.Fatal("[Error] failed to parse received JSON document", err)
	}

	// unwrap temp to a mapping of json.RawMessage to hold the "near_earth_objects" data
	var neo map[string]json.RawMessage
	err = json.Unmarshal(temp["near_earth_objects"], &neo)
	if err != nil {
		log.Fatal("[Error] failed to unwrap temp mapping to NEO map", err)
	}

	// unwrap neo map to responses array of structs
	for key := range neo {
		err = json.Unmarshal(neo[key], &gs.GoNeoWSResponse)
		if err != nil {
			log.Fatal("[Error] failed to unwrap Neo mapping", err)
		}
	}

}
