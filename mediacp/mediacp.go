// Package mediacp manages media renderers and servers on the UPnP network.
//
package mediacp

import (
	"github.com/sqp/gupnp/upnpcp"

	"io/ioutil"
	"path"
	"time"
)

//
//------------------------------------------------------------[ MEDIACONTROL ]--

// MediaControl manages media renderers and servers on the UPnP network.
//
type MediaControl struct {
	upnpcp.ControlPoint

	renderer  *upnpcp.Renderer
	renderers upnpcp.Renderers

	server  *upnpcp.Server
	servers map[string]*upnpcp.Server

	hooks map[string]*MediaHook

	preferredRenderer string
	preferredServer   string
	seekDelta         int
	volumeDelta       int

	tmpDir string
}

// New creates a new MediaControl manager.
//
func New() (*MediaControl, error) {

	media := &MediaControl{
		renderers: make(upnpcp.Renderers),
		servers:   make(map[string]*upnpcp.Server),
		hooks:     make(map[string]*MediaHook),
	}

	media.ControlPoint = *upnpcp.NewControlPoint(upnpcp.ControlPointEvents{
		OnRendererFound: media.onRendererFound,
		OnServerFound:   media.onMediaServerFound,
		OnRendererLost:  media.onMediaRendererLost,
		OnServerLost:    media.onMediaServerLost})

	var e error
	media.tmpDir, e = ioutil.TempDir("", "tvplay")

	return media, e
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Actions defined by the media controler.
//
const (
	ActionNone = iota
	ActionToggleMute
	ActionVolumeDown
	ActionVolumeUp
	ActionPlayPause
	ActionStop
	ActionSeekBackward
	ActionSeekForward
)

// Action sends an action message to the selected renderer.
//
func (cp *MediaControl) Action(action int) {
	renderer := cp.Renderer()
	if renderer == nil {
		return
	}

	switch action {
	case ActionToggleMute:
		renderer.ToggleMute()

	case ActionVolumeDown:
		renderer.SetVolumeDelta(-cp.volumeDelta)

	case ActionVolumeUp:
		renderer.SetVolumeDelta(cp.volumeDelta)

	case ActionPlayPause:
		renderer.PlayPause()

	case ActionStop:
		renderer.Stop()

	case ActionSeekBackward:
		cp.SetSeek(cp.GetCurrentTime() - cp.seekDelta)

	case ActionSeekForward:
		cp.SetSeek(cp.GetCurrentTime() + cp.seekDelta)
	}
}

//
//---------------------------------------------------------------[ RENDERERS ]--

// Renderer return the current renderer if any.
//
func (cp *MediaControl) Renderer() *upnpcp.Renderer {
	return cp.renderer
}

// RendererExists return true if a renderer is selected.
//
func (cp *MediaControl) RendererExists() bool {
	return cp.renderer != nil
}

// RendererIsActive return true if the provided server is the one selected.
//
func (cp *MediaControl) RendererIsActive(rend *upnpcp.Renderer) bool {
	return cp.renderer == rend
}

// GetRenderer return the renderer referenced by the udn argument.
//
func (cp *MediaControl) GetRenderer(udn string) *upnpcp.Renderer {
	return cp.renderers[udn]
}

// SetRenderer sets the current renderer by its udn reference.
//
func (cp *MediaControl) SetRenderer(udn string) {
	cp.renderer = cp.GetRenderer(udn)
	cp.onRendererSelected(cp.renderer)
}

func (cp *MediaControl) setRendererDefault() {
	if !cp.RendererExists() && cp.preferredRenderer != "" {
		for _, r := range cp.Renderers() {
			if cp.preferredRenderer == r.Name {
				cp.SetRenderer(r.Udn)
			}
		}
	}
}

// Renderers return the list of all known renderers.
//
func (cp *MediaControl) Renderers() upnpcp.Renderers {
	return cp.renderers
}

//
//-----------------------------------------------------------------[ SERVERS ]--

// Server return the current server if any.
//
func (cp *MediaControl) Server() *upnpcp.Server {
	return cp.server
}

// ServerExists return true if a server is selected.
//
func (cp *MediaControl) ServerExists() bool {
	return cp.server != nil
}

// ServerIsActive return true if the provided server is the one selected.
//
func (cp *MediaControl) ServerIsActive(srv *upnpcp.Server) bool {
	return cp.server == srv
}

// GetServer return the server referenced by the udn argument.
//
func (cp *MediaControl) GetServer(udn string) *upnpcp.Server {
	return cp.servers[udn]
}

// SetServer sets the current server by its udn reference.
//
func (cp *MediaControl) SetServer(udn string) {
	cp.server = cp.GetServer(udn)
	cp.onServerSelected(cp.server)
}

func (cp *MediaControl) setServerDefault() {
	if !cp.ServerExists() && cp.preferredServer != "" {
		for _, r := range cp.Servers() {
			if cp.preferredServer == r.Name {
				cp.SetServer(r.Udn)
			}
		}
	}
}

// Servers return the list of all known servers.
//
func (cp *MediaControl) Servers() map[string]*upnpcp.Server {
	return cp.servers
}

//
//-----------------------------------------------------------------[ ?? ]--

// SetVolumeDelta configures the default volume delta for user actions.
//
func (cp *MediaControl) SetVolumeDelta(delta int) {
	cp.volumeDelta = delta
	cp.onSetVolumeDelta(delta)
}

// SetVolumeDelta configures the default volume delta for user actions.
//
func (cp *MediaControl) SetSeekDelta(delta int) {
	cp.seekDelta = delta
	cp.onSetSeekDelta(delta)
}

// SetPreferredRenderer sets the renderer that will be selected if found (unless anoter is selected).
//
func (cp *MediaControl) SetPreferredRenderer(name string) {
	cp.preferredRenderer = name
	cp.setRendererDefault()
	// cp.onSetPreferredRenderer(name)
}

// SetPreferredServer sets the server that will be selected if found (unless anoter is selected)..
//
func (cp *MediaControl) SetPreferredServer(name string) {
	cp.preferredServer = name
	cp.setServerDefault()
}

//
//------------------------------------------------------------------[ BROWSE ]--

// Browse lists files on a server.
//
func (cp *MediaControl) Browse(container string, startingIndex uint) ([]upnpcp.Container, []upnpcp.Item, uint, uint) {

	return cp.server.Browse(container, startingIndex, uint(upnpcp.MAX_BROWSE))
}

// BrowseMetadata starts the playback of the given file on the selected renderer.
//
func (cp *MediaControl) BrowseMetadata(container string, startingIndex uint) { //([]upnpcp.Container, []upnpcp.Item, uint, uint) {
	_, items, didlxml := cp.server.BrowseMetadata(container, startingIndex, uint(upnpcp.MAX_BROWSE))
	for _, item := range items {

		// log.Info("RES", len(item.Res))
		if len(item.Res) > 0 {
			// log.Info("RES", didlxml)
			if cp.RendererExists() {
				cp.Renderer().SetURL(item.Res[0].URL, didlxml)
			}
		}
	}
}

//
//--------------------------------------------------------------------[ TIME ]--

// SetSeek seeks to new time in track. Input in seconds.
//
func (cp *MediaControl) SetSeek(secs int) {
	if cp.Renderer() != nil {
		if secs < 0 {
			secs = 0
		}
		cp.Renderer().SetSeek(TimeToString(secs))
	}
}

// SetSeekPercent seeks to new time in track. Input is the percent position in track. Range 0 to 100.
//
func (cp *MediaControl) SetSeekPercent(value float64) {
	if cp.Renderer() != nil {
		cp.SetSeek(int(float64(cp.Renderer().Duration) * value / 100))
	}
}

// target = "%d:%02d:%02d" % (hours,minutes,seconds)

// GetCurrentTime returns the current track position on selected server in seconds.
//
func (cp *MediaControl) GetCurrentTime() int {
	if cp.Renderer() != nil {
		return cp.Renderer().GetCurrentTime()
	}
	return -1
}

// TimeToString format time input in seconds. output as "15:04:05" format for seek requests.
//
func TimeToString(sec int) string {
	newtime := time.Time{}.Add(time.Duration(sec) * time.Second)
	return newtime.Format("15:04:05")
}

//
//---------------------------------------------------------[ LOCAL CALLBACKS ]--

func (cp *MediaControl) onRendererSelected(r *upnpcp.Renderer) {
	for _, instance := range cp.hookTest(testRendererSelected) {
		instance.OnRendererSelected(r)
	}
}

func (cp *MediaControl) onServerSelected(s *upnpcp.Server) {
	for _, instance := range cp.hookTest(testServerSelected) {
		instance.OnServerSelected(s)
	}
}

func (cp *MediaControl) onSetVolumeDelta(delta int) {
	for _, instance := range cp.hookTest(testSetVolumeDelta) {
		instance.OnSetVolumeDelta(delta)
	}
}

func (cp *MediaControl) onSetSeekDelta(delta int) {
	for _, instance := range cp.hookTest(testSetSeekDelta) {
		instance.OnSetSeekDelta(delta)
	}
}

//
//----------------------------------------------------------[ UPNP CALLBACKS ]--

func (cp *MediaControl) onRendererFound(r *upnpcp.Renderer) {
	cp.renderers[r.Udn] = r

	if cp.tmpDir != "" {
		r.Icon = r.GetIconFile(path.Join(cp.tmpDir, r.Udn)) // Get device icon.
	}

	for _, instance := range cp.hookTest(testRendererFound) { // forward device found event.
		instance.OnRendererFound(r)
	}

	cp.setRendererDefault() // Now we can test if we need to select it.

	// Connect renderer events to renderer hooks.
	r.Events.OnTransportState = func(rcb *upnpcp.Renderer, value upnpcp.PlaybackState) {
		for _, instance := range cp.hookTestRenderer(rcb, testTransportState) {
			instance.OnTransportState(rcb, value)
		}
	}

	r.Events.OnCurrentTrackDuration = func(rcb *upnpcp.Renderer, value int) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTrackDuration) {
			instance.OnCurrentTrackDuration(rcb, value)
		}
	}

	r.Events.OnCurrentTrackMetaData = func(rcb *upnpcp.Renderer, value *upnpcp.Item) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTrackMetaData) {
			instance.OnCurrentTrackMetaData(rcb, value)
		}
	}

	r.Events.OnMute = func(rcb *upnpcp.Renderer, value bool) {
		for _, instance := range cp.hookTestRenderer(rcb, testMute) {
			instance.OnMute(rcb, value)
		}
	}

	r.Events.OnVolume = func(rcb *upnpcp.Renderer, value uint) {
		for _, instance := range cp.hookTestRenderer(rcb, testVolume) {
			instance.OnVolume(rcb, value)
		}
	}

	r.Events.OnCurrentTime = func(rcb *upnpcp.Renderer, value int, percent float64) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTime) {
			instance.OnCurrentTime(rcb, value, percent)
		}
	}
}

