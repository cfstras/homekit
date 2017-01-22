package main

import (
	"log"

	"net/http"
	"net/url"

	"io/ioutil"

	"encoding/json"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/pkg/errors"

	"github.com/cfstras/homekit/pilight"
)

type Conf struct {
	Pilights []*PilightSwitch
	Urls     []*UrlSwitch
}

type Switch struct {
	Info      accessory.Info
	switchAcc *accessory.Switch
}
type UrlSwitch struct {
	Switch
	BaseURL string
	Param   string
	Values  [2]string
}
type PilightSwitch struct {
	Switch
	Protocols    []string
	System, Unit int
}

var urls = map[string]string{
	"redLight": "http://rail.fd:8080/Hutschiene/RedLight",
}

func main() {
	conf, err := loadConfig("conf.json")
	p(err, "loading config")
	switchInfo := accessory.Info{
		Name:         "RedLight",
		Manufacturer: "flipdot Kassel",
		Model:        "hutschienenlampe",
		SerialNumber: "42",
	}
	acc := accessory.NewSwitch(switchInfo)

	var accs []*accessory.Accessory
	for _, sw := range conf.Pilights {
		acc := accessory.NewSwitch(sw.Info)
		sw.Switch.switchAcc = acc
		acc.Switch.On.OnValueRemoteUpdate(sw.OnClick)
		accs = append(accs, sw.Switch.switchAcc.Accessory)
	}
	for _, sw := range conf.Urls {
		sw.Switch.switchAcc = accessory.NewSwitch(sw.Info)
		acc.Switch.On.OnValueRemoteUpdate(sw.OnClick)
		accs = append(accs, sw.Switch.switchAcc.Accessory)
	}

	if len(accs) == 0 {
		log.Fatalln("need at least one config.")
		return
	}

	config := hc.Config{Pin: "00000000", Port: "12345", StoragePath: "./db"}
	t, err := hc.NewIPTransport(config, accs[0], accs[1:]...)
	p(err, "Starting HomeKit server")

	hc.OnTermination(func() {
		t.Stop()
	})
	t.Start()
}
func btoi(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
func (s PilightSwitch) OnClick(on bool) {
	log.Println(s.Info.Name+" switch:", on)
	msg := pilight.Message{
		pilight.ActionSend,
		pilight.Code{
			Off:        btoi(!on),
			On:         btoi(on),
			Protocol:   s.Protocols,
			SystemCode: s.System,
			UnitCode:   s.Unit,
		},
	}
	msg.Send()
}

func (s UrlSwitch) OnClick(on bool) {
	log.Println(s.Info.Name+" switch:", on)
	vals := url.Values{s.Param: {s.Values[btoi(on)]}}
	res, err := http.PostForm(s.BaseURL+"?"+vals.Encode(), nil)
	if err != nil {
		log.Println("request error:", err)
		return
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("error in body:", err)
		return
	}
	log.Println(string(b))
}

func loadConfig(path string) (Conf, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		var c Conf
		c.Pilights = append(c.Pilights, &PilightSwitch{
			Switch{Info: accessory.Info{
				"somebody", "no123",
				"something", "0xbadf00d"}},
			[]string{"elro_800_switch"},
			25, 8,
		})
		c.Urls = append(c.Urls, &UrlSwitch{
			Switch{Info: accessory.Info{"somebody else", "no321",
				"something else", "0xbadf00d"}},
			"http://my.server/switch",
			"blink",
			[2]string{"true", "false"},
		})

		b, err := json.MarshalIndent(&c, "", "  ")
		p(err, "generating template file")
		p(ioutil.WriteFile(path, b, 0644), "writing template file")
		return c, errors.New("- I've created an example config you can use at '" + path + "'")
	}
	var c Conf
	p(json.Unmarshal(b, &c), "parsing configuration '"+path+"'")
	return c, nil
}

func p(err error, msg string, args ...interface{}) {
	if err != nil {
		args2 := append([]interface{}{"error", msg, err}, args...)
		log.Fatalln(args2)
	}
}
