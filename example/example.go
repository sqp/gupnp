// Package example is a demo for the UPnP control point GUI.
package main

import (
	"github.com/conformal/gotk3/gtk"

	// "github.com/sqp/gupnp/backendsonos" // go UPnP backend.
	"github.com/sqp/gupnp/backendgupnp" // gupnp backend.

	"github.com/sqp/gupnp"          // UPnP control point.
	"github.com/sqp/gupnp/guigtk"   // UPnP gui.
	"github.com/sqp/gupnp/upnptype" // UPnP common types.

	"fmt"
)

func main() {
	gtk.Init(nil)

	// Create the UPnP device manager.
	handler := NewHandler(&logger{})

	// Connect an UPnP backend to the manager.
	// mgr := backendsonos.NewManager(&logger{app.Log()})
	// mgr.SetEvents(app.cp.DefineEvents())
	// go mgr.Start(true)

	backend := backendgupnp.NewControlPoint()
	backend.SetEvents(handler.cp.DefineEvents())

	// Create the control window.
	guigtk.WindowTitle = "gupnp example"

	gui, win := guigtk.NewGui(handler.cp)
	if gui == nil {
		return
	}
	gui.Load()

	win.Connect("delete-event", gtk.MainQuit)
	// win.SetIconFromFile(file)

	gtk.Main() // gtk main loop is required by the GUI and the gupnp backend.
}

// Handler defines a gupnp client with some connected callbacks.
type Handler struct {
	cp  upnptype.MediaControl
	log *logger
}

// NewHandler creates a handler to show how to use gupnp callbacks.
func NewHandler(log *logger) *Handler {
	cp, e := gupnp.New(&logger{})
	if e != nil {
		log.Warningf("temp dir: %s\n", e)
	}

	handler := &Handler{
		cp:  cp,
		log: log,
	}

	// Connect local tests (use this to extend or replace the GUI).
	hook := cp.SubscribeHook("maincall")
	hook.OnRendererFound = handler.onMediaRendererFound
	hook.OnServerFound = handler.onMediaServerFound
	hook.OnRendererLost = handler.onMediaRendererLost
	hook.OnServerLost = handler.onMediaServerLost

	return handler
}

func (o *Handler) onMediaRendererFound(r upnptype.Renderer) {
	o.log.Infof("Renderer Found: %s  -  %s\n", r.Name(), r.UDN())
}

func (o *Handler) onMediaServerFound(srv upnptype.Server) {
	o.log.Infof("Server Found: %s  -  %s\n", srv.Name(), srv.UDN())
}

func (o *Handler) onMediaRendererLost(r upnptype.Renderer) {
	o.log.Infof("Renderer Lost: %s  -  %s\n", r.Name(), r.UDN())
}

func (o *Handler) onMediaServerLost(srv upnptype.Server) {
	o.log.Infof("Server Lost: %s  -  %s\n", srv.Name(), srv.UDN())
}

//
//------------------------------------------------------------------[ LOGGER ]--

type logger struct{}

func (l *logger) Infof(pattern string, args ...interface{}) { fmt.Printf(pattern, args...) }

func (l *logger) Warningf(pattern string, args ...interface{}) { fmt.Printf(pattern, args...) }