func (cp *MediaControl) onMediaServerFound(srv *upnpcp.Server) {
	cp.servers[srv.Udn] = srv

	if cp.tmpDir != "" {
		srv.Icon = srv.GetIconFile(path.Join(cp.tmpDir, srv.Udn)) // Get device icon.
	}

	for _, instance := range cp.hookTest(testServerFound) { // forward device found event.
		instance.OnServerFound(srv)
	}

	// cp.setServerDefault() // Now we can test if we need to select it.
}

func (cp *MediaControl) onMediaRendererLost(rtest *upnpcp.Renderer) {
	for _, rend := range cp.Renderers() {
		if rend.CompareProxy(rtest) {
			if cp.RendererIsActive(rend) { // was active, unselect and forward info.
				cp.SetRenderer("")
			}

			delete(cp.renderers, rend.Udn) // delete from our index.

			for _, instance := range cp.hookTest(testRendererLost) { // forward device lost event.
				instance.OnRendererLost(rend)
			}
		}
	}
}

func (cp *MediaControl) onMediaServerLost(stest *upnpcp.Server) {
	for _, srv := range cp.Servers() {
		if srv.CompareProxy(stest) {
			if cp.ServerIsActive(srv) { // was active, unselect and forward info.
				cp.SetServer("")
			}

			delete(cp.servers, srv.Udn) // delete from our index.

			for _, instance := range cp.hookTest(testServerLost) { // forward device lost event.
				instance.OnServerLost(srv)
			}
		}
	}
}

