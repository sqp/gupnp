// Package gupnp provides an API and GUI to control DLNA/UPnP devices like network TV and radio.
//
// Its goal is to locate media servers (with files) and media players.
// You can then send commands to players (volume, pause...) and let them play
// the selected music or video content.
// Can also be used as just a limited remote controler for supported renderers.
//
// Digital media servers - DMS
//
// A database of multimedia content, that other devices can play media from.
//
// Digital media renderers - DMR
//
// Plays stuff, that is it makes sound and in required cases shows moving images.
//
// Digital media controllers - DMC
//
// A device that works as a remote control, can play stop, skip, pause, change
// loudness, brightness etcetera.
//
// The manager keeps a server and a renderer as current target devices to use
// for fast user actions.
//
package gupnp

import (
	"github.com/sqp/gupnp/upnptype"

	"io/ioutil"
	"path"
)

//
//------------------------------------------------------------[ MEDIACONTROL ]--

// MediaControl manages media renderers and servers on the UPnP network.
//
type MediaControl struct {
	curRend   upnptype.Renderer
	renderers upnptype.Renderers

	curSrv  upnptype.Server
	servers map[string]upnptype.Server

	hooks map[string]*upnptype.MediaHook

	// User settings.
	preferredRenderer string
	preferredServer   string
	seekDelta         int
	volumeDelta       int

	tmpDir string

	log upnptype.Logger
}

// New creates a new MediaControl manager.
//
func New(log upnptype.Logger) (*MediaControl, error) {
	tmpDir, e := ioutil.TempDir("", "tvplay")
	return &MediaControl{
		renderers: make(upnptype.Renderers),
		servers:   make(map[string]upnptype.Server),
		hooks:     make(map[string]*upnptype.MediaHook),

		tmpDir: tmpDir,
		log:    log,
	}, e
}

