package main // import "github.com/portainer/authenticator"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"github.com/portainer/authenticator/cli"
)

type authenticationRequestPayload struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

// TODO: find these in Docker and use from there...
type endpoints struct {
	Host          string
	SkipTLSVerify bool
}
type context struct {
	Name      string
	Metadata  map[string]string
	Endpoints map[string]endpoints
}

func main() {

	options := cli.ParseOptions()

	apiURL, err := url.Parse(*options.PortainerAPI)
	if err != nil {
		log.Fatalf("Invalid Portainer URL: %s", err.Error())
	}

	if options.Password == nil || *options.Password == "" {
		//prompt for password...

	}

	apiURL.Path = path.Join(apiURL.Path, "/api/auth")
	authenticationURL := apiURL.String()

	payload := authenticationRequestPayload{
		Username: *options.Username,
		Password: *options.Password,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Unable to encode payload: %s", err.Error())
	}

	response, err := http.Post(authenticationURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatalf("Unable to execute authentication request: %s", err.Error())
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnprocessableEntity {
			log.Fatalf("Invalid credentials: %s / %s", payload.Username, payload.Password)
		} else {
			log.Fatalf("An error occured during HTTP authentication")
		}
	}

	data, err := getResponseBodyAsJSONObject(response)
	if err != nil {
		log.Fatalf("Unable to read authentication response: %s", err.Error())
	}

	token := data["jwt"]

	raw, err := ioutil.ReadFile(*options.ConfigFilePath)
	if err != nil {
		log.Fatalf("Unable to read configuration file: %s", err.Error())
	}

	var fileData map[string]interface{}
	err = json.Unmarshal(raw, &fileData)
	if err != nil {
		log.Fatalf("Unable to decode configuration file: %s", err.Error())
	}

	if fileData["HttpHeaders"] == nil {
		fileData["HttpHeaders"] = make(map[string]interface{})
	}

	//TODO: this presumes there's only one portainer we're talking to :()

	headersObject := fileData["HttpHeaders"].(map[string]interface{})
	authorizationHeaderValue := "Bearer " + token.(string)
	headersObject["Authorization"] = authorizationHeaderValue

	buf, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		log.Fatalf("Unable to encode configuration file content: %s", err.Error())
	}

	err = ioutil.WriteFile(*options.ConfigFilePath, buf, 0644)
	if err != nil {
		log.Fatalf("Unable to write to configuration file: %s", err.Error())
	}

	if *options.AddContexts {
		// https://app.swaggerhub.com/apis/deviantony/Portainer/1.23.2/#/endpoints/EndpointList
		// get all the /endpoints, and create / update them as contexts
		apiURL, _ := url.Parse(*options.PortainerAPI)
		endpoints, err := getEndpoints(apiURL, authorizationHeaderValue)
		if err != nil {
			log.Fatalf("Unable to get endpoints list: %s", err.Error())
		}
		for i, v := range endpoints {
			//fmt.Printf("%v: %v\n", i, v)
			data := v.(map[string]interface{})
			fmt.Printf("%d\n", i)
			// for k, vv := range data {
			// 	fmt.Printf("  %s: %s\n", k, vv)
			// }
			// TODO: grab the struct from api/portainer.go
			Name := data["Name"].(string)
			fmt.Printf("  %s: %s\n", "Name", Name)
			Type := int(data["Type"].(float64))
			fmt.Printf("  %s: %d\n", "Type", Type)
			EndpointID := int(data["Id"].(float64))
			fmt.Printf("  %s: %d\n", "EndpointID", EndpointID)
			// TODO: TLS config...
			// TODO: can we use edgeid and others to identify it if the name changes..
			// Type 1, 2, 4 are Docker (3 is ACI, and IDK if i can Docker API that...)
			if Type == 1 || Type == 2 || Type == 4 {
				// TODO: see if the context is already defined, and update it
				var stderr bytes.Buffer
				SimplifiedName := strings.ReplaceAll(Name, " ", "-") // TODO: this needs to be regex based: names are validated against regexp "^[a-zA-Z0-9][a-zA-Z0-9_.+-]+$"
				dockerURL, _ := url.Parse(*options.PortainerAPI)
				dockerURL.Scheme = "tcp"
				dockerURL.Path = fmt.Sprintf("/api/endpoints/%d/docker", EndpointID)

				cmd := exec.Command("docker", "context", "create", SimplifiedName, "--description", "portainer "+*options.PortainerAPI, "--docker", "host="+dockerURL.String())

				cmd.Stderr = &stderr

				output, err := cmd.Output()
				if err != nil {
					//return errors.New(stderr.String())
					fmt.Printf("Unable to create endpoint: %s: %s", err, stderr.String())
				}
				fmt.Printf("  %s\n", output)
			} else {
				fmt.Printf("  Skipping, not Docker\n")
			}
		}
		// remember that "local" is special, and is likely not /var/lib/docker.sock - its local to portainer.
		// also, the user might do this to several portainer instances
		// fear Ip addresses, they may change - use context metadata to match on other info
	}
}

func getEndpoints(apiURL *url.URL, authorizationHeaderValue string) (endpointArray []interface{}, err error) {
	apiURL.Path = path.Join(apiURL.Path, "/api/endpoints")
	endpointsURL := apiURL.String()

	fmt.Printf("Requesting: %v\n", endpointsURL)

	//response, err := http.Get(endpointsURL)
	req, _ := http.NewRequest("GET", endpointsURL, nil)
	req.Header.Set("Authorization", authorizationHeaderValue)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to execute authentication request: %s", err.Error())
	}

	// if response.StatusCode != http.StatusOK {
	// 	if response.StatusCode == http.StatusUnprocessableEntity {
	// 		log.Fatalf("Invalid credentials: %s / %s", payload.Username, payload.Password)
	// 	} else {
	// 		log.Fatalf("An error occured during HTTP authentication")
	// 	}
	// }

	fmt.Printf("response length: %v\n", response.ContentLength)
	//body, err := ioutil.ReadAll(response.Body)
	//fmt.Printf("response: %v\n", string(body))

	err = json.NewDecoder(response.Body).Decode(&endpointArray)
	return endpointArray, err
}

func getResponseBodyAsJSONObject(response *http.Response) (map[string]interface{}, error) {
	var data map[string]interface{}

	err := json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
