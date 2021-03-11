package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/icedream/go-stagelinq"
)

const (
	appName    = "Icedream StagelinQ Receiver"
	appVersion = "0.0.0"
	timeout    = 5 * time.Second
)

var stateValues = []string{

	stagelinq.EngineDeck1PlayState,
	stagelinq.EngineDeck1CurrentBPM,
	stagelinq.EngineDeck1TrackArtistName,
	stagelinq.EngineDeck1TrackSongName,
	stagelinq.MixerCH1faderPosition,

	stagelinq.EngineDeck2PlayState,
	stagelinq.EngineDeck2CurrentBPM,
	stagelinq.EngineDeck2TrackArtistName,
	stagelinq.EngineDeck2TrackSongName,
	stagelinq.MixerCH2faderPosition,

	stagelinq.EngineDeck3PlayState,
	stagelinq.EngineDeck3CurrentBPM,
	stagelinq.EngineDeck3TrackArtistName,
	stagelinq.EngineDeck3TrackSongName,
	stagelinq.MixerCH3faderPosition,

	stagelinq.EngineDeck4PlayState,
	stagelinq.EngineDeck4CurrentBPM,
	stagelinq.EngineDeck4TrackArtistName,
	stagelinq.EngineDeck4TrackSongName,
	stagelinq.MixerCH4faderPosition,
}

func makeStateMap() map[string]bool {
	retval := map[string]bool{}
	for _, value := range stateValues {
		retval[value] = false
	}
	return retval
}

func allStateValuesReceived(v map[string]bool) bool {
	for _, value := range v {
		if !value {
			return false
		}
	}
	return true
}

var artistName1 string
var artistName2 string
var artistName3 string
var artistName4 string
var songName1 string
var songName2 string
var songName3 string
var songName4 string
var BPM1 float64
var BPM2 float64
var BPM3 float64
var BPM4 float64
var fltft1 float64
var fltft2 float64
var fltft3 float64
var fltft4 float64
var play1 bool
var play2 bool
var play3 bool
var play4 bool