// DefineEvents returns pointers to control events callbacks.
//
func (cp *MediaControl) DefineEvents() upnptype.ControlPointEvents {
	return upnptype.ControlPointEvents{
		OnRendererFound: cp.onRendererFound,
		OnServerFound:   cp.onMediaServerFound,
		OnRendererLost:  cp.onMediaRendererLost,
		OnServerLost:    cp.onMediaServerLost,
	}
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Action sends an action to the selected renderer.
//
func (cp *MediaControl) Action(action upnptype.Action) error {
	if cp.curRend == nil {
		return nil
	}

	var e error
	switch action {

	case upnptype.ActionToggleMute:
		muted, e := cp.curRend.GetMute(0, upnptype.ChannelMaster)
		if e == nil {
			e = cp.curRend.SetMute(0, upnptype.ChannelMaster, !muted)
		}

	case upnptype.ActionVolumeDown:
		vol, e := cp.curRend.GetVolume(0, upnptype.ChannelMaster)
		if e == nil {
			e = cp.curRend.SetVolume(0, upnptype.ChannelMaster, vol-uint16(cp.volumeDelta))
		}

	case upnptype.ActionVolumeUp:
		vol, e := cp.curRend.GetVolume(0, upnptype.ChannelMaster)
		if e == nil {
			e = cp.curRend.SetVolume(0, upnptype.ChannelMaster, vol+uint16(cp.volumeDelta))
		}

	case upnptype.ActionPlayPause:
		e = cp.curRend.PlayPause(0, upnptype.PlaySpeedNormal)

	case upnptype.ActionStop:
		e = cp.curRend.Stop(0)

	case upnptype.ActionSeekBackward:
		e = cp.curRend.Seek(0, upnptype.SeekModeAbsTime, upnptype.TimeToString(cp.GetCurrentTime()-cp.seekDelta))

	case upnptype.ActionSeekForward:
		e = cp.curRend.Seek(0, upnptype.SeekModeAbsTime, upnptype.TimeToString(cp.GetCurrentTime()+cp.seekDelta))
	}

	if e != nil {
		cp.log.Warningf("action failed #%d: %s", action, e)
	}
	return e
}

//
//---------------------------------------------------------------[ RENDERERS ]--

// Renderer return the current renderer if any.
//
func (cp *MediaControl) Renderer() upnptype.Renderer {
	return cp.curRend
}

// RendererExists return true if a renderer is selected.
//
func (cp *MediaControl) RendererExists() bool {
	return cp.curRend != nil
}

// RendererIsActive return true if the provided server is the one selected.
//
func (cp *MediaControl) RendererIsActive(rend upnptype.Renderer) bool {
	return cp.curRend == rend
}

// GetRenderer return the renderer referenced by the udn argument.
//
func (cp *MediaControl) GetRenderer(udn string) upnptype.Renderer {
	return cp.renderers[udn]
}

// SetRenderer sets the current renderer by its udn reference.
//
func (cp *MediaControl) SetRenderer(udn string) {
	cp.curRend = cp.GetRenderer(udn)
	cp.onRendererSelected(cp.curRend)

	if cp.curRend == nil {
		return
	}

	// Get current settings and forward to connected clients.
	vol, e := cp.curRend.GetVolume(0, upnptype.ChannelMaster)
	if e == nil {

		println("cp.curRend.Events().OnVolume", cp.curRend.Events().OnVolume != nil)

		cp.curRend.Events().OnVolume(cp.curRend, uint(vol))
	}

	mute, e := cp.curRend.GetMute(0, upnptype.ChannelMaster)
	if e == nil {
		cp.curRend.Events().OnMute(cp.curRend, mute)
	}

	transport, e := cp.curRend.GetTransportInfo(0)
	if e == nil {
		cp.curRend.Events().OnTransportState(cp.curRend, upnptype.PlaybackStateFromName(transport.CurrentTransportState))
	}

	positionInfo, e := cp.curRend.GetPositionInfo(0)
	if e == nil {
		// cp.log.Infof("GetPositionInfo(0) ->")
		// cp.log.Infof("\tTrack = %d", positionInfo.Track)
		// cp.log.Infof("\tTrackDuration = \"%s\"", positionInfo.TrackDuration)
		// cp.log.Infof("\tTrackMetaData = \"%s\"", positionInfo.TrackMetaData)

		duration := upnptype.TimeToSecond(positionInfo.TrackDuration)
		cp.curRend.Events().OnCurrentTrackDuration(cp.curRend, duration)
		// cp.curRend.Events().OnCurrentTrackMetaData(cp.curRend, positionInfo.TrackMetaData)
	}

	// if transportInfo, err := cp.curRend.GetTransportInfo(0); nil != err {
	// 	// t.Fatal(err)
	// } else {
	// 	cp.log.Infof("GetTransportInfo(0) ->")
	// 	cp.log.Infof("\tCurrentTransportState = \"%s\"", transportInfo.CurrentTransportState)
	// 	cp.log.Infof("\tTransportState ID = \"%d\"", upnptype.PlaybackStateFromName(transportInfo.CurrentTransportState))

	// 	cp.log.Infof("\tCurrentTransportStatus = \"%s\"", transportInfo.CurrentTransportStatus)
	// 	cp.log.Infof("\tCurrentSpeed = \"%s\"", transportInfo.CurrentSpeed)
	// }

	// if mediaInfo, err := cp.curRend.GetMediaInfo(0); nil != err {
	// 	// t.Fatal(err)
	// } else {
	// 	cp.log.Infof("GetMediaInfo(0) ->")
	// 	cp.log.Infof("\tNrTracks = %d", mediaInfo.NrTracks)
	// 	cp.log.Infof("\tMediaDuration = \"%s\"", mediaInfo.MediaDuration)
	// 	cp.log.Infof("\tCurrentURI = \"%s\"", mediaInfo.CurrentURI)
	// 	cp.log.Infof("\tCurrentURIMetaData = \"%s\"", mediaInfo.CurrentURIMetaData)
	// 	cp.log.Infof("\tNextURI = \"%s\"", mediaInfo.NextURI)
	// 	cp.log.Infof("\tNextURIMetaData = \"%s\"", mediaInfo.NextURIMetaData)
	// 	cp.log.Infof("\tPlayMedium = \"%s\"", mediaInfo.PlayMedium)
	// 	cp.log.Infof("\tRecordMedium = \"%s\"", mediaInfo.RecordMedium)
	// 	cp.log.Infof("\tWriteStatus = \"%s\"", mediaInfo.WriteStatus)
	// }

	// if positionInfo, err := cp.curRend.GetPositionInfo(0); nil != err {
	// 	// t.Fatal(err)
	// } else {
	// 	cp.log.Infof("GetPositionInfo(0) ->")
	// 	cp.log.Infof("\tTrack = %d", positionInfo.Track)
	// 	cp.log.Infof("\tTrackDuration = \"%s\"", positionInfo.TrackDuration)
	// 	cp.log.Infof("\tTrackMetaData = \"%s\"", positionInfo.TrackMetaData)
	// 	cp.log.Infof("\tTrackURI = \"%s\"", positionInfo.TrackURI)
	// 	cp.log.Infof("\tRelTime = \"%s\"", positionInfo.RelTime)
	// 	cp.log.Infof("\tAbsTime = \"%s\"", positionInfo.AbsTime)
	// 	cp.log.Infof("\tRelCount = %d", positionInfo.RelCount)
	// 	cp.log.Infof("\tAbsCount = %d", positionInfo.AbsCount)
	// }

	// if deviceCapabilities, err := cp.curRend.GetDeviceCapabilities(0); nil != err {
	// 	// t.Fatal(err)
	// } else {
	// 	cp.log.Infof("GetDeviceCapabilities() ->")
	// 	cp.log.Infof("\tPlayMedia = \"%s\"", deviceCapabilities.PlayMedia)
	// 	cp.log.Infof("\tRecMedia = \"%s\"", deviceCapabilities.RecMedia)
	// 	cp.log.Infof("\tRecQualityModes = \"%s\"", deviceCapabilities.RecQualityModes)
	// }

	// if transportSettings, err := cp.curRend.GetTransportSettings(0); nil != err {
	// 	// t.Fatal(err)
	// } else {
	// 	cp.log.Infof("GetTransportSettings() ->")
	// 	cp.log.Infof("\tPlayMode = \"%s\"", transportSettings.PlayMode)
	// 	cp.log.Infof("\tRecQualityMode = \"%s\"", transportSettings.RecQualityMode)
	// }
}

func (cp *MediaControl) setRendererDefault() {
	if !cp.RendererExists() && cp.preferredRenderer != "" {
		for _, r := range cp.Renderers() {
			if cp.preferredRenderer == r.Name() {
				cp.SetRenderer(r.UDN())
			}
		}
	}
}

// Renderers return the list of all known renderers.
//
func (cp *MediaControl) Renderers() upnptype.Renderers {
	return cp.renderers
}

//
//-----------------------------------------------------------------[ SERVERS ]--

// Server return the current server if any.
//
func (cp *MediaControl) Server() upnptype.Server {
	return cp.curSrv
}

// ServerExists return true if a server is selected.
//
func (cp *MediaControl) ServerExists() bool {
	return cp.curSrv != nil
}

// ServerIsActive return true if the provided server is the one selected.
//
func (cp *MediaControl) ServerIsActive(srv upnptype.Server) bool {
	return cp.curSrv == srv
}

// GetServer return the server referenced by the udn argument.
//
func (cp *MediaControl) GetServer(udn string) upnptype.Server {
	return cp.servers[udn]
}

// SetServer sets the current server by its udn reference.
//
func (cp *MediaControl) SetServer(udn string) {
	cp.curSrv = cp.GetServer(udn)
	cp.onServerSelected(cp.curSrv)
}

func (cp *MediaControl) setServerDefault() {
	if !cp.ServerExists() && cp.preferredServer != "" {
		for _, r := range cp.Servers() {
			if cp.preferredServer == r.Name() {
				cp.SetServer(r.UDN())
			}
		}
	}
}

// Servers return the list of all known servers.
//
func (cp *MediaControl) Servers() map[string]upnptype.Server {
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

// SetSeekDelta configures the default seek delta for user actions.
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
func (cp *MediaControl) Browse(container string, startingIndex uint32) ([]upnptype.Container, []upnptype.Object, uint, uint) {
	if !cp.ServerExists() {
		return nil, nil, 0, 0
	}
	req := &upnptype.BrowseRequest{
		ObjectID:      container,
		BrowseFlag:    "BrowseDirectChildren",
		Filter:        "@childCount",
		StartingIndex: startingIndex,
		RequestCount:  uint32(upnptype.MaxBrowse),
		// SortCriteria:  req.SortCriteria,
	}
	res, e := cp.curSrv.Browse(req)
	if e != nil {
		return nil, nil, 0, 0
	}
	return res.Container, res.Item, 0, 0

	// return cp.server.Browse(container, startingIndex, uint(upnptype.MaxBrowse))
}

// BrowseMetadata starts the playback of the given file on the selected renderer.
//
func (cp *MediaControl) BrowseMetadata(container string, startingIndex uint) error { //([]upnptype.Container, []upnptype.Item, uint, uint) {
	if cp.curRend == nil {
		return nil
	}
	_, items, didlxml := cp.curSrv.BrowseMetadata(container, startingIndex, uint(upnptype.MaxBrowse))
	for _, item := range items {

		// log.Info("RES", len(item.Res))
		if len(item.Res) > 0 {
			// log.Info("RES", didlxml)

			// if cp.RendererExists() {
			return cp.curRend.SetAVTransportURI(0, item.Res[0].URL, didlxml)
			// }
		}
	}
	return nil
}

// SetNextAVTransportURI sets the next playback URI,
//
func (cp *MediaControl) SetNextAVTransportURI(nextURI, nextURIMetaData string) error {
	if !cp.RendererExists() {
		return nil
	}
	return cp.Renderer().SetNextAVTransportURI(0, nextURI, nextURIMetaData)
}

// AddURIToQueue is TODO.
func (cp *MediaControl) AddURIToQueue(req *upnptype.AddURIToQueueIn) (*upnptype.AddURIToQueueOut, error) {
	if cp.curRend == nil {
		return nil, nil
	}

	println("EnqueuedURI", req.EnqueuedURI)

	_, items, didlxml := cp.curSrv.BrowseMetadata(req.EnqueuedURI, 0, uint(upnptype.MaxBrowse))
	for _, item := range items {

		// log.Info("RES", len(item.Res))
		if len(item.Res) > 0 {
			// log.Info("RES", didlxml)

			// if cp.RendererExists() {
			return cp.curRend.AddURIToQueue(0, &upnptype.AddURIToQueueIn{
				EnqueuedURI:         item.Res[0].URL,
				EnqueuedURIMetaData: didlxml,
			})
			// }
		}
	}

	return nil, nil
}

//
//--------------------------------------------------------------------[ TIME ]--

// Seek seeks to new time in track. Input in seconds.
//
func (cp *MediaControl) Seek(unit, target string) error {
	if cp.curRend == nil {
		return nil
	}
	return cp.curRend.Seek(0, unit, target)
}

// SeekPercent seeks to new time in track. Input is the percent position in track. Range 0 to 100.
//
func (cp *MediaControl) SeekPercent(value float64) error {
	if cp.curRend == nil {
		return nil
	}
	positionInfo, e := cp.curRend.GetPositionInfo(0)
	if e != nil {
		return e
	}

	percent := float64(upnptype.TimeToSecond(positionInfo.TrackDuration)) * value / 100

	// println("seek", positionInfo.TrackDuration, percent, upnptype.TimeToString(int(percent)))

	cp.Seek(upnptype.SeekModeAbsTime, upnptype.TimeToString(int(percent)))
	// cp.Renderer().Duration()  // was used
	return nil
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

//
//---------------------------------------------------------[ LOCAL CALLBACKS ]--

func (cp *MediaControl) onRendererSelected(r upnptype.Renderer) {
	for _, instance := range cp.hookTest(testRendererSelected) {
		instance.OnRendererSelected(r)
	}
}

func (cp *MediaControl) onServerSelected(s upnptype.Server) {
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

func (cp *MediaControl) onRendererFound(r upnptype.Renderer) {
	cp.renderers[r.UDN()] = r

	if cp.tmpDir != "" {
		r.SetIcon(r.GetIconFile(path.Join(cp.tmpDir, r.UDN()))) // Get device icon.
	}

	for _, instance := range cp.hookTest(testRendererFound) { // forward device found event.
		instance.OnRendererFound(r)
	}

	// Connect renderer events to renderer hooks.
	r.Events().OnTransportState = func(rcb upnptype.Renderer, value upnptype.PlaybackState) {
		for _, instance := range cp.hookTestRenderer(rcb, testTransportState) {
			instance.OnTransportState(rcb, value)
		}
	}

	r.Events().OnCurrentTrackDuration = func(rcb upnptype.Renderer, value int) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTrackDuration) {
			instance.OnCurrentTrackDuration(rcb, value)
		}
	}

	r.Events().OnCurrentTrackMetaData = func(rcb upnptype.Renderer, value *upnptype.Item) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTrackMetaData) {
			instance.OnCurrentTrackMetaData(rcb, value)
		}
	}

	r.Events().OnMute = func(rcb upnptype.Renderer, value bool) {
		for _, instance := range cp.hookTestRenderer(rcb, testMute) {
			instance.OnMute(rcb, value)
		}
	}

	r.Events().OnVolume = func(rcb upnptype.Renderer, value uint) {
		for _, instance := range cp.hookTestRenderer(rcb, testVolume) {
			instance.OnVolume(rcb, value)
		}
	}

	r.Events().OnCurrentTime = func(rcb upnptype.Renderer, value int, percent float64) {
		for _, instance := range cp.hookTestRenderer(rcb, testCurrentTime) {
			instance.OnCurrentTime(rcb, value, percent)
		}
	}

	cp.setRendererDefault() // Now we can test if we need to select it.
}

