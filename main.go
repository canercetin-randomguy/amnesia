package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

type TestComm struct {
	Message string `json:"message,omitempty"`
}
type ChatPage struct {
	app.Compo
	Name            string
	MsgArray        []string
	UpdateAvailable bool
}

func (g *ChatPage) onInputChange(ctx app.Context, e app.EventHandler) {
	name := ctx.JSSrc().Get("value").String()
	ctx.NewActionWithValue("ChatPage", name) // Creating "ChatPage" action.
}

func (g *ChatPage) ReadWSTesting(ctx context.Context, conn *websocket.Conn, ctxApp app.Context) string {
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
		return temp.Message
	} else {
		return ""
	}
}
func (g *ChatPage) OnAppUpdate(ctx app.Context) {
	g.UpdateAvailable = ctx.AppUpdateAvailable()
	ctx.Reload()
}
func (g *ChatPage) MountWS(ctx app.Context) {
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
		ctx.SetState("message", g.MsgArray)
		time.Sleep(time.Second * 2)
	}
}
func (g *ChatPage) OnMount(ctx app.Context) {
	ctx.ObserveState("message").While(func() bool {
		return g.MsgArray != nil
	}).OnChange(func() {
		g.Update()
	})
}
func (g *ChatPage) onClick(ctx app.Context, e app.Event) {
	// connect to ws://localhost:8000/ws
	go g.MountWS(ctx)
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (g *ChatPage) Render() app.UI {
	return app.Div().Body(
		// connect to the websocket and display the message from websocket
		app.Button().OnClick(g.onClick).Body(
			app.Text("Click me"),
		),
		// If UpdateAvailable from ReadWSTesting, then update the UI
		app.Range(g.MsgArray).Slice(func(i int) app.UI {
			return app.Div().Body(
				app.Li().Body(
					app.Text(g.MsgArray[i]),
				),
			)
		}))

}

func main() {
	app.Route("/", &ChatPage{})
	app.Route("/ChatPage", &ChatPage{})
	app.RunWhenOnBrowser()

	// add a route for the websocket without the app.Route
	http.HandleFunc("/ws", Mount)
	http.Handle("/", &app.Handler{})

	if err := http.ListenAndServe(":3169", nil); err != nil {
		log.Fatal(err)
	}
}