//
//-------------------------------------------------------------------[ HOOKS ]--

// MediaHook provides a registration method to media events for multiple clients.
//
type MediaHook struct {
	upnpcp.ControlPointEvents
	upnpcp.RendererEvents

	// Local events.
	OnSetVolumeDelta   func(int)
	OnSetSeekDelta     func(int)
	OnRendererSelected func(*upnpcp.Renderer)
	OnServerSelected   func(*upnpcp.Server)
}

// SubscribeHook register a new hook client and returns the MediaHook to connect to.
//
func (cp *MediaControl) SubscribeHook(id string) *MediaHook {
	newHook := &MediaHook{RendererEvents: upnpcp.RendererEvents{}}
	cp.hooks[id] = newHook
	return newHook
}

// UnsubscribeHook removes a client hook.
//
func (cp *MediaControl) UnsubscribeHook(id string) {
	delete(cp.hooks, id)
}

func (cp *MediaControl) hookTestRenderer(rend *upnpcp.Renderer, test func(instance *MediaHook) bool) (ret []*MediaHook) {
	if cp.RendererIsActive(rend) {
		return cp.hookTest(test)
	}
	return nil
}

func (cp *MediaControl) hookTest(test func(instance *MediaHook) bool) (ret []*MediaHook) {
	for _, instance := range cp.hooks {
		if test(instance) {
			ret = append(ret, instance)
		}
	}
	return
}

