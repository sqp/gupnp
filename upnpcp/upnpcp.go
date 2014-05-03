package upnpcp

import (
	"github.com/clbanning/mxj"
	"github.com/conformal/gotk3/glib"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/ternary"

	"github.com/sqp/gupnp"

	"encoding/xml"
	"fmt"

	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	CONNECTION_MANAGER = "urn:schemas-upnp-org:service:ConnectionManager"
	AV_TRANSPORT       = "urn:schemas-upnp-org:service:AVTransport"
	RENDERING_CONTROL  = "urn:schemas-upnp-org:service:RenderingControl"
	CONTENT_DIR        = "urn:schemas-upnp-org:service:ContentDirectory"
	MEDIA_RENDERER     = "urn:schemas-upnp-org:device:MediaRenderer:1"
	MEDIA_SERVER       = "urn:schemas-upnp-org:device:MediaServer:1"

	MAX_BROWSE = 64
)

type ControlPointEvents struct {
	OnRendererFound func(*Renderer)
	OnRendererLost  func(*Renderer)
	OnServerFound   func(*Server)
	OnServerLost    func(*Server)
}

type ControlPoint struct {
	Events ControlPointEvents

	dmrCP *gupnp.ControlPoint
	dmsCP *gupnp.ControlPoint
}

func NewControlPoint(ev ControlPointEvents) *ControlPoint {
	cp := &ControlPoint{Events: ev}

	context := gupnp.ContextManagerCreate(0)
	_, e := context.Connect("context-available", cp.onContextAvailable)
	log.Err(e, "connect context")

	return cp
}

//
//-----------------------------------------------------[ DISCOVERY CALLBACKS ]--

func (cp *ControlPoint) onContextAvailable(one *glib.Object, two *glib.Object) {
	cm := gupnp.WrapContextManager(one)
	context := gupnp.WrapContext(two)

	cp.dmrCP = gupnp.ControlPointNew(context, MEDIA_RENDERER)
	cp.dmsCP = gupnp.ControlPointNew(context, MEDIA_SERVER)

	_, er := cp.dmrCP.Connect("device-proxy-available", cp.onDmrProxyAvailable, cp.Events.OnRendererFound)
	_, es := cp.dmsCP.Connect("device-proxy-available", cp.onDmsProxyAvailable, cp.Events.OnServerFound)
	_, erl := cp.dmrCP.Connect("device-proxy-unavailable", cp.onDmrProxyLost, cp.Events.OnRendererLost)
	_, esl := cp.dmsCP.Connect("device-proxy-unavailable", cp.onDmsProxyLost, cp.Events.OnServerLost)

	log.Err(er, "connect device-proxy")
	log.Err(es, "connect device-proxy")
	log.Err(erl, "connect device-proxy")
	log.Err(esl, "connect device-proxy")

	cp.dmrCP.SSDPResourceBrowser.SetActive(true)
	cp.dmsCP.SSDPResourceBrowser.SetActive(true)

	// Let context manager take care of the control point life cycle
	cm.ManageControlPoint(cp.dmrCP)
	cm.ManageControlPoint(cp.dmsCP)
}

// Rescan network for servers and renderers.
//
func (cp *ControlPoint) Rescan() {
	cp.dmrCP.SSDPResourceBrowser.Rescan()
	cp.dmsCP.SSDPResourceBrowser.Rescan()
}