func (cp *MediaControl) onMediaServerFound(srv upnptype.Server) {
	cp.servers[srv.UDN()] = srv

	if cp.tmpDir != "" {
		srv.SetIcon(srv.GetIconFile(path.Join(cp.tmpDir, srv.UDN()))) // Get device icon.
	}

	for _, instance := range cp.hookTest(testServerFound) { // forward device found event.
		instance.OnServerFound(srv)
	}

	// cp.setServerDefault() // Now we can test if we need to select it.
}

func (cp *MediaControl) onMediaRendererLost(rtest upnptype.Renderer) {
	for _, rend := range cp.Renderers() {
		if rend.CompareProxy(rtest) {
			if cp.RendererIsActive(rend) { // was active, unselect and forward info.
				cp.SetRenderer("")
			}

			delete(cp.renderers, rend.UDN()) // delete from our index.

			for _, instance := range cp.hookTest(testRendererLost) { // forward device lost event.
				instance.OnRendererLost(rend)
			}
		}
	}
}

func (cp *MediaControl) onMediaServerLost(stest upnptype.Server) {
	for _, srv := range cp.Servers() {
		if srv.CompareProxy(stest) {
			if cp.ServerIsActive(srv) { // was active, unselect and forward info.
				cp.SetServer("")
			}

			delete(cp.servers, srv.UDN()) // delete from our index.

			for _, instance := range cp.hookTest(testServerLost) { // forward device lost event.
				instance.OnServerLost(srv)
			}
		}
	}
}

