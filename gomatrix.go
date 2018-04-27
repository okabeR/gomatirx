package main

import (
    "net/http"
    "encoding/json"
	"fmt"
    "bytes"
	"io/ioutil"
	"log"
    "strings"
	//"time"

	"github.com/jroimartin/gocui"
    "github.com/buger/jsonparser"
)

type User struct{
    Type      string `json:"type"`
    User      string `json:"user"`
    Password  string `json:"password"`

}

type  Response struct {
    Access_token string
    Home_server  string
    User_id	 string
    Device_id    string
}

type  Room struct {
	Room_id string `json:"room_id"`
}

type Text struct {
	Msgtype string  `json:"msgtype"`
	Body    string  `json:"body"`
}

type Sync struct {
    Next_batch string `json:"next_batch"`
}

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

var MessageView *gocui.View
var Gtoken string
var resp Response
var room_key string = "!ZrzlKsobMlMKSrwjyQ:amadeus0.science"

func Login() {
	u := User{Type:"m.login.password", User:"xxx", Password:"xxxxx"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	res, _ := http.Post("http://amadeus0.science:8008/_matrix/client/r0/login", "application/json; charset=utf-8", b)
	temp, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(temp, &resp)
    Gtoken = resp.Access_token
}

func SendMessage(g *gocui.Gui, v *gocui.View) error{
    stuff := v.ViewBuffer()

    var roomid Room
	room, _ := http.Get("http://amadeus0.science:8008/_matrix/client/r0/publicRooms")
	tmpr,_ := ioutil.ReadAll(room.Body)
	json.Unmarshal(tmpr, &roomid)
	//fmt.Println(roomid.Room_id)

	m := Text{Msgtype: "m.text", Body:stuff}
	p := new(bytes.Buffer)
	json.NewEncoder(p).Encode(m)
	var room_key string = "!ZrzlKsobMlMKSrwjyQ:amadeus0.science"
    v.Clear()
    v.SetCursor(0,0)

	http.Post("http://amadeus0.science:8008/_matrix/client/r0/rooms/" + room_key + "/send/m.room.message?access_token=" + Gtoken, "application/json; charset=utf-8", p)
    SyncNew()   
    return nil

}


var Since string

func SyncNew() {
        oldmsg,_ := http.Get("http://amadeus0.science:8008/_matrix/client/r0/rooms/" + room_key + "/messages?from="+ Since +"&dir=f&access_token=" + Gtoken)
        old,_ := ioutil.ReadAll(oldmsg.Body)
   jsonparser.ArrayEach(old, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	clean,_,_,_ := jsonparser.Get(value, "content", "body") 
    fmt.Fprint(MessageView, strings.Replace(string(clean), `\n`, "\n", -1))
}, "chunk")

sync, _ := http.Get("http://amadeus0.science:8008/_matrix/client/r0/sync?since="+ Since +"&access_token=" + Gtoken)
        tmp_sync,_ := ioutil.ReadAll(sync.Body)    
    
        var sync2 Sync
        json.Unmarshal(tmp_sync, &sync2)    

    Since = sync2.Next_batch
}

func SyncOld() {
sync, _ := http.Get("http://amadeus0.science:8008/_matrix/client/r0/sync?access_token=" + Gtoken)
        tmp_sync,_ := ioutil.ReadAll(sync.Body)    
    
        var sync2 Sync
        json.Unmarshal(tmp_sync, &sync2)    
        Since = sync2.Next_batch

        oldmsg,_ := http.Get("http://amadeus0.science:8008/_matrix/client/r0/rooms/" + room_key + "/messages?from="+ sync2.Next_batch +"&dir=b&access_token=" + Gtoken)
        old,_ := ioutil.ReadAll(oldmsg.Body)
   jsonparser.ArrayEach(old, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	clean,_,_,_ := jsonparser.Get(value, "content", "body")
    fmt.Fprint(MessageView,  strings.Replace(string(clean), `\n`, "\n", -1))
}, "chunk")
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
