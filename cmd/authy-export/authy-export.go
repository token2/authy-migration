package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
 	"encoding/base64"
	
	"github.com/token2/authy-migration"
	"golang.org/x/crypto/ssh/terminal"
)

import qrcode "github.com/skip2/go-qrcode"

// We'll persist this to the filesystem so we don't need to
// re-register the device every time
type deviceRegistration struct {
	UserID   uint64 `json:"user_id,omitempty"`
	DeviceID uint64 `json:"device_id,omitempty"`
	Seed     string `json:"seed,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// Struct for Raivo JSON format
type raivo struct {
        Digits  string `json:"digits"`
        Kind    string `json:"kind"`
        Algo    string `json:"algorithm"`
        Counter string `json:"counter"`
        Timer   string `json:"timer"`
        Secret  string `json:"secret"`
        Account string `json:"account"`
        Issuer  string `json:"issuer"`
}


func lineCounter(fileName string) int {
    f, _ := os.Open(fileName)
    // Create new Scanner.
    scanner := bufio.NewScanner(f)
    result := 0
    // Use Scan.
    for scanner.Scan() {
        //line := scanner.Text()
        // Append line to result.
        result = result + 1
    }
    return result
}

func main() {
	var filename string
	var name1 string
	var line int
	var err error
	var raivos []raivo

	sc := bufio.NewScanner(os.Stdin)

	fmt.Print("\nExport file name: end with .txt for Molto2, with wa.txt for WinAuth import file, .raivos for Raivo, and with .html for regular TOTP profiles: ")
	if !sc.Scan() {
		fmt.Print("A filename is required")
	}
	filename = strings.TrimSpace(sc.Text())

	length := len(filename)
	if length < 4 {
		log.Fatalf("Filename %s too short; did you include the extension?", filename)
	}
	last4 :=  filename[length - 4:length]
	last6 := ""
	if length > 5 {
		last6 =  filename[length - 6:length]
	}

	fmt.Print("File: "+ filename)

	// If we don't already have a registered device, prompt the user for one
	regr, err := loadExistingDeviceRegistration()
	if err == nil {
		log.Println("Found existing device registration")
	} else if os.IsNotExist(err) {
		log.Println("No existing device registration found, will perform registration now")
		regr, err = newInteractiveDeviceRegistration()
		if err != nil {
			log.Fatalf("Device registration failed: %v", err)
		}
	} else if err != nil {
		log.Fatalf("Could not check for existing device registration: %v", err)
	}

	// By now we have a valid user and device ID
	log.Printf("Authy User ID %d, Device ID %d", regr.UserID, regr.DeviceID)

	cl, err := authy.NewClient()
	if err != nil {
		log.Fatalf("Couldn't create API client: %v", err)
	}

	// Fetch the apps
	appsResponse, err := cl.QueryAuthenticatorApps(nil, regr.UserID, regr.DeviceID, regr.Seed)
	if err != nil {
		log.Fatalf("Could not fetch authenticator apps: %v", err)
	}
	if !appsResponse.Success {
		log.Fatalf("Failed to fetch authenticator apps: %+v", appsResponse)
	}

	// Fetch the actual tokens now
	tokensResponse, err := cl.QueryAuthenticatorTokens(nil, regr.UserID, regr.DeviceID, regr.Seed)
	if err != nil {
		log.Fatalf("Could not fetch authenticator tokens: %v", err)
	}
	if !tokensResponse.Success {
		log.Fatalf("Failed to fetch authenticator tokens: %+v", tokensResponse)
	}

	// We'll need the prompt the user to give the decryption password
	pp := []byte(os.Getenv("AUTHY_EXPORT_PASSWORD"))
	if len(pp) == 0 {
		log.Printf("Please provide your Authy TOTP backup password: ")
		pp, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Failed to read the password: %v", err)
		}
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	if (last4 == "html") {
		_, err = f.WriteString(`<!DOCTYPE html>
			<html>
			<head>
			    <title>2FA Codes</title>
			    <style>
				body > div {display: grid; grid-template-columns: repeat(auto-fill,minmax(256px, 1fr));}
				div > div {text-align: center; border: 1px dashed #ccc; padding: 5px; margin: 5px; overflow: hidden; text-overflow: ellipsis;}
				div > div p {height: 2rem;}
				img {filter: blur(6px);}
				div:hover > img {filter: none;}
				@media print {
					img {filter: none !important;}
					body > p {display: none;}
					div > div {word-wrap: break-word;}
				}
			    </style>
			</head>
			<body>
			    <p>
				Hover over the QR code to display it for capture, or double click on the text code to copy it to your clipboard.
			    </p>
			    <div>
		`)
		if err != nil {
			log.Fatalf("Can't write file %v", err)
		}
	} else if last4 != ".txt" && last6 != "raivos" && last6 != ".wa.txt" {
		log.Fatalf("Invalid filename %s must end with .html, .raivos, .wa.txt, or .txt", filename)
	}

	log.Println("TOTP profile migration file is being generated:\n")
	for _, tok := range tokensResponse.AuthenticatorTokens {
		decrypted, err := tok.Decrypt(string(pp))
		if err != nil {
			log.Printf("Failed to decrypt token %s: %v", tok.Description(), err)
			continue
		}

		params := url.Values{}
		params.Set("secret", decrypted)
		params.Set("digits", strconv.Itoa(tok.Digits))
		s := strings.Split(params.Encode(), "&")
		d := strings.Split(s[0],"=")

		if (last6 == "raivos") {
			raivos = append(raivos, raivo{
				Digits:  "6",
				Kind:    "TOTP",
				Algo:    "SHA1",
				Counter: "0",
				Timer:   "30",
				Secret:  decrypted,
				Account: "authy-export",
				Issuer:  tok.Description()})
		}

		if (last6 == "wa.txt") {
			u := url.URL{
				Scheme:   "otpauth",
				Host:     "totp",
				Path:     tok.Description(),
				RawQuery: params.Encode(),
			}

		  
			_, err = f.WriteString(u.String() + "\n")
			if err != nil {
				log.Printf("Error writing to file: %v", err)
			}
		}
		
		if (last4 == "html") {
			u := url.URL{
				Scheme:   "otpauth",
				Host:     "totp",
				Path:     tok.Description(),
				RawQuery: params.Encode(),
			}

			png, err := qrcode.Encode(u.String(), qrcode.Medium, 256)
			if err != nil {
				log.Printf("Failed to generate QR code: %v", err)
				continue
			}
			sEnc := base64.StdEncoding.EncodeToString([]byte(png))
			_, err = f.WriteString("<div><p>" + tok.Description() + "</p><img src='data:image/png;base64," + sEnc + "'><br/><kbd>" + decrypted + "</kbd></div>\n")
			if err != nil {
				log.Printf("Error writing to file: %v", err)
			}
		} else if (last4 == ".txt" && last6 != "wa.txt") {
			line = lineCounter(filename)
			line = line + 1
			name1 = tok.Description()
			name1 = strings.Replace(name1, ":", "_", 10)
			name1 = strings.Replace(name1, "@", "_", 10)
			if len(name1) > 12 {
				name1 = name1[0:12]
			}
			_, err = f.WriteString(strconv.Itoa(line - 1) + "                   " + decrypted + "                   sha1                   " + d[1] + "                   30                   yes                   yes                   "  + name1 + "\n")
			if err != nil {
				log.Println(err)
			}
		}
	}

	if (last4 == "html") {
		_, err = f.WriteString(`    </div>
			    <script>
				let kbds = document.getElementsByTagName("kbd");
				for (const kbd of kbds) {
				    kbd.addEventListener("dblclick", e => navigator.clipboard.writeText(e.target.textContent))
				}
			    </script>
			</body>
			</html>
		`)
		if err != nil {
			log.Println(err)
		}
	}

	if last6 == "raivos" {
		json, _ := json.MarshalIndent(raivos, "", "   ")
		_, err = f.WriteString(string(json))
	}


	defer f.Close()

	for _, app := range appsResponse.AuthenticatorApps {
		tok, err := app.Token()
		if err != nil {
			log.Printf("Failed to decode app %s: %v", app.Name, err)
			continue
		}
		params := url.Values{}
		params.Set("secret", tok)
		params.Set("digits", strconv.Itoa(app.Digits))
		params.Set("period", "10")
	}

	fmt.Print("\nPress 'Enter' to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func newInteractiveDeviceRegistration() (deviceRegistration, error) {
	var regr deviceRegistration
	// The first part of device registration requires the user's phone number that
	// is attached to their Authy account
	var phoneCC int
	var phoneNumber string

	var err error
	sc := bufio.NewScanner(os.Stdin)
	fmt.Print("\nWhat is your phone number's country code? (digits only): ")
	if !sc.Scan() {
		return regr, errors.New("Please provide a phone country code, e.g. 61")
	}
	if phoneCC, err = strconv.Atoi(strings.TrimSpace(sc.Text())); err != nil {
		return regr, err
	}
	fmt.Print("\nWhat is your phone number? (digits only): ")
	if !sc.Scan() {
		return regr, errors.New("Please provide a phone country code, e.g. 12341234")
	}
	phoneNumber = strings.TrimSpace(sc.Text())
	if err := sc.Err(); err != nil {
		return regr, err
	}

	// Query the existence of the Authy account
	cl, err := authy.NewClient()
	if err != nil {
		return regr, nil
	}
	userStatus, err := cl.QueryUser(nil, phoneCC, phoneNumber)
	if err != nil {
		return regr, err
	}
	if !userStatus.IsActiveUser() {
		return regr, errors.New("There doesn't seem to be an Authy account attached to that phone number")
	}

	// Begin a device registration using Authy app push notification
	regStart, err := cl.RequestDeviceRegistration(nil, userStatus.AuthyID, authy.ViaMethodPush)
	if err != nil {
		return regr, err
	}
	if !regStart.Success {
		return regr, fmt.Errorf("Authy did not accept the device registration request: %+v", regStart)
	}

	// Poll for a while until the user has responded to the device registration
	var regPIN string
	timeout := time.Now().Add(5 * time.Minute)
	for {
		if timeout.Before(time.Now()) {
			return regr, errors.New("Gave up waiting for user to respond to Authy device registration request")
		}

		log.Printf("Checking device registration status (%s until we give up)", time.Until(timeout).Truncate(time.Second))

		regStatus, err := cl.CheckDeviceRegistration(nil, userStatus.AuthyID, regStart.RequestID)
		if err != nil {
			return regr, err
		}
		if regStatus.Status == "accepted" {
			regPIN = regStatus.PIN
			break
		} else if regStatus.Status != "pending" {
			return regr, fmt.Errorf("Invalid status while waiting for device registration: %s", regStatus.Status)
		}

		time.Sleep(5 * time.Second)
	}

	// We have the registration PIN, complete the registration
	regComplete, err := cl.CompleteDeviceRegistration(nil, userStatus.AuthyID, regPIN)
	if err != nil {
		return regr, err
	}
	if regComplete.Device.SecretSeed == "" {
		return regr, errors.New("Something went wrong completing the device registration")
	}

	regr.UserID = regComplete.AuthyID
	regr.DeviceID = regComplete.Device.ID
	regr.Seed = regComplete.Device.SecretSeed
	regr.APIKey = regComplete.Device.APIKey

	// We have everything we need, persist it to disk
	regrPath, err := configPath()
	if err != nil {
		return regr, err
	}
	f, err := os.OpenFile(regrPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return regr, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(regr); err != nil {
		return regr, err
	}

	return regr, nil
}

func loadExistingDeviceRegistration() (deviceRegistration, error) {
	regrPath, err := configPath()
	if err != nil {
		return deviceRegistration{}, err
	}

	f, err := os.Open(regrPath)
	if err != nil {
		return deviceRegistration{}, err
	}
	defer f.Close()

	var regr deviceRegistration
	return regr, json.NewDecoder(f).Decode(&regr)
}

func configPath() (string, error) {
	// TODO: In Go 1.13, use os.UserConfigDir()
	regrPath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(regrPath, "authy-go.json"), nil
}