//
//-------------------------------------------------------------------[ HOOKS ]--

// SubscribeHook register a new hook client and returns the MediaHook to connect to.
//
func (cp *MediaControl) SubscribeHook(id string) *upnptype.MediaHook {
	newHook := &upnptype.MediaHook{RendererEvents: upnptype.RendererEvents{}}
	cp.hooks[id] = newHook
	return newHook
}

// UnsubscribeHook removes a client hook.
//
func (cp *MediaControl) UnsubscribeHook(id string) {
	delete(cp.hooks, id)
}

func (cp *MediaControl) hookTestRenderer(rend upnptype.Renderer, test func(instance *upnptype.MediaHook) bool) (ret []*upnptype.MediaHook) {
	if cp.RendererIsActive(rend) {
		return cp.hookTest(test)
	}
	return nil
}

// hookTest builds the list of registered clients implementing test.
//
func (cp *MediaControl) hookTest(test func(instance *upnptype.MediaHook) bool) (ret []*upnptype.MediaHook) {
	for _, instance := range cp.hooks {
		if test(instance) {
			ret = append(ret, instance)
		}
	}
	return
}

func testTransportState(h *upnptype.MediaHook) bool       { return h.OnTransportState != nil }
func testCurrentTrackDuration(h *upnptype.MediaHook) bool { return h.OnCurrentTrackDuration != nil }
func testCurrentTrackMetaData(h *upnptype.MediaHook) bool { return h.OnCurrentTrackMetaData != nil }
func testMute(h *upnptype.MediaHook) bool                 { return h.OnMute != nil }
func testVolume(h *upnptype.MediaHook) bool               { return h.OnVolume != nil }
func testCurrentTime(h *upnptype.MediaHook) bool          { return h.OnCurrentTime != nil }

func testSetVolumeDelta(h *upnptype.MediaHook) bool   { return h.OnSetVolumeDelta != nil }
func testSetSeekDelta(h *upnptype.MediaHook) bool     { return h.OnSetSeekDelta != nil }
func testRendererFound(h *upnptype.MediaHook) bool    { return h.OnRendererFound != nil }
func testServerFound(h *upnptype.MediaHook) bool      { return h.OnServerFound != nil }
func testRendererLost(h *upnptype.MediaHook) bool     { return h.OnRendererLost != nil }
func testServerLost(h *upnptype.MediaHook) bool       { return h.OnServerLost != nil }
func testRendererSelected(h *upnptype.MediaHook) bool { return h.OnRendererSelected != nil }
func testServerSelected(h *upnptype.MediaHook) bool   { return h.OnServerSelected != nil }

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