func (cp *ControlPoint) onDmrProxyAvailable(one *glib.Object, two *glib.Object, onMediaRendererFound func(*Renderer)) {
	proxy := gupnp.WrapDeviceProxy(two)

	avTransport := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(AV_TRANSPORT)}
	renderControl := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(RENDERING_CONTROL)}

	udn := proxy.GetUdn()

	// if (udn != NULL)
	// if (G_UNLIKELY (cm != NULL))
	// if (av_transport != NULL)
	// if (rendering_control != NULL)

	r := &Renderer{
		Udn:           udn,
		Name:          proxy.GetFriendlyName(),
		proxy:         proxy,
		avTransport:   avTransport,
		renderControl: renderControl}

	cp.Events.OnRendererFound(r)

	avTransport.AddNotify("LastChange", glib.TYPE_STRING, r.onMsgAVT)
	renderControl.AddNotify("LastChange", glib.TYPE_STRING, r.onMsgRCS)

	avTransport.SetSubscribed(true)
	renderControl.SetSubscribed(true)

	// state_name := ""
	// avTransport.SendAction("GetTransportInfo", nil, "CurrentTransportState", &state_name)
	// log.Info("CurrentTransportState", state_name)

	// duration := ""
	// avTransport.SendAction("GetMediaInfo", nil, "MediaDuration", &duration)
	// log.Info("MediaDuration", duration)

	// GetVolume

	// cm = get_connection_manager (proxy);
	//gupnp_service_proxy_begin_action (g_object_ref (cm), "GetProtocolInfo", get_protocol_info_cb, NULL, NULL);

	// g_object_unref (cm);
}

func (cp *ControlPoint) onDmsProxyAvailable(one *glib.Object, two *glib.Object, onMediaServerFound func(*Server)) {
	proxy := gupnp.WrapDeviceProxy(two)

	contentDir := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(CONTENT_DIR)}

	s := &Server{
		Udn:         proxy.GetUdn(),
		Name:        proxy.GetFriendlyName(),
		proxy:       proxy,
		contentDir:  contentDir,
		id:          "0",
		isContainer: true}

	// Forward event.
	cp.Events.OnServerFound(s)

	// if (G_LIKELY (friendly_name != NULL && content_dir != NULL)) {

	//         /* Append the embedded devices */
	//         child = gupnp_device_info_list_devices (info);
	//         while (child) {
	//                 append_media_server (GUPNP_DEVICE_PROXY (child->data),
	//                                      model,
	//                                      &device_iter);
	//                 g_object_unref (child->data);
	//                 child = g_list_delete_link (child, child);
	//         }

	//         gupnp_service_proxy_add_notify (content_dir,
	//                                         "ContainerUpdateIDs",
	//                                         G_TYPE_STRING,
	//                                         on_container_update_ids,
	//                                         NULL);
	//         gupnp_service_proxy_set_subscribed (content_dir, TRUE);
	//         g_hash_table_insert (initial_notify, content_dir, content_dir);

	// }

	// if (!expanded) {
	//         gtk_tree_view_expand_all (GTK_TREE_VIEW (treeview));
	//         expanded = TRUE;
	// }
}

func (cp *ControlPoint) onDmrProxyLost(one *glib.Object, two *glib.Object) {
	cp.Events.OnRendererLost(&Renderer{proxy: gupnp.WrapDeviceProxy(two)})
}

func (cp *ControlPoint) onDmsProxyLost(one *glib.Object, two *glib.Object) {
	cp.Events.OnServerLost(&Server{proxy: gupnp.WrapDeviceProxy(two)})
}

//
//------------------------------------------------------------------[ SERVER ]--

type Server struct {
	Udn   string
	Name  string
	proxy *gupnp.DeviceProxy
	// mediaServer *gupnp.ServiceProxy
	contentDir *gupnp.ServiceProxy

	Icon string // path to icon file on disk.

	id          string
	isContainer bool
	childCount  int
}

func (srv *Server) CompareProxy(stest *Server) bool {
	return stest.proxy.Native() == srv.proxy.Native()
}

func (srv *Server) GetIconFile(filename string) string {
	url, _, _, _, _ := srv.proxy.GetIconUrl("", -1, 24, 24, true)
	return getIconFile(url, filename)
}

