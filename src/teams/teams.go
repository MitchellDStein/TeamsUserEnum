package teams

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	color "github.com/fatih/color"
)

var URL_PRESENCE_TEAMS = "https://presence.teams.microsoft.com/v1/presence/getpresence/"
var URL_TEAMS = "https://teams.microsoft.com/api/mt/emea/beta/users/%s/externalsearchv3"
var CLIENT_VERSION = "27/1.0.0.2021011237"

var output = os.Stdout

// Enumuser request the Teams API to retrieve information about the email
func Enumuser(email string, bearer string, verbose bool) error {

	url := fmt.Sprintf(URL_TEAMS, email)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)
	req.Header.Add("x-ms-client-version", CLIENT_VERSION)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonInterface interface{}
	var usefulInformation []struct {
		DisplayName       string `json:"displayName"`
		Mri               string `json:"mri"`
		UserPrincipleName string `json:"userPrincipalName"`
		GivenName         string `json:"givenName"`
	}

	json.Unmarshal([]byte(body), &jsonInterface)
	json.Unmarshal([]byte(body), &usefulInformation)

	if verbose {

		fmt.Println("Email: " + email)
		fmt.Println("Status code: " + strconv.Itoa(resp.StatusCode))
		fmt.Println("Response: ")

		bytes, _ := json.MarshalIndent(jsonInterface, "", " ")
		fmt.Println(string(bytes))
	}

	if resp.StatusCode == 200 {
		if reflect.ValueOf(jsonInterface).Len() > 0 && usefulInformation[0].DisplayName != usefulInformation[0].GivenName {
			if reflect.ValueOf(jsonInterface).Len() > 0 {
				presence, device := getPresence(usefulInformation[0].Mri, bearer, verbose)
				color.Green("[+] " + email + " - " + usefulInformation[0].DisplayName + " - " + presence + " - " + device)
				fmt.Fprintln(output, email)
			} else {
				fmt.Println("[-] " + email)
			}

		} else {
			fmt.Println("[-] " + email)
		}
	} else if resp.StatusCode == 403 {
		color.Green("[+] " + email)
		// if output is specified, write the email to the file
		fmt.Fprintln(output, email)
		

	} else if resp.StatusCode == 401 {
		fmt.Println("[-] " + email)
		fmt.Println("The token may be invalid or expired. The status code returned by the server is 401")
		return errors.New(fmt.Sprint(resp.StatusCode))
	} else {
		fmt.Println("[-] " + email)
		fmt.Println("Something went wrong. The status code returned by the server is " + strconv.Itoa(resp.StatusCode))
		return errors.New(fmt.Sprint(resp.StatusCode))
	}

	return nil
}

// Parsefile will call the function Enumuser with the line as email's argument
func Parsefile(filenPath string, bearer string, verbose bool, threads int, outputFile string) {
	var err error

	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			log.Printf("Can't open %v: %v\n\n", outputFile, err)
		}
		defer output.Close()
	}

	input := os.Stdin
	if filenPath != "" {
		input, err = os.Open(filenPath)
		if err != nil {
			log.Fatalf("Can't open %v: %v", filenPath, err)
		}
		defer input.Close()
	}

	emails := bufio.NewScanner(input)
	emails.Split(bufio.ScanLines)

	inputBuffer := make(chan string, 256)

	var jobs sync.WaitGroup

	jobs.Add(threads) // threads workers
	for i := 0; i < threads; i++ {
		go func() {
			for email := range inputBuffer {
				Enumuser(email, bearer, verbose)
			}
			jobs.Done()
		}()
	}

	// Start reading users from file and put them in inputBuffer
	go func() {
		var line int
		for emails.Scan() {

			user := emails.Text()
			if user != "" {
				if strings.ContainsAny(user, `"/\:;|=,+*?<>`) {
					continue // skip invalid characters
				}
				inputBuffer <- user
			}
			line++
		}
		close(inputBuffer)
	}()

	jobs.Wait()
}

// getPresence request the Teams API to get additional details about the user with its mri
func getPresence(mri string, bearer string, verbose bool) (string, string) {

	var json_data = []byte(`[{"mri":"` + mri + `"}]`)
	req, _ := http.NewRequest("POST", URL_PRESENCE_TEAMS, bytes.NewBuffer(json_data))
	req.Header.Add("Authorization", bearer)
	req.Header.Add("x-ms-client-version", CLIENT_VERSION)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var status []struct {
		Mri      string `json:"mri"`
		Presence struct {
			Availability string `json:"availability"`
			DeviceType   string `json:"deviceType"`
		} `json:"presence"`
	}

	json.Unmarshal([]byte(body), &status)

	// if staus is empty, the user is offline
	if len(status) == 0 {
		return "error", "error"
	}
	return status[0].Presence.Availability, status[0].Presence.DeviceType
}
