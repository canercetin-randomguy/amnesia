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
	Name     string
	MsgArray []string
}

// ReadWSTesting is supposed to work for reading from the websocket.
//
// It works, it appends the values actually.
//
// TODO: filter received messages, e.g: if temp.Message includes WELCOME:xxxxx, read the part after welcome.
func (g *ChatPage) ReadWSTesting(ctx context.Context, conn *websocket.Conn, ctxApp app.Context) error {
	// declare a {}interface to hold the message as string
	var temp TestComm
	// read the message from the websocket
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, r, err := conn.Reader(ctx)
		if err != nil {
			return err
		}

		// decode
		err = json.NewDecoder(r).Decode(&temp)
		if err != nil {
			return err
		}

		// write and return bool to indicate new message
		if temp.Message != "" {
			fmt.Println("Received message from websocket:", temp.Message)
			err = nil
			ctxApp.Dispatch(func(ctxApp app.Context) {
				g.MsgArray = append(g.MsgArray, temp.Message)
			})
		} else {
			return err
		}
	}
	return nil
}

// OnAppUpdate is called when application has updates.
//
// By updates, not component updates, literally app code updates.
func (g *ChatPage) OnAppUpdate(ctx app.Context) {
	fmt.Println("App updated")
}

// MountWS is a process that will be running concurrently when clicked to a button.
//
// Will handle connections to the websocket.
func (g *ChatPage) MountWS(ctx app.Context) {
	ctxWS, cancel := context.WithTimeout(context.Background(), time.Minute*60)
	// cancel the WS connection after 1 minute, or cancel it at the end of the function
	defer cancel()
	// connect to the ws://localhost:3169/ws endpoint
	conn, _, err := websocket.Dial(ctxWS, "ws://localhost:3169/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	// close connection after done
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")
	fmt.Println("Connected to websocket") //localhost:3169/ws

	// read the message from the websocket
	for { // Read incoming signals every 2 seconds, append them to an array, and set newMessage state to true.
		err = g.ReadWSTesting(ctxWS, conn, ctx)
		if err != nil {
			// woops, connection is fucked. break the loop.
			fmt.Println("Error reading from websocket:", err)
			break
		}
	}
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
