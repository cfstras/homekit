package main

import (
	"log"
	"time"

	"net/http"
	"net/url"

	"io/ioutil"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
)

var urls = map[string]string{
	"redLight": "http://rail.fd:8080/Hutschiene/RedLight",
}

func main() {
	switchInfo := accessory.Info{
		Name:         "RedLight",
		Manufacturer: "flipdot Kassel",
		Model:        "hutschienenlampe",
		SerialNumber: "42",
	}
	acc := accessory.NewSwitch(switchInfo)

	config := hc.Config{Pin: "00000000", Port: "12345", StoragePath: "./db"}
	t, err := hc.NewIPTransport(config, acc.Accessory)

	if err != nil {
		log.Fatal(err)
	}

	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		var vals url.Values
		if on {
			log.Println("RedLight on!")
			vals = url.Values{"blink": {"true"}}
		} else {
			log.Println("RedLight off")
			vals = url.Values{"blink": {"false"}}
		}
		res, err := http.PostForm(urls["redLight"]+"?"+vals.Encode(), nil)
		if err != nil {
			log.Println("request error:", err)
			return
		} else {
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Println("error in body:", err)
				return
			}
			log.Println(string(b))
		}

		if on {
			go func() {
				time.Sleep(30 * time.Second)
				log.Println("lamp turns off after 30s")
				acc.Switch.On.SetValue(false)

			}()

		}
	})

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}
