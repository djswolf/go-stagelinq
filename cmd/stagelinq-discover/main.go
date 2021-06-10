package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
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
var ch1 int = 0
var ch2 int = 0
var ch3 int = 0
var ch4 int = 0
var stateplay1 int = 0
var stateplay2 int = 0
var stateplay3 int = 0
var stateplay4 int = 0
var check4 int
var check3a int
var check3b int
var check3c int
var check2a int
var check2b int
var check2c int
var check2d int
var check2e int
var check2f int
var test1 []int
var check4play int
var check3playa int
var check3playb int
var check3playc int
var check2playa int
var check2playb int
var check2playc int
var check2playd int
var check2playe int
var check2playf int
var test2 []int
var channel1 []int
var channel1up bool
var channel2 []int
var channel2up bool
var channel3 []int
var channel3up bool
var channel4 []int
var channel4up bool
var deck1play []int
var deck2play []int
var deck3play []int
var deck4play []int
var deck1playingon bool
var deck2playingon bool
var deck3playingon bool
var deck4playingon bool
var nodecksplaying []int
var nodeckscanbeheard []int
var alldecksoff bool
var allfadersdown bool

func main() {
	cleartextfiles1()
	cleartextfiles2()
	cleartextfiles3()
	cleartextfiles4()
	writeFile1("")
	writeFile2("")
	writeFile3("")
	writeFile4("\n")
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
								if play1 == true {
									stateplay1 = 1
								} else {
									stateplay1 = 0
								}

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
								ch1 = int(fltft1)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft1)
							}
							//scan variables into temp files
							//deck2
							if state.Name == "/Engine/Deck2/PlayState" {
								play2 = state.Value["state"].(bool)
								if play2 == true {
									stateplay2 = 1
								} else {
									stateplay2 = 0
								}
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
								ch2 = int(fltft2)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft2)
							}
							//scan variables into temp files
							//deck3
							if state.Name == "/Engine/Deck3/PlayState" {
								play3 = state.Value["state"].(bool)
								if play3 == true {
									stateplay3 = 1
								} else {
									stateplay3 = 0
								}
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
								ch3 = int(fltft3)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft3)
							}
							//scan variables into temp files
							//deck4
							if state.Name == "/Engine/Deck4/PlayState" {
								play4 = state.Value["state"].(bool)
								if play4 == true {
									stateplay4 = 1
								} else {
									stateplay4 = 0
								}
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
								ch4 = int(fltft4)
								//log.Printf("%s %s %s", device.IP.String(), state.Name, fltft4)
							}
							if allStateValuesReceived(m) {
								checkstatesums()
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
func checkstatesums() {
	//chechchannel possitions using xor
	check4 = ch1 ^ ch2 ^ ch3 ^ ch4
	check3a = ch1 ^ ch2 ^ ch3
	check3b = ch1 ^ ch2 ^ ch4
	check3c = ch2 ^ ch3 ^ ch4
	check2a = ch1 ^ ch2
	check2b = ch1 ^ ch3
	check2c = ch1 ^ ch4
	check2d = ch2 ^ ch3
	check2e = ch2 ^ ch4
	check2f = ch3 ^ ch4
	test1 = nil
	channel1 = nil
	channel2 = nil
	channel3 = nil
	channel4 = nil
	nodeckscanbeheard = nil
	test1 = append(test1, check4, check3a, check3b, check3c, check2a, check2b, check2c, check2d, check2e, check2f)
	channel1 = append(channel1, 1, 1, 1, 0, 1, 1, 1, 0, 0, 0)
	channel2 = append(channel2, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0)
	channel3 = append(channel3, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1)
	channel4 = append(channel4, 1, 0, 1, 1, 0, 0, 1, 0, 1, 1)
	nodeckscanbeheard = append(nodeckscanbeheard, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	channel1up = (reflect.DeepEqual(channel1, test1))
	channel2up = (reflect.DeepEqual(channel2, test1))
	channel3up = (reflect.DeepEqual(channel3, test1))
	channel4up = (reflect.DeepEqual(channel4, test1))
	allfadersdown = (reflect.DeepEqual(nodeckscanbeheard, test1))
	//checkplaystates using xor
	check4play = stateplay1 ^ stateplay2 ^ stateplay3 ^ stateplay4
	check3playa = stateplay1 ^ stateplay2 ^ stateplay3
	check3playb = stateplay1 ^ stateplay2 ^ stateplay4
	check3playc = stateplay2 ^ stateplay3 ^ stateplay4
	check2playa = stateplay1 ^ stateplay2
	check2playb = stateplay1 ^ stateplay3
	check2playc = stateplay1 ^ stateplay4
	check2playd = stateplay2 ^ stateplay3
	check2playe = stateplay2 ^ stateplay4
	check2playf = stateplay3 ^ stateplay4
	test2 = nil
	deck1play = nil
	deck2play = nil
	deck3play = nil
	deck4play = nil
	nodecksplaying = nil
	test2 = append(test2, check4play, check3playa, check3playb, check3playc, check2playa, check2playb, check2playc, check2playd, check2playe, check2playf)
	deck1play = append(deck1play, 1, 1, 1, 0, 1, 1, 1, 0, 0, 0)
	deck2play = append(deck2play, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0)
	deck3play = append(deck3play, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1)
	deck4play = append(deck4play, 1, 0, 1, 1, 0, 0, 1, 0, 1, 1)
	nodecksplaying = append(nodecksplaying, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	deck1playingon = (reflect.DeepEqual(deck1play, test2))
	deck2playingon = (reflect.DeepEqual(deck2play, test2))
	deck3playingon = (reflect.DeepEqual(deck3play, test2))
	deck4playingon = (reflect.DeepEqual(deck4play, test2))
	alldecksoff = (reflect.DeepEqual(nodecksplaying, test2))
	if allfadersdown == true {
		writeFile1("")
		writeFile2("")
		writeFile3("")
		writeFile4("\n")
		cleartextfiles1()
		cleartextfiles2()
		cleartextfiles3()
		cleartextfiles4()
		fmt.Println("All Decks Volume Off")

	}
	if alldecksoff == true {
		writeFile1("")
		writeFile2("")
		writeFile3("")
		writeFile4("\n")
		cleartextfiles1()
		cleartextfiles2()
		cleartextfiles3()
		cleartextfiles4()
		fmt.Println("All Decks Stopped")

	}
	if stateplay1 == 1 { //deck1playing
		if ch1 == 1 { //channel 1 up 100%
			if channel1up == true {
				//if the only channel up to 100% output to file
				fmt.Println("Deck 1 Now Playing: " + artistName1 + " - " + songName1 + " BPM: " + strconv.FormatFloat(BPM1, 'f', 2, 64))
				writeFile1(artistName1 + "-|-" + songName1 + "\n")
				nowplaying1()
			}
			if deck1playingon {
				//if the only deck playing output to file
				fmt.Println("Deck 1 Now Playing: " + artistName1 + " - " + songName1 + " BPM: " + strconv.FormatFloat(BPM1, 'f', 2, 64))
				writeFile1(artistName1 + "-|-" + songName1 + "\n")
				nowplaying1()
			}
		}
		fmt.Println("Deck 1 Coming Up: " + artistName1 + " - " + songName1 + " BPM: " + strconv.FormatFloat(BPM1, 'f', 2, 64))
	} else {
		cleartextfiles1()
		fmt.Println("Deck 1 Now Playing: ")
	}
	if stateplay2 == 1 { //deck1playing
		if ch2 == 1 { //channel 1 up 100%
			if channel2up == true { //if the only channel up to 100% output to file
				fmt.Println("Deck 2 Now Playing: " + artistName2 + " - " + songName2 + " BPM: " + strconv.FormatFloat(BPM2, 'f', 2, 64))
				writeFile2(artistName2 + "-|-" + songName2 + "\n")
				nowplaying2()
			}
			if deck2playingon {
				fmt.Println("Deck 2 Now Playing: " + artistName2 + " - " + songName2 + " BPM: " + strconv.FormatFloat(BPM2, 'f', 2, 64))
				writeFile2(artistName2 + "-|-" + songName2 + "\n")
				//if the only deck playing output to file
				nowplaying2()
			}
		}
		fmt.Println("Deck 2 Coming Up: " + artistName2 + " - " + songName2 + " BPM: " + strconv.FormatFloat(BPM2, 'f', 2, 64))
	} else {
		cleartextfiles2()
		fmt.Println("Deck 2 Now Playing: ")
	}
	if stateplay3 == 1 { //deck3playing
		if ch3 == 1 { //channel 3 up 100%
			if channel3up == true { //if the only channel up to 100% output to file
				fmt.Println("Deck 3 Now Playing: " + artistName3 + " - " + songName3 + " BPM: " + strconv.FormatFloat(BPM3, 'f', 2, 64))
				writeFile3(artistName3 + "-|-" + songName3 + "\n")
				nowplaying3()
			}
			if deck3playingon {
				fmt.Println("Deck 3 Now Playing: " + artistName3 + " - " + songName3 + " BPM: " + strconv.FormatFloat(BPM3, 'f', 2, 64))
				writeFile3(artistName3 + "-|-" + songName3 + "\n")
				nowplaying3()
				//if the only deck playing output to file
			}
		}
		fmt.Println("Deck 3 Coming Up: " + artistName3 + " - " + songName3 + " BPM: " + strconv.FormatFloat(BPM3, 'f', 2, 64))
	} else {
		cleartextfiles3()
		fmt.Println("Deck 3 Now Playing: ")
	}
	if stateplay4 == 1 { //deck4playing
		if ch4 == 1 { //channel 4 up 100%
			if channel4up == true { //if the only channel up to 100% output to file
				writeFile4(artistName4 + "-|-" + songName4 + "\n")
				fmt.Println("Deck 4 Now Playing: " + artistName4 + " - " + songName4 + " BPM: " + strconv.FormatFloat(BPM4, 'f', 2, 64))
				nowplaying4()
			}
			if deck4playingon {
				fmt.Println("Deck 4 Now Playing: " + artistName4 + " - " + songName4 + " BPM: " + strconv.FormatFloat(BPM4, 'f', 2, 64))
				writeFile4(artistName4 + "-|-" + songName4 + "\n")
				nowplaying4()
				//if the only deck playing output to file
			}
		}
		fmt.Println("Deck 4 Coming Up: " + artistName4 + " - " + songName4 + " BPM: " + strconv.FormatFloat(BPM4, 'f', 2, 64))
	} else {
		cleartextfiles4()
		fmt.Println("Deck 4 Now Playing: ")
	}
}

func cleartextfiles1() {
	writefileart1("")
	writefiletitle1("")
}
func cleartextfiles2() {
	writefileart2("")
	writefiletitle2("")
}
func cleartextfiles3() {
	writefileart3("")
	writefiletitle3("")
}
func cleartextfiles4() {
	writefileart4("")
	writefiletitle4("")
}
func nowplaying1() {

	writefileart1(artistName1)
	writefiletitle1(songName1)

}
func nowplaying2() {

	writefileart2(artistName2)
	writefiletitle2(songName2)

}
func nowplaying3() {
	writefileart3(artistName3)
	writefiletitle3(songName3)
}
func nowplaying4() {
	writefileart4(artistName4)
	writefiletitle4(songName4)
}
func writeFile1(text string) {

	file, err := os.OpenFile(`./output.txt`, os.O_WRONLY|os.O_CREATE, 0666)
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

	file, err := os.OpenFile(`./output.txt`, os.O_WRONLY|os.O_CREATE, 0666)
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

	file, err := os.OpenFile(`./output.txt`, os.O_WRONLY|os.O_CREATE, 0666)
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

	file, err := os.OpenFile(`./output.txt`, os.O_WRONLY|os.O_CREATE, 0666)
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
