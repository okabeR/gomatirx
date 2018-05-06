package main

import (
	"fmt"
	"log"

    "github.com/jroimartin/gocui"
    "github.com/matrix-org/gomatrix"
)

func nextView(g *gocui.Gui, v *gocui.View) error {

	switch v.Name() {

	case "side":
		_, err := g.SetCurrentView("bottom")
		return err
	case "bottom":
		_, err := g.SetCurrentView("side")
		return err
	default:
		_, err := g.SetCurrentView("bottom")
		return err
	}

}
var cli *gomatrix.Client
var MessageView *gocui.View

func Login() {
    cli, _ = gomatrix.NewClient("http://amadeus0.science:8008", "", "")
	resp, err := cli.Login(&gomatrix.ReqLogin{
	Type:     "m.login.password",
	User:     "xxxx",
	Password: "xxxx",
	})
if err != nil {
    panic(err)
}

cli.SetCredentials(resp.UserID, resp.AccessToken)
}

func SendMessage(g *gocui.Gui, v *gocui.View) error{
    stuff := v.ViewBuffer()

    rooms, _ := cli.JoinedRooms()
    room_key := rooms.JoinedRooms[0]

    cli.SendMessageEvent(room_key,"m.room.message", gomatrix.TextMessage{MsgType: "m.room.message", Body: stuff})

    v.Clear()
    v.SetCursor(0,0)
    SyncNew()
    return nil
}


var Since string

func SyncNew() {

rooms, _ := cli.JoinedRooms()
room_key := rooms.JoinedRooms[0]

syn,_ := cli.SyncRequest(30000,"","",true, "online")
	new_mess,_ := cli.Messages(room_key, Since, "", 'f' , 100)
	   for _, value := range new_mess.Chunk {
	   fmt.Fprintln(MessageView, value.Content["body"])
	   }
   Since = syn.NextBatch
}

func SyncOld() {

	rooms, _ := cli.JoinedRooms()
	room_key := rooms.JoinedRooms[0]

	syn,_ := cli.SyncRequest(30000,"","",true, "online")
        listmap := syn.Rooms.Join
        messages := listmap[room_key].Timeline.Events
		for _, value := range messages {
		fmt.Fprintln(MessageView, value.Content["body"])
		}
	Since = syn.NextBatch
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, l)
		if _, err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}
	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("side"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("bottom", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("bottom", gocui.KeyEnter, gocui.ModNone, SendMessage); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", -1, -1, 15, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "Item 1")
		fmt.Fprintln(v, "Item 2")
		fmt.Fprintln(v, "Item 3")
		fmt.Fprint(v, "\rWill be")
		fmt.Fprint(v, "deleted\rItem 4\nItem 5")
	}
	if v, err := g.SetView("bottom", 15, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = false
		v.Editable = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "")
	}

	if v, err := g.SetView("main", 15, -1, maxX, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
        MessageView = v
        SyncOld()
        
		v.Editable = false
		v.Wrap = true
		if _, err := g.SetCurrentView("bottom"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
    
    Login()    

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

