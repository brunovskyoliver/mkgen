package main

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Mikrotik Conf")

	dhcp_ip := widget.NewEntry()
	dhcp_url := "http://172.29.5.92:8081"
	dhcp_output := widget.NewMultiLineEntry()
	dhcp_output.SetMinRowsVisible(10)
	connect := widget.NewButton("connect", func() {
		ip := dhcp_ip.Text
		if net.ParseIP(ip) == nil {
			dialog.NewCustom("Toto nie je validna adresa", "OK", container.NewCenter(widget.NewLabel("Prosim zmen adresu")), myWindow).Show()
			return
		}
		res, err := http.Get(fmt.Sprintf("%s?ip=%s", dhcp_url, ip))
		if err != nil {
			dialog.NewError(fmt.Errorf("Nepodarilo sa spojenie - error %s", err), myWindow).Show()
			return
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		dhcp_output.SetText(string(body))

	})

	dhcpCanvasObj := container.NewVBox(
		widget.NewLabel("IP adresa"),
		dhcp_ip,
		connect,
		dhcp_output,
	)

	tabs := container.NewAppTabs(
		BuildS2STab(myWindow),
		container.NewTabItem("DHCP", dhcpCanvasObj),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