func (s *Server) Browse(container string, startingIndex, requestedCount uint) ([]Container, []Item, uint, uint) {

	var didlXml string
	var numberReturned uint
	var totalMatches uint

	s.contentDir.SendAction("Browse",
		"ObjectID", container,
		"BrowseFlag", "BrowseDirectChildren",
		"Filter", "@childCount",
		"StartingIndex", startingIndex,
		"RequestedCount", requestedCount,
		"SortCriteria", "",
		nil, // separator between in and out args.
		"Result", &didlXml,
		"NumberReturned", &numberReturned,
		"TotalMatches", &totalMatches)

	item := didlContainer{}
	xml.Unmarshal([]byte(didlXml), &item)

	if uint(len(item.Container)+len(item.Item)) != numberReturned {
		log.DEV("browse count problem. Said", numberReturned, "found, but parsed", len(item.Container), "and", len(item.Item))
	}

	return item.Container, item.Item, numberReturned, totalMatches
}

func (s *Server) BrowseMetadata(container string, startingIndex, requestedCount uint) ([]Container, []Item, string) {

	var result string

	s.contentDir.SendAction("Browse",
		"ObjectID", container,
		"BrowseFlag", "BrowseMetadata",
		"Filter", "*",
		"StartingIndex", startingIndex,
		"RequestedCount", requestedCount,
		"SortCriteria", "",
		nil,
		"Result", &result)

	// log.DEV("BrowseMetadata", result)
	// log.DEV("browsenew", numberReturned, totalMatches)

	item := didlContainer{}
	xml.Unmarshal([]byte(result), &item)

	return item.Container, item.Item, result
}

//
//---------------------------------------------------------------[ RENDERERS ]--

type RendererEvents struct {
	OnTransportState       func(*Renderer, PlaybackState)
	OnCurrentTrackDuration func(*Renderer, int)
	OnCurrentTrackMetaData func(*Renderer, *Item)

	OnMute   func(*Renderer, bool)
	OnVolume func(*Renderer, uint)

	OnCurrentTime func(r *Renderer, secs int, percent float64)
}

type Renderers map[string]*Renderer

type Renderer struct {
	Udn           string
	Name          string
	proxy         *gupnp.DeviceProxy
	avTransport   *gupnp.ServiceProxy
	renderControl *gupnp.ServiceProxy
	Icon          string // path to icon file on disk.

	Events RendererEvents

	Current  int // current timer position in seconds.
	Duration int // Total number of seconds of current track.
	State    PlaybackState

	tick *time.Ticker // update timer loop when media is playing.
}

func (rend *Renderer) CompareProxy(rtest *Renderer) bool {
	return rtest.proxy.Native() == rend.proxy.Native()
}

func (rend *Renderer) GetIconFile(filename string) string {
	url, _, _, _, _ := rend.proxy.DeviceInfo.GetIconUrl("", -1, 24, 24, true)
	return getIconFile(url, filename)
}

//-----------------------------------------------------------[ RENDERCONTROL ]--

func (rend *Renderer) GetMuted() bool {
	var current bool
	e := rend.renderControl.SendAction("GetMute", "Channel", "Master", nil, "CurrentMute", &current)
	log.Err(e, "GetMuted")
	return current
}

func (rend *Renderer) GetVolume() uint {
	var current uint
	e := rend.renderControl.SendAction("GetVolume", "Channel", "Master", nil, "CurrentVolume", &current)
	log.Err(e, "GetVolume")
	return current
}

func (rend *Renderer) SetVolume(vol uint) {
	e := rend.renderControl.SendAction("SetVolume", "Channel", "Master", "DesiredVolume", vol)
	log.Err(e, "SetVolume")
}

func (rend *Renderer) SetVolumeDelta(delta int) {
	vol := int(rend.GetVolume()) + delta

	vol = ternary.Min(vol, 100)
	vol = ternary.Max(vol, 0) // can't be negative.

	rend.SetVolume(uint(vol))
}

func (rend *Renderer) ToggleMute() {
	current := rend.GetMuted()
	e := rend.renderControl.SendAction("SetMute", "Channel", "Master", "DesiredMute", !current)
	log.Err(e, "ToggleMute")
}

