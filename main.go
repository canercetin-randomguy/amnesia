package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

var cnt = 0

type TestComm struct {
	Message string `json:"message,omitempty"`
}
type greet struct {
	app.Compo
	Name            string
	MsgArray        []string
	UpdateAvailable bool
}

func (g *greet) onInputChange(ctx app.Context, e app.EventHandler) {
	name := ctx.JSSrc().Get("value").String()
	ctx.NewActionWithValue("greet", name) // Creating "greet" action.
}

func (g *greet) ReadWSTesting(ctx context.Context, conn *websocket.Conn, ctxApp app.Context) string {
	// declare a {}interface to hold the message as string
	var temp TestComm
	// read the message from the websocket
	_, r, err := conn.Reader(ctx)
	if err != nil {
		log.Fatal("reader opening:", err)
	}

	// decode
	err = json.NewDecoder(r).Decode(&temp)
	if err != nil {
		log.Fatal("decode:", err)
	}

	// write and return bool to indicate new messages
	if temp.Message != "" {
		fmt.Println("Received message from websocket:", temp.Message)
		g.OnAppUpdate(ctxApp)
		return temp.Message
	} else {
		return ""
	}
}
func (g *greet) OnAppUpdate(ctx app.Context) {
	fmt.Println("Setting condition to true")
	g.UpdateAvailable = ctx.AppUpdateAvailable()
	cnt++
	ctx.Reload()
}
func (g *greet) MountWS(ctx app.Context, e app.Event, wg *sync.WaitGroup) {
	var tempString string
	// connect to the ws://localhost:3169/ws endpoint
	ctxWS, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	// cancel the WS connection after 1 minute, or cancel it at the end of the function
	defer cancel()
	conn, _, err := websocket.Dial(ctxWS, "ws://localhost:3169/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")
	fmt.Println("Connected to websocket") //localhost:3169/ws

	// read the message from the websocket
	for {
		if err != nil {
			log.Fatal("readeropening:", err)
		}
		tempString = g.ReadWSTesting(ctxWS, conn, ctx)
		g.MsgArray = append(g.MsgArray, tempString)
		time.Sleep(time.Second * 2)
	}
}
func (g *greet) onClick(ctx app.Context, e app.Event) {
	// connect to ws://localhost:8000/ws
	wg := sync.WaitGroup{}
	wg.Add(1)
	go g.MountWS(ctx, e, &wg)
	wg.Wait()
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (g *greet) Render() app.UI {
	return app.Div().Body(
		// connect to the websocket and display the message from websocket
		app.Button().OnClick(g.onClick).Body(
			app.Text("Click me"),
		),
		app.H5().Body(
			app.Text(cnt),
		),
		// If UpdateAvailable from ReadWSTesting, then update the UI
		app.If(g.UpdateAvailable,
			app.Button().
				Text("Update to see the messages.")))
}

func main() {
	app.Route("/", &greet{})
	app.Route("/greet", &greet{})
	app.RunWhenOnBrowser()

	// add a route for the websocket without the app.Route
	http.HandleFunc("/ws", Mount)
	http.Handle("/", &app.Handler{})

	if err := http.ListenAndServe(":3169", nil); err != nil {
		log.Fatal(err)
	}
}
