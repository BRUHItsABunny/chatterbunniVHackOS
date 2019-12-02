package main

import (
	"chatterBunniVhackOS/clevergo"
	"chatterBunniVhackOS/color"
	"chatterBunniVhackOS/vhackapi"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/jroimartin/gocui"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type ChatterBunni struct {
	IrcClient    *irc.Conn
	VHackClient  vhackapi.VHackOSClient
	CleverClient clevergo.CleverClient
	Gui          *gocui.Gui
	Editor       gocui.EditorFunc
}

var bunni ChatterBunni

func main() {
	clever := clevergo.GetClient()
	vhack := vhackapi.GetClient("settings.json")
	if !vhack.Init() {
		fmt.Println("Error logging in, call the Login function manually to find out why")
	} else {
		g, _ := gocui.NewGui(gocui.Output256)
		g.SetManagerFunc(layout)

		g.Cursor = true
		g.Mouse = false

		if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return gocui.ErrQuit
			}); err != nil {
			log.Panicln(err)
		}
		if err := g.SetKeybinding("chat", gocui.KeyCtrlT, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				_, err := g.SetCurrentView("input")
				return err
			}); err != nil {
			log.Panicln(err)
		}
		if err := g.SetKeybinding("chat", gocui.KeyArrowUp, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				scrollView(v, -1)
				return nil
			}); err != nil {
		}
		if err := g.SetKeybinding("chat", gocui.KeyArrowDown, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				scrollView(v, 1)
				return nil
			}); err != nil {
		}
		cfg := irc.NewConfig(vhack.ClientSettings.Account.Username)
		cfg.SSL = false
		cfg.Me.Ident = vhack.ClientSettings.Account.Uid
		cfg.Me.Name = "vhack.Mobile User"
		cfg.Server = vhack.ClientSettings.ChatAddress
		cfg.NewNick = func(n string) string { return n + "^" }
		c := irc.Client(cfg)
		bunni = ChatterBunni{c, vhack, clever, g, simpleEditor}

		c.HandleFunc(irc.CONNECTED,
			func(conn *irc.Conn, line *irc.Line) {
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					_, err := fmt.Fprintln(output, "CONNECTED")
					return err
				})
				conn.Privmsg("vMobileGuard", ".letmein "+vhack.ClientSettings.Account.Uid+" "+vhack.ClientSettings.Account.Token+" en")
			})
		// And a signal on disconnect
		quit := make(chan bool)
		c.HandleFunc(irc.DISCONNECTED,
			func(conn *irc.Conn, line *irc.Line) { quit <- true })
		c.HandleFunc(irc.PRIVMSG,
			func(conn *irc.Conn, line *irc.Line) {
				cmd := strings.Split(line.Text(), " ")
				speak := true
				var co int
				var visual string
				switch {
				case strings.HasPrefix(line.Host, "vip"):
					co = 153
				case strings.HasPrefix(line.Host, "vip1"):
					co = 153
				case strings.HasPrefix(line.Host, "vip2"):
					co = 11
				case strings.HasPrefix(line.Host, "vip3"):
					co = 201
				case strings.HasPrefix(line.Host, "vip4"):
					co = 33
				case strings.HasPrefix(line.Host, "vip5"):
					co = 128
				case strings.HasPrefix(line.Host, "vip6"):
					co = 214
				case strings.HasPrefix(line.Host, "admin"):
					co = 9
				}
				switch {
				case strings.Contains(line.Text(), "@bunnibunbun"):
					visual = color.String(46, line.Text())
				case strings.Contains(line.Text(), "bunnibunbun"):
					visual = color.String(46, line.Text())
				case strings.Contains(line.Text(), "@bunni"):
					visual = color.String(46, line.Text())
				case strings.Contains(line.Text(), "bunni"):
					visual = color.String(46, line.Text())
				default:
					visual = line.Text()
				}
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					_, err := fmt.Fprintln(output, color.String(co, line.Nick)+" said <- "+visual)
					return err
				})
				if cmd[0] == "bunbun" {
					speak = false
					// Check what cmd is being issued
					if len(cmd) > 2 {
						switch cmd[1] {
						case "ping":
							bunni.IrcClient.Privmsg("#english", "pong")
						}
					}
				}

				if speak {
					if strings.Contains(line.Text(), "@bunnibunbun") {
						session, err := clever.Ask(line.Text())
						time.Sleep(time.Duration(len(session.Response)/randInt(1, 10)) * time.Second)
						if err == nil {
							if randInt(10, 110) < 90 {
								if randInt(10, 110) < 55 {
									session.Response = "@" + line.Nick + " " + session.Response
								} else {
									session.Response = session.Response + " " + "@" + line.Nick
								}
							}
							bunni.IrcClient.Privmsg("#english", session.Response)
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "BOT said -> "+session.Response)
								return err
							})
						} else {
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "ERROR:"+err.Error())
								return err
							})
							//sendText(conn, "#english", "Eh? What?")
						}
					} else if strings.Contains(line.Text(), "@bunni") {
						session, err := clever.Ask(line.Text())
						time.Sleep(time.Duration(len(session.Response)/randInt(1, 10)) * time.Second)
						if err == nil {
							if randInt(10, 110) < 90 {
								if randInt(10, 110) < 55 {
									session.Response = "@" + line.Nick + " " + session.Response
								} else {
									session.Response = session.Response + " " + "@" + line.Nick
								}
							}
							bunni.IrcClient.Privmsg("#english", session.Response)
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "BOT said -> "+session.Response)
								return err
							})
						} else {
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "ERROR:"+err.Error())
								return err
							})
							//sendText(conn, "#english", "Eh? What?")
						}
					} else if randInt(10, 110) > 65 {
						session, err := clever.Ask(line.Text())
						time.Sleep(time.Duration(len(session.Response)/randInt(1, 10)) * time.Second)
						if err == nil {
							if randInt(10, 110) > 85 {
								if randInt(10, 110) < 55 {
									session.Response = "@" + line.Nick + " " + session.Response
								} else {
									session.Response = session.Response + " " + "@" + line.Nick
								}
							}
							bunni.IrcClient.Privmsg("#english", session.Response)
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "BOT said -> "+session.Response)
								return err
							})
						} else {
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "ERROR:"+err.Error())
								return err
							})
							//sendText(conn, "#english", "Eh? What?")
						}
					}
				}

				// Keep RAM usage at bay, ~11MB
				clever.Messages = clever.Messages[len(clever.Messages)-7:]
			})
		if err := bunni.IrcClient.Connect(); err != nil {
			fmt.Println("Connection error: %s\n", err.Error())
		}

		err := g.MainLoop()
		<-quit
		if err != nil || err != gocui.ErrQuit {
			os.Exit(1)
		}
	}

}

func simpleEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		if line := v.ViewBuffer(); len(line) > 0 {
			bunni.IrcClient.Privmsg("#english", line)
			bunni.Gui.Update(func(gui *gocui.Gui) error {
				output, _ := bunni.Gui.View("chat")
				_, err := fmt.Fprint(output, "I said -> "+line)
				return err
			})
			v.Clear()
			v.SetCursor(0, 0)
			v.SetOrigin(0, 0)
		}
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:

		cx, _ := v.Cursor()
		line := v.ViewBuffer()
		if cx < len(line)-1 {
			v.MoveCursor(1, 0, false)
		}
	case key == gocui.KeyCtrlT:
		bunni.Gui.SetCurrentView("chat")
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("chat", 0, 0, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		_, err := g.SetCurrentView("chat")

		if err != nil {
			return err
		}

		v.BgColor = gocui.ColorDefault

		v.Autoscroll = true
		v.Wrap = true
		v.Frame = false

	}
	if v, err := g.SetView("input", 0, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		_, err := g.SetCurrentView("input")

		if err != nil {
			return err
		}

		v.Editor = bunni.Editor

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.ColorDefault

		v.Autoscroll = false
		v.Editable = true
		v.Wrap = false
		v.Frame = false

	}
	return nil

}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func scrollView(v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}