//-------------------------------------------------------------[ AVTRANSPORT ]--

func (rend *Renderer) Play() {
	e := rend.avTransport.SendAction("Play", "Speed", "1")
	log.Err(e, "ToggleMute")
}

func (rend *Renderer) Pause() {
	e := rend.avTransport.SendAction("Pause")
	log.Err(e, "Pause")
}

func (rend *Renderer) PlayPause() {
	switch rend.State {
	case PlaybackStatePaused, PlaybackStateStopped:
		rend.Play()

	case PlaybackStatePlaying:
		rend.Pause()
	}
}

func (rend *Renderer) Stop() {
	rend.avTransport.SendAction("Stop")
}

func (rend *Renderer) SetURL(uri, metadata string) {
	rend.Stop()
	rend.avTransport.SendAction("SetAVTransportURI", "CurrentURI", uri, "CurrentURIMetaData", metadata)
	rend.Play()
}

func (rend *Renderer) SetSeek(time string) {
	if !log.Err(rend.avTransport.SendAction("Seek", "Unit", "ABS_TIME", "Target", time), "Seek") {
		rend.GetCurrentTime()
		rend.DisplayCurrentTime()
	}
}

func (rend *Renderer) GetCurrentTime() int {
	current := ""
	if !log.Err(rend.avTransport.SendAction("GetPositionInfo", nil, "AbsTime", &current), "AbsTime") {
		// log.Info("AbsTime", current)
		rend.Current = TimeToSecond(current)
	}
	return rend.Current
}

func (rend *Renderer) DisplayCurrentTime() {
	rend.Events.OnCurrentTime(rend, rend.Current, float64(rend.Current)/float64(rend.Duration))
}

//
//-------------------------------------------------------[ RENDERER MESSAGES ]--

// Parse received message from device.
//
func unmarshalMessage(x string) map[string]interface{} {
	m, err := mxj.NewMapXml([]byte(x))
	log.Err(err, "parse xml message")

	gg, e := m.ValuesForPath("Event.InstanceID")
	if e == nil && len(gg) > 0 {

		values := make(map[string]interface{})

		for k, v := range gg[0].(map[string]interface{}) {
			if k != "-val" {
				values[k] = v.(map[string]interface{})["-val"]
			}
		}
		return values
	}
	return nil
}

func (rend *Renderer) onMsgAVT(str string) {
	// log.Info("onMsgAVT", str)
	// log.DEV("AVT")

	for k, v := range unmarshalMessage(str) {
		_ = v
		switch k {
		case "TransportState":
			rend.State = PlaybackStateFromName(v.(string))
			rend.Events.OnTransportState(rend, rend.State)
			rend.clock()

		case "CurrentTrackDuration":
			rend.Duration = TimeToSecond(v.(string))
			rend.Events.OnCurrentTrackDuration(rend, TimeToSecond(v.(string)))

		case "CurrentTrackMetaData":
			rend.Events.OnCurrentTrackMetaData(rend, UnmarshalDidl(v.(string)))

			// log.DETAIL(item)
			// log.Info(k, v)

			// case "CurrentTransportActions":

			// default:
			//  log.Info("AVT Unused", k, v)
		}
	}
}

// hook.OnMute = func(r *upnpcp.Renderer, muted bool) { media.gui.SetTitle(item.Title) }
// hook.OnVolume = func(r *upnpcp.Renderer, vol int) { media.gui.SetTitle(item.Title) }

func (rend *Renderer) onMsgRCS(str string) {
	// log.Info("onMsgRCS", str)
	// log.DEV("RCS")

	for k, v := range unmarshalMessage(str) {
		_ = v
		switch k {
		case "Mute":
			if i, e := strconv.Atoi(v.(string)); e == nil {
				rend.Events.OnMute(rend, i == 1)
			}

		case "Volume":
			if i, e := strconv.Atoi(v.(string)); e == nil {
				rend.Events.OnVolume(rend, uint(i))
			}

			// default:
			//  log.Info("RCS Unused", k, v)

		}

	}
}