func testTransportState(h *MediaHook) bool       { return h.OnTransportState != nil }
func testCurrentTrackDuration(h *MediaHook) bool { return h.OnCurrentTrackDuration != nil }
func testCurrentTrackMetaData(h *MediaHook) bool { return h.OnCurrentTrackMetaData != nil }
func testMute(h *MediaHook) bool                 { return h.OnMute != nil }
func testVolume(h *MediaHook) bool               { return h.OnVolume != nil }
func testCurrentTime(h *MediaHook) bool          { return h.OnCurrentTime != nil }

func testSetVolumeDelta(h *MediaHook) bool   { return h.OnSetVolumeDelta != nil }
func testSetSeekDelta(h *MediaHook) bool     { return h.OnSetSeekDelta != nil }
func testRendererFound(h *MediaHook) bool    { return h.OnRendererFound != nil }
func testServerFound(h *MediaHook) bool      { return h.OnServerFound != nil }
func testRendererLost(h *MediaHook) bool     { return h.OnRendererLost != nil }
func testServerLost(h *MediaHook) bool       { return h.OnServerLost != nil }
func testRendererSelected(h *MediaHook) bool { return h.OnRendererSelected != nil }
func testServerSelected(h *MediaHook) bool   { return h.OnServerSelected != nil }

//
//---------------------------------------------------[ RenderControl PARSING ]--

// type RCS struct {
// 	PresetNameList   Vstring
// 	Contrast         Vint
// 	Sharpness        Vint
// 	ColorTemperature Vint
// 	Mute             Vint // dropped attr channel
// 	Volume           Vint // dropped attr channel
// 	Brightness       Vint
// }

// type gAVT struct {
// 	InstanceID AVT
// }

// type AVT struct {
// 	TransportState               string
// 	TransportStatus              string
// 	TransportPlaySpeed           int
// 	NumberOfTracks               int
// 	CurrentMediaDuration         string // ?
// 	AVTransportURI               string // ?
// 	AVTransportURIMetaData       string // ?
// 	PlaybackStorageMedium        string
// 	CurrentTrack                 int
// 	CurrentTrackDuration         string
// 	CurrentTrackMetaData         string // ?
// 	CurrentTrackURI              string // ?
// 	CurrentTransportActions      string // ?
// 	NextAVTransportURI           string
// 	NextAVTransportURIMetaData   string
// 	RecordStorageMedium          string
// 	RecordMediumWriteStatus      string
// 	PossiblePlaybackStorageMedia string
// 	PossibleRecordStorageMedia   string
// 	PossibleRecordQualityModes   string
// 	CurrentPlayMode              string
// 	CurrentRecordQualityMode     string
// }