func main() {
	cleartextfiles1()
	cleartextfiles2()
	cleartextfiles3()
	cleartextfiles4()
	listener, err := stagelinq.ListenWithConfiguration(&stagelinq.ListenerConfiguration{
		DiscoveryTimeout: timeout,
		SoftwareName:     appName,
		SoftwareVersion:  appVersion,
		Name:             "OBS_Plug",
	})
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	listener.AnnounceEvery(time.Second)

	deadline := time.After(timeout)
	foundDevices := []*stagelinq.Device{}

	log.Printf("Listening for devices for %s", timeout)

discoveryLoop:
	for {
		select {
		case <-deadline:
			break discoveryLoop
		default:
			device, deviceState, err := listener.Discover(timeout)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
				continue discoveryLoop
			}
			if device == nil {
				continue
			}
			// ignore device leaving messages since we do a one-off list
			if deviceState != stagelinq.DevicePresent {
				continue discoveryLoop
			}
			// check if we already found this device before
			for _, foundDevice := range foundDevices {
				if foundDevice.IsEqual(device) {
					continue discoveryLoop
				}
			}
			foundDevices = append(foundDevices, device)
			log.Printf("%s %q %q %q", device.IP.String(), device.Name, device.SoftwareName, device.SoftwareVersion)

			// discover provided services
			log.Println("\tattempting to connect to this device…")
			deviceConn, err := device.Connect(listener.Token(), []*stagelinq.Service{})
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			} else {
				defer deviceConn.Close()
				log.Println("\trequesting device data services…")
				services, err := deviceConn.RequestServices()
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
					continue
				}

				for _, service := range services {
					log.Printf("\toffers %s at port %d", service.Name, service.Port)
					switch service.Name {
					case "StateMap":
						stateMapTCPConn, err := device.Dial(service.Port)
						defer stateMapTCPConn.Close()
						if err != nil {
							log.Printf("WARNING: %s", err.Error())
							continue
						}
						stateMapConn, err := stagelinq.NewStateMapConnection(stateMapTCPConn, listener.Token())
						if err != nil {
							log.Printf("WARNING: %s", err.Error())
							continue
						}

						m := makeStateMap()
						//read values from console
						for _, stateValue := range stateValues {
							stateMapConn.Subscribe(stateValue)
						}

						for state := range stateMapConn.StateC() {
							m[state.Name] = true
							//scan variables into temp files
							//deck1
							if state.Name == "/Engine/Deck1/PlayState" {
								play1 = state.Value["state"].(bool)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, play1)
							}
							if state.Name == "/Engine/Deck1/CurrentBPM" {
								BPM1 = state.Value["value"].(float64)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, BPM1)
							}
							if state.Name == "/Engine/Deck1/Track/ArtistName" {
								artistName1 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, artistName1)
							}
							if state.Name == "/Engine/Deck1/Track/SongName" {
								songName1 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, songName1)
							}
							if state.Name == "/Mixer/CH1faderPosition" {
								fltft1 = state.Value["value"].(float64)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft1)
							}
							//scan variables into temp files
							//deck2
							if state.Name == "/Engine/Deck2/PlayState" {
								play2 = state.Value["state"].(bool)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, play2)
							}
							if state.Name == "/Engine/Deck2/CurrentBPM" {
								BPM2 = state.Value["value"].(float64)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, BPM2)
							}
							if state.Name == "/Engine/Deck2/Track/ArtistName" {
								artistName2 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, artistName2)
							}
							if state.Name == "/Engine/Deck2/Track/SongName" {
								songName2 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, songName2)
							}
							if state.Name == "/Mixer/CH2faderPosition" {
								fltft2 = state.Value["value"].(float64)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft2)
							}
							//scan variables into temp files
							//deck3
							if state.Name == "/Engine/Deck3/PlayState" {
								play3 = state.Value["state"].(bool)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, play3)
							}
							if state.Name == "/Engine/Deck3/CurrentBPM" {
								BPM3 = state.Value["value"].(float64)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, BPM3)
							}
							if state.Name == "/Engine/Deck3/Track/ArtistName" {
								artistName3 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, artistName3)
							}
							if state.Name == "/Engine/Deck3/Track/SongName" {
								songName3 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, songName3)
							}
							if state.Name == "/Mixer/CH3faderPosition" {
								fltft3 = state.Value["value"].(float64)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft3)
							}
							//scan variables into temp files
							//deck4
							if state.Name == "/Engine/Deck4/PlayState" {
								play4 = state.Value["state"].(bool)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, play4)
							}
							if state.Name == "/Engine/Deck4/CurrentBPM" {
								BPM4 = state.Value["value"].(float64)
								//log.Printf("%s %s %v", device.IP.String(), state.Name, BPM4)
							}
							if state.Name == "/Engine/Deck4/Track/ArtistName" {
								artistName4 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, artistName4)
							}
							if state.Name == "/Engine/Deck4/Track/SongName" {
								songName4 = state.Value["string"].(string)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, songName4)
							}
							if state.Name == "/Mixer/CH4faderPosition" {
								fltft4 = state.Value["value"].(float64)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft4)
							}
							if allStateValuesReceived(m) {
								checkifplaying1()
								checkifplaying2()
								checkifplaying3()
								checkifplaying4()
							}
						}
						select {
						case err := <-stateMapConn.ErrorC():
							log.Printf("WARNING: %s", err.Error())
						default:
						}
						stateMapTCPConn.Close()
					}
				}

				log.Println("\tend of list of device data services")
			}
		}
	}

	log.Printf("Found devices: %d", len(foundDevices))
}
func cleartextfiles1() {
	writeFile1("")
	writefileart1("")
	writefiletitle1("")
}
func cleartextfiles2() {
	writeFile2("")
	writefileart2("")
	writefiletitle2("")
}
func cleartextfiles3() {
	writeFile3("")
	writefileart3("")
	writefiletitle3("")
}
func cleartextfiles4() {
	writeFile4("")
	writefileart4("")
	writefiletitle4("")
}
func checkifplaying1() {
	if play1 == true {
		if fltft1 > 0.99 {
			//The fader is not a zero value
			//writeFile1("Deck 1 Now Playing: " + artistName1 + " - " + songName1 + " BPM: " + strconv.FormatFloat(BPM1, 'f', 2, 64) + "    ")
			writeFile1("Deck - 1: BPM:" + strconv.FormatFloat(BPM1, 'f', 2, 64))
			writefileart1(artistName1)
			writefiletitle1(songName1)
			fmt.Println("Deck 1 Now Playing: " + artistName1 + " - " + songName1 + " BPM: " + strconv.FormatFloat(BPM1, 'f', 2, 64))
		} else {
			cleartextfiles1()
			fmt.Println("Deck 1 Now Playing: ")
		}
	} else {
		cleartextfiles1()
		fmt.Println("Deck 1 Now Playing: ")
	}
}
func checkifplaying2() {
	if play2 == true {
		if fltft2 > 0.99 {
			//The fader is not a zero value
			//writeFile2("Now Playing: " + artistName2 + " - " + songName2 + " BPM: " + strconv.FormatFloat(BPM2, 'f', 2, 64) + "    ")
			writeFile2("Deck - 2: BPM:" + strconv.FormatFloat(BPM2, 'f', 2, 64))
			writefileart2(artistName2)
			writefiletitle2(songName2)
			fmt.Println("Deck 2 Now Playing: " + artistName2 + " - " + songName2 + " BPM: " + strconv.FormatFloat(BPM2, 'f', 2, 64))
		} else {
			cleartextfiles2()
			fmt.Println("Deck 2 Now Playing: ")
		}
	} else {
		cleartextfiles2()
		fmt.Println("Deck 2 Now Playing: ")
	}
}
func checkifplaying3() {
	if play3 == true {
		if fltft3 > 0.99 {
			//The fader is not a zero value
			//writeFile3("Now Playing: " + artistName3 + " - " + songName3 + " BPM: " + strconv.FormatFloat(BPM3, 'f', 2, 64) + "    ")
			fmt.Println("Deck 3 Now Playing: " + artistName3 + " - " + songName3 + " BPM: " + strconv.FormatFloat(BPM3, 'f', 2, 64))
			writeFile3("Deck - 3: BPM:" + strconv.FormatFloat(BPM3, 'f', 2, 64))
			writefileart3(artistName3)
			writefiletitle3(songName3)
		} else {
			cleartextfiles3()
			fmt.Println("Deck 3 Now Playing: ")
		}
	} else {
		cleartextfiles3()
		fmt.Println("Deck 3 Now Playing: ")
	}
}
func checkifplaying4() {
	if play4 == true {
		if fltft4 > 0.99 {
			//The fader is not a zero value
			//writeFile4("Now Playing: " + artistName4 + " - " + songName4 + " BPM: " + strconv.FormatFloat(BPM4, 'f', 2, 64) + "    ")
			writeFile4("Deck - 4: BPM:" + strconv.FormatFloat(BPM4, 'f', 2, 64))
			fmt.Println("Deck 4 Now Playing: " + artistName4 + " - " + songName4 + " BPM: " + strconv.FormatFloat(BPM4, 'f', 2, 64))
			writefileart4(artistName4)
			writefiletitle4(songName4)
		} else {
			cleartextfiles4()
			fmt.Println("Deck 4 Now Playing: ")
		}
	} else {
		cleartextfiles4()
		fmt.Println("Deck 4 Now Playing: ")
	}
}
func writeFile1(text string) {

	file, err := os.OpenFile(`./Deck1.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writeFile2(text string) {

	file, err := os.OpenFile(`./Deck2.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writeFile3(text string) {

	file, err := os.OpenFile(`./Deck3.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writeFile4(text string) {

	file, err := os.OpenFile(`./Deck4.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefileart1(text string) {

	file, err := os.OpenFile(`./SnipDeck1_Artist.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefileart2(text string) {

	file, err := os.OpenFile(`./SnipDeck2_Artist.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefileart3(text string) {

	file, err := os.OpenFile(`./SnipDeck3_Artist.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefileart4(text string) {

	file, err := os.OpenFile(`./SnipDeck4_Artist.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefiletitle1(text string) {

	file, err := os.OpenFile(`./SnipDeck1_Track.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefiletitle2(text string) {

	file, err := os.OpenFile(`./SnipDeck2_Track.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefiletitle3(text string) {

	file, err := os.OpenFile(`./SnipDeck3_Track.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
func writefiletitle4(text string) {

	file, err := os.OpenFile(`./SnipDeck4_Track.txt`, os.O_WRONLY|os.O_CREATE, 0666)
	file.Truncate(0)
	if err != nil {
		log.Printf("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, "%v", text)

	w.Flush()

}