//
//--------------------------------------------------------------------[ TIME ]--

func (rend *Renderer) clock() {

	rend.GetCurrentTime()
	rend.DisplayCurrentTime()

	if rend.tick != nil {
		rend.tick.Stop()
	}

	switch rend.State {
	case PlaybackStatePlaying:
		rend.tick = time.NewTicker(time.Second)
		go func() {
			for _ = range rend.tick.C {
				rend.GetCurrentTime()
				rend.DisplayCurrentTime()
			}
		}()
	}
}

func TimeToSecond(str string) int {
	var h, m, s int
	fmt.Sscanf(str, "%d:%d:%d", &h, &m, &s) // (n int, err error)
	// log.DEV("time", h, m, s)
	// secs = 60*60*h + 60*m + s
	return (h*60+m)*60 + s
}

//
//-------------------------------------------------------------------[ ICONS ]--

func getIconFile(url, filename string) string {
	file, e := os.Create(filename)
	if log.Err(e, "Create icon file") {
		return ""
	}
	defer file.Close()
	_, e = file.Write(download(url))
	if log.Err(e, "Write icon file") {
		return ""
	}
	return filename
}

func download(addr string) []byte {
	resp, err := http.Get(addr)
	if err != nil {
		log.Err(err, "http.Get")
		// handle error
		return []byte{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Err(err, "ioutil.ReadAll")
	return body
}

//
//-----------------------------------------------------------[ PLAYBACKSTATE ]--

type PlaybackState int

const (
	PlaybackStateUnknown PlaybackState = iota
	PlaybackStateTransitioning
	PlaybackStateStopped
	PlaybackStatePaused
	PlaybackStatePlaying
)

func PlaybackStateFromName(name string) PlaybackState {
	switch name {
	case "STOPPED":
		return PlaybackStateStopped

	case "PLAYING":
		return PlaybackStatePlaying

	case "PAUSED_PLAYBACK":
		return PlaybackStatePaused

	case "TRANSITIONING":
		return PlaybackStateTransitioning
	}

	return PlaybackStateUnknown
}

//
//-------------------------------------------------------[ DIDL-Lite PARSING ]--

func UnmarshalDidl(str string) *Item {
	item := didl{}
	xml.Unmarshal([]byte(str), &item)
	return &item.Item
}

type didlContainer struct {
	Container []Container `xml:"container"`
	Item      []Item      `xml:"item"`
}

type didl struct {
	Item Item
}

type Resource struct {
	// XMLName      xml.Name `xml:"res"`
	ProtocolInfo string `xml:"protocolInfo,attr"`
	URL          string `xml:",chardata"`
	Size         uint64 `xml:"size,attr,omitempty"`
	Bitrate      uint   `xml:"bitrate,attr,omitempty"`
	Duration     string `xml:"duration,attr,omitempty"`
	Resolution   string `xml:"resolution,attr,omitempty"`
}

type Container struct {
	Object
	XMLName    xml.Name `xml:"container"`
	ChildCount int      `xml:"childCount,attr"`
}

type Item struct {
	Object
	XMLName xml.Name   `xml:"item"`
	Res     []Resource `xml:"res"`
}

type Object struct {
	ID         string `xml:"id,attr"`
	ParentID   string `xml:"parentID,attr"`
	Restricted int    `xml:"restricted,attr"`  // indicates whether the object is modifiable
	Class      string `xml:"class"`            // upnp:
	Icon       string `xml:"icon,omitempty"`   // upnp:
	Title      string `xml:"title"`            // dc:
	Artist     string `xml:"artist,omitempty"` // upnp:
	Album      string `xml:"album,omitempty"`  // upnp:
	Genre      string `xml:"genre,omitempty"`  // upnp:
}
