package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/curve25519"
)

func onlyDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

func generateWireGuardKeys() (privateKeyBase64, publicKeyBase64 string, err error) {
	var privateKey [32]byte
	_, err = rand.Read(privateKey[:])
	if err != nil {
		return "", "", err
	}
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64
	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	if err != nil {
		return "", "", err
	}
	privateKeyBase64 = base64.StdEncoding.EncodeToString(privateKey[:])
	publicKeyBase64 = base64.StdEncoding.EncodeToString(publicKey)
	return
}

func BuildS2STab(myWindow fyne.Window) *container.TabItem {
	client_endpoint := widget.NewEntry()
	client_address := widget.NewEntry()
	client_port := widget.NewEntry()
	client_name := widget.NewEntry()
	client_output := widget.NewMultiLineEntry()
	client_output.SetMinRowsVisible(10)
	server_output := widget.NewMultiLineEntry()
	server_output.SetMinRowsVisible(10)

	generate := widget.NewButton("Generovat", func() {
		client_priv, client_pub, err := generateWireGuardKeys()
		if err != nil {
			panic(err)
		}

		runValidation := func() {
			using_default_port := client_port.Text == "13231"
			if client_address.Text == "" || client_name.Text == "" || client_endpoint.Text == "" {
				dialog.NewCustom("Nevyplnil si vsetky udaje", "OK", container.NewCenter(widget.NewLabel("Prosim vypln udaje")), myWindow).Show()
				return
			}

			if !strings.HasPrefix(client_address.Text, "10.10.10.") {
				dialog.NewCustom("MGMT adresa nie je z daneho subnetu", "OK", container.NewCenter(widget.NewLabel("Prosim zmen adresu")), myWindow).Show()
				return
			}

			if net.ParseIP(client_address.Text) == nil {
				dialog.NewCustom("Nie je validna adresa", "OK", container.NewCenter(widget.NewLabel("Prosim zmen adresu")), myWindow).Show()
				return
			}

			if net.ParseIP(client_endpoint.Text) == nil {
				dialog.NewCustom("Nie je validna verejna adresa", "OK", container.NewCenter(widget.NewLabel("Prosim zmen verejnu adresu")), myWindow).Show()
				return
			}

			if !onlyDigits(client_port.Text) && !using_default_port {
				dialog.NewCustom("port musi obsahovat iba cisla", "OK", container.NewCenter(widget.NewLabel("Prosim zmen port")), myWindow).Show()
				return
			}

			client_cfg := fmt.Sprintf(`/interface wireguard add name="WireGuard S2S" private-key="%s" listen-port=%s
`, client_priv, client_port.Text)
			client_cfg += `/interface wireguard peers add allowed-address=0.0.0.0/0 endpoint-address=94.228.84.20 endpoint-port=13231 interface="WireGuard S2S" name=clientToServer persistent-keepalive=25s public-key="maeWVrYaGRPOQZZ1G97+xZH1FPbc4u5y//xLeEa3Fwc="
`
			client_cfg += fmt.Sprintf(`/ip/address/add address=%s/24 interface="WireGuard S2S" network=10.10.10.0
`, client_address.Text)
			client_cfg += fmt.Sprintf(`/ip/firewall/filter add action=accept chain=input comment="WG input" in-interface="WireGuard S2S" place-before=2
`)
			client_cfg += fmt.Sprintf(`/ip/firewall/filter add action=accept chain=input comment="WG port" dst-port=%s protocol=udp place-before=2
`, client_port.Text)
			server_cfg := fmt.Sprintf(`/interface wireguard peers add allowed-address=%s/32 endpoint-address=%s endpoint-port=%s interface=wg-s2s name=%s persistent-keepalive=25s public-key="%s"
`, client_address.Text, client_endpoint.Text, client_port.Text, client_name.Text, client_pub)

			client_output.SetText(client_cfg)
			server_output.SetText(server_cfg)
		}

		if client_port.Text == "" {
			dialog.NewConfirm("Neuvedol si port", "Pouzijem port 13231, je to ok?", func(b bool) {
				if b {
					client_port.SetText("13231")
					runValidation()
				}
			}, myWindow).Show()
		} else {
			runValidation()
		}
	})

	s2sCanvasObj := container.NewVBox(
		widget.NewLabel("Nazov klienta"),
		client_name,
		widget.NewLabel("Verejna IP klienta"),
		client_endpoint,
		widget.NewLabel("Mgmt IP klienta"),
		client_address,
		widget.NewLabel("Port pre WG klienta"),
		client_port,
		widget.NewLabel("Konfig pre klienta"),
		client_output,
		widget.NewLabel("Konfig pre server"),
		server_output,
		generate,
	)

	return container.NewTabItem("S2S", s2sCanvasObj)
}
