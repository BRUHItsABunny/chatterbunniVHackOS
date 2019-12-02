package main

import (
	"chatterBunniVhackOS/clevergo"
	"chatterBunniVhackOS/color"
	"chatterBunniVhackOS/vhackapi"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/jroimartin/gocui"
	"github.com/tmjvonboss/gopedia"
	"github.com/tmjvonboss/ugodict"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type ChatterBunni struct {
	IrcClient       *irc.Conn
	VHackClient     vhackapi.VHackOSClient
	CleverClient    clevergo.CleverClient
	UrbanDictClient ugodict.UrbanClient
	WikipediaClient gopedia.WikiClient
	Gui             *gocui.Gui
	Editor          gocui.EditorFunc
}

type Job struct {
	Message    string
	Sleep      time.Duration
	IrcChannel string
}

var bunni ChatterBunni
var odds vhackapi.TalkingOdds
var timezone *time.Location
var channelOut chan Job

func main() {
	channelOut = make(chan Job)
	clever := clevergo.GetClient()
	vhack := vhackapi.GetClient("settings.json")
	urban := ugodict.GetClient()
	wiki := gopedia.GetClient()
	odds = vhack.ClientSettings.Account.Odds
	location, err := time.LoadLocation(vhack.ClientSettings.Account.Timezone)
	if err == nil {
		timezone = location
	} else {
		timezone = time.Local
	}
	clever.Language = vhack.ClientSettings.Account.Language
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
		if err := g.SetKeybinding("chat", gocui.KeyCtrlA, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				v.Autoscroll = true
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
		bunni = ChatterBunni{c, vhack, clever, urban, wiki, g, simpleEditor}

		c.HandleFunc(irc.CONNECTED,
			func(conn *irc.Conn, line *irc.Line) {
				go ircOutSynchronous()
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"CONNECTED AS "+bunni.VHackClient.ClientSettings.Account.Username+"["+bunni.VHackClient.ClientSettings.Account.Uid+"]")
					return err
				})
				conn.Privmsg("vMobileGuard", ".letmein "+vhack.ClientSettings.Account.Uid+" "+vhack.ClientSettings.Account.Token+" "+vhack.ClientSettings.Account.Language)
			})
		// And a signal on disconnect
		quit := make(chan bool)
		c.HandleFunc(irc.DISCONNECTED,
			func(conn *irc.Conn, line *irc.Line) {
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					_, _ = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] DISCONNECTED")
					return gocui.ErrQuit
				})
				quit <- true
			})
		c.HandleFunc(irc.JOIN,
			func(conn *irc.Conn, line *irc.Line) {
				co := 0
				switch {
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
					co = 208
				case strings.HasPrefix(line.Host, "admin"):
					co = 9
				default:
					if strings.HasPrefix(line.Host, "vip") {
						co = 153
					}
				}
				level := strings.Split(line.Host, ".")[0]
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					var err error
					if co == 0 {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+line.Nick+" JOINED AS REGULAR PLAYER")
					} else {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+color.String(co, line.Nick+" JOINED AS "+level))
					}
					return err
				})
			})
		c.HandleFunc(irc.QUIT,
			func(conn *irc.Conn, line *irc.Line) {
				co := 0
				switch {
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
					co = 208
				case strings.HasPrefix(line.Host, "admin"):
					co = 9
				default:
					if strings.HasPrefix(line.Host, "vip") {
						co = 153
					}
				}
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					var err error
					if co == 0 {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+line.Nick+" LEFT")
					} else {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+color.String(co, line.Nick+" LEFT"))
					}
					return err
				})
			})
		c.HandleFunc(irc.PRIVMSG,
			func(conn *irc.Conn, line *irc.Line) {
				userId := strings.Split(strings.Split(strings.Split(line.Raw, ":")[1], "@")[0], "!")[1]
				cmd := strings.Split(line.Text(), " ")
				speak := true
				co := 0
				var visual string
				switch {
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
					co = 208
				case strings.HasPrefix(line.Host, "admin"):
					co = 9
				default:
					if strings.HasPrefix(line.Host, "vip") {
						co = 153
					}
				}
				switch {
				case strings.Contains(line.Text(), bunni.VHackClient.ClientSettings.Account.Username):
					visual = color.String(46, line.Text())
				case len(bunni.VHackClient.ClientSettings.Account.Alias) > 0:
					if strings.Contains(line.Text(), bunni.VHackClient.ClientSettings.Account.Alias) {
						visual = color.String(46, line.Text())
					} else {
						visual = line.Text()
					}
				default:
					visual = line.Text()
				}
				bunni.Gui.Update(func(gui *gocui.Gui) error {
					output, _ := bunni.Gui.View("chat")
					var err error
					//_, err = fmt.Println(output, line.Raw)
					if co == 0 {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+line.Nick+" said <- "+visual)
					} else {
						_, err = fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+color.String(co, line.Nick)+" said <- "+visual)
					}
					return err
				})
				if cmd[0] == "bunbun" {
					speak = false
					// Check what cmd is being issued
					if len(cmd) > 2 {
						switch cmd[1] {
						case "ping":
							bunni.VHackClient.PrivateMessageUser(userId, "pong")
						case "urban":
							term := strings.Join(cmd[2:], " ")
							result, err, _ := bunni.UrbanDictClient.DefineByTerm(term)
							if err == nil {
								def := "Definition:\n" + result[0].Definition
								exa := "Example:\n" + result[0].Example
								vot := "Votes:\n\nThumbs up:" + strconv.Itoa(result[0].ThumbsUp) + "\nThumbs down:" + strconv.Itoa(result[0].ThumbsDown)
								go func() {
									bunni.VHackClient.PrivateMessageUser(userId, def)
									time.Sleep(time.Second)
									bunni.VHackClient.PrivateMessageUser(userId, exa)
									time.Sleep(time.Second)
									bunni.VHackClient.PrivateMessageUser(userId, vot)
								}()
							}
						case "wiki":
							term := strings.Join(cmd[2:], " ")
							result, err, _ := bunni.WikipediaClient.SearchBasic(term)
							if err == nil {
								reply := result[0].Title + "(" + strconv.Itoa(result[0].WordCount) + ") - " + result[0].TimeStamp + "\n\n" + result[0].Snippet
								bunni.VHackClient.PrivateMessageUser(userId, reply)
							}
						}
					}
				}

				if speak {
					random := randInt(10, 110)
					if strings.Contains(line.Text(), "@"+bunni.VHackClient.ClientSettings.Account.Username) {
						session, err := clever.Ask(line.Text())
						sleepTime := time.Duration(len(session.Response)/randInt(bunni.VHackClient.ClientSettings.Account.MinSleep, bunni.VHackClient.ClientSettings.Account.MaxSleep)) * time.Second
						if err == nil {
							if random < odds.Mentioned {
								if random < 55 {
									session.Response = "@" + line.Nick + " " + session.Response
								} else {
									session.Response = session.Response + " " + "@" + line.Nick
								}
							}
							job := Job{session.Response, sleepTime, bunni.VHackClient.ClientSettings.Account.Channel}
							channelOut <- job
						} else {
							bunni.Gui.Update(func(gui *gocui.Gui) error {
								output, _ := bunni.Gui.View("chat")
								_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"ERROR:"+err.Error())
								return err
							})
						}
					} else {
						if len(bunni.VHackClient.ClientSettings.Account.Alias) > 0 && strings.Contains(line.Text(), "@"+bunni.VHackClient.ClientSettings.Account.Alias) {
							session, err := clever.Ask(line.Text())
							sleepTime := time.Duration(len(session.Response)/randInt(bunni.VHackClient.ClientSettings.Account.MinSleep, bunni.VHackClient.ClientSettings.Account.MaxSleep)) * time.Second
							if err == nil {
								if random < odds.Mentioned {
									if random < 55 {
										session.Response = "@" + line.Nick + " " + session.Response
									} else {
										session.Response = session.Response + " " + "@" + line.Nick
									}
								}
								job := Job{session.Response, sleepTime, bunni.VHackClient.ClientSettings.Account.Channel}
								channelOut <- job
							} else {
								bunni.Gui.Update(func(gui *gocui.Gui) error {
									output, _ := bunni.Gui.View("chat")
									_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"ERROR:"+err.Error())
									return err
								})
							}
						} else {
							if random > odds.Random {
								session, err := clever.Ask(line.Text())
								sleepTime := time.Duration(len(session.Response)/randInt(bunni.VHackClient.ClientSettings.Account.MinSleep, bunni.VHackClient.ClientSettings.Account.MaxSleep)) * time.Second
								if err == nil {
									if random > 85 {
										if random < 55 {
											session.Response = "@" + line.Nick + " " + session.Response
										} else {
											session.Response = session.Response + " " + "@" + line.Nick
										}
									}
									job := Job{session.Response, sleepTime, bunni.VHackClient.ClientSettings.Account.Channel}
									channelOut <- job
								} else {
									bunni.Gui.Update(func(gui *gocui.Gui) error {
										output, _ := bunni.Gui.View("chat")
										_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"ERROR:"+err.Error())
										return err
									})
								}
							}
						}
					}
					clever.Messages = clever.Messages[len(clever.Messages)-7:]
				}
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
			bunni.IrcClient.Privmsg(bunni.VHackClient.ClientSettings.Account.Channel, line)
			bunni.Gui.Update(func(gui *gocui.Gui) error {
				output, _ := bunni.Gui.View("chat")
				_, err := fmt.Fprint(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"I said -> "+line)
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

func ircOutSynchronous() {
	for job := range channelOut {
		bunni.Gui.Update(func(gui *gocui.Gui) error {
			output, _ := bunni.Gui.View("chat")
			_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"BOT IS GONNA SAY -> "+job.Message)
			return err
		})
		time.Sleep(job.Sleep)
		bunni.IrcClient.Privmsg(job.IrcChannel, job.Message)
		bunni.Gui.Update(func(gui *gocui.Gui) error {
			output, _ := bunni.Gui.View("chat")
			_, err := fmt.Fprintln(output, "["+time.Now().In(timezone).Format("2006-01-02 15:04:05")+"] "+"BOT said -> "+job.Message)
			return err
		})
	}
}
