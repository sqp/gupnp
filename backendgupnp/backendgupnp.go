// Package backendgupnp provides interaction with UPnP ressources on the network
// using the C gupnp backend.
package backendgupnp

import (
	"github.com/clbanning/mxj"

	"github.com/conformal/gotk3/glib"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/ternary"

	"github.com/sqp/gupnp/gupnp"
	"github.com/sqp/gupnp/upnptype"

	"encoding/xml"
	// "fmt"

	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// UPnP schemas names.
//
const (
	SchemaAVTransport      = "urn:schemas-upnp-org:service:AVTransport"
	SchemaRenderingControl = "urn:schemas-upnp-org:service:RenderingControl"
	SchemaContentDirectory = "urn:schemas-upnp-org:service:ContentDirectory"
	SchemaMediaRenderer    = "urn:schemas-upnp-org:device:MediaRenderer:1"
	SchemaMediaServer      = "urn:schemas-upnp-org:device:MediaServer:1"
)

// controlPointEvents defines discovery events connected to the C backend.
type controlPointEvents struct {
	onRendererFound func(*Renderer)
	onRendererLost  func(*Renderer)
	onServerFound   func(*Server)
	onServerLost    func(*Server)
}

// ControlPoint handles UPnP devices on the network.
//
type ControlPoint struct {
	events controlPointEvents

	dmrCP *gupnp.ControlPoint
	dmsCP *gupnp.ControlPoint
}

// NewControlPoint creates an UPnP devices manager.
//
func NewControlPoint() *ControlPoint {
	cp := &ControlPoint{}

	context := gupnp.ContextManagerCreate(0)
	_, e := context.Connect("context-available", cp.onContextAvailable)
	log.Err(e, "connect context")

	return cp
}

// SetEvents sets the manager callbacks.
//
func (cp *ControlPoint) SetEvents(events upnptype.ControlPointEvents) {
	// Reassert interfaces to real objects for the C backend.
	cp.events = controlPointEvents{
		onRendererFound: func(r *Renderer) { events.OnRendererFound(r) },
		onRendererLost:  func(r *Renderer) { events.OnRendererLost(r) },
		onServerFound:   func(s *Server) { events.OnServerFound(s) },
		onServerLost:    func(s *Server) { events.OnServerLost(s) },
	}
}

//
//-----------------------------------------------------[ DISCOVERY CALLBACKS ]--

// TODO: use glib to wrap objects, and merge with New func to remove controlPointEvents.
// also check provided callbacks are really used.
func (cp *ControlPoint) onContextAvailable(one *glib.Object, two *glib.Object) {
	cm := gupnp.WrapContextManager(one)
	context := gupnp.WrapContext(two)

	cp.dmrCP = gupnp.ControlPointNew(context, SchemaMediaRenderer)
	cp.dmsCP = gupnp.ControlPointNew(context, SchemaMediaServer)

	_, er := cp.dmrCP.Connect("device-proxy-available", cp.onDmrProxyAvailable, cp.events.onRendererFound)
	_, es := cp.dmsCP.Connect("device-proxy-available", cp.onDmsProxyAvailable, cp.events.onServerFound)
	_, erl := cp.dmrCP.Connect("device-proxy-unavailable", cp.onDmrProxyLost, cp.events.onRendererLost)
	_, esl := cp.dmsCP.Connect("device-proxy-unavailable", cp.onDmsProxyLost, cp.events.onServerLost)

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

	avTransport := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(SchemaAVTransport)}
	renderControl := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(SchemaRenderingControl)}

	udn := proxy.GetUdn()

	// if (udn != NULL)
	// if (G_UNLIKELY (cm != NULL))
	// if (av_transport != NULL)
	// if (rendering_control != NULL)

	r := &Renderer{
		udn:           udn,
		name:          proxy.GetFriendlyName(),
		proxy:         proxy,
		avTransport:   avTransport,
		renderControl: renderControl}

	cp.events.onRendererFound(r)

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

	contentDir := &gupnp.ServiceProxy{*proxy.DeviceInfo.GetService(SchemaContentDirectory)}

	s := &Server{
		udn:         proxy.GetUdn(),
		name:        proxy.GetFriendlyName(),
		proxy:       proxy,
		contentDir:  contentDir,
		id:          "0",
		isContainer: true}

	// Forward event.
	cp.events.onServerFound(s)

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
	cp.events.onRendererLost(&Renderer{proxy: gupnp.WrapDeviceProxy(two)})
}

func (cp *ControlPoint) onDmsProxyLost(one *glib.Object, two *glib.Object) {
	cp.events.onServerLost(&Server{proxy: gupnp.WrapDeviceProxy(two)})
}

//
//------------------------------------------------------------------[ SERVER ]--

type Server struct {
	udn   string
	name  string
	proxy *gupnp.DeviceProxy
	// mediaServer *gupnp.ServiceProxy
	contentDir *gupnp.ServiceProxy

	icon string // path to icon file on disk.

	id          string
	isContainer bool
	childCount  int
}

func (srv *Server) UDN() string         { return srv.udn }
func (srv *Server) Name() string        { return srv.name }
func (srv *Server) Icon() string        { return srv.icon }
func (srv *Server) SetIcon(icon string) { srv.icon = icon }

func (srv *Server) CompareProxy(utest upnptype.UDNer) bool {
	stest, ok := interface{}(utest).(*Server)
	if !ok {
		println("Renderer.CompareProxy: cast failed")
		return false
	}
	// return false
	return stest.proxy.Native() == srv.proxy.Native()
}

func (srv *Server) GetIconFile(filename string) string {
	url, _, _, _, _ := srv.proxy.GetIconUrl("", -1, 24, 24, true)
	return getIconFile(url, filename)
}

func (s *Server) Browse(req *upnptype.BrowseRequest) (browseResult *upnptype.BrowseResult, err error) {
	var didlXml string
	var numberReturned uint
	var totalMatches uint

	s.contentDir.SendAction("Browse",
		"ObjectID", req.ObjectID,
		"BrowseFlag", "BrowseDirectChildren",
		"Filter", "@childCount",
		"StartingIndex", uint(req.StartingIndex),
		"RequestedCount", uint(req.RequestCount),
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

	listObj := make([]upnptype.Object, len(item.Item))
	for k, v := range item.Item {
		listObj[k] = v.Object
	}

	return &upnptype.BrowseResult{
		NumberReturned: int32(numberReturned),
		TotalMatches:   int32(totalMatches),
		// UpdateID:       out.UpdateID,
		Container: item.Container,
		Item:      listObj,
	}, nil
}

func (s *Server) BrowseMetadata(container string, startingIndex, requestedCount uint) ([]upnptype.Container, []upnptype.Item, string) {

	var result string

	s.contentDir.SendAction("Browse",
		"ObjectID", container, //req.ObjectID,
		"BrowseFlag", "BrowseMetadata",
		"Filter", "*",
		"StartingIndex", startingIndex, // req.StartingIndex,
		"RequestedCount", requestedCount, //req.RequestCount,
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

type Renderer struct {
	udn           string
	name          string
	proxy         *gupnp.DeviceProxy
	avTransport   *gupnp.ServiceProxy
	renderControl *gupnp.ServiceProxy
	icon          string // path to icon file on disk.

	events upnptype.RendererEvents

	Current  int // current timer position in seconds.
	duration int // Total number of seconds of current track.
	state    upnptype.PlaybackState

	tick *time.Ticker // update timer loop when media is playing.
}

func (rend *Renderer) CompareProxy(utest upnptype.UDNer) bool {
	rtest, ok := interface{}(utest).(*Renderer)
	if !ok {
		println("Renderer.CompareProxy: cast failed")
		return false
	}
	return rtest.proxy.Native() == rend.proxy.Native()
}

func (rend *Renderer) GetIconFile(filename string) string {
	url, _, _, _, _ := rend.proxy.DeviceInfo.GetIconUrl("", -1, 24, 24, true)
	return getIconFile(url, filename)
}

func (rend *Renderer) UDN() string                      { return rend.udn }
func (rend *Renderer) Name() string                     { return rend.name }
func (rend *Renderer) Icon() string                     { return rend.icon }
func (rend *Renderer) SetIcon(icon string)              { rend.icon = icon }
func (rend *Renderer) Events() *upnptype.RendererEvents { return &rend.events }
func (rend *Renderer) Duration() int                    { return rend.duration }

//--------------------------------------------------------[ RENDERINGCONTROL ]--

func (rend *Renderer) GetMute(instanceId uint32, channel string) (bool, error) {
	var current bool
	e := rend.renderControl.SendAction("GetMute", "Channel", channel, nil, "CurrentMute", &current)
	return current, e
}

func (rend *Renderer) GetVolume(instanceId uint32, channel string) (uint16, error) {
	var current uint
	e := rend.renderControl.SendAction("GetVolume", "Channel", channel, nil, "CurrentVolume", &current)
	return uint16(current), e
}

func (rend *Renderer) SetVolume(instanceId uint32, channel string, vol uint16) error {
	return rend.renderControl.SendAction("SetVolume", "Channel", "Master", "DesiredVolume", uint(vol))
}

func (rend *Renderer) SetRelativeVolume(instanceId uint32, channel string, adjustment int32) (newVolume uint16, e error) {
	uintvol, e := rend.GetVolume(instanceId, channel)
	if e != nil {
		return 0, e
	}
	vol := ternary.Min(int(int32(uintvol)+adjustment), 100)
	vol = ternary.Max(vol, 0) // can't be negative.

	e = rend.SetVolume(instanceId, channel, uint16(vol))
	if e != nil {
		return 0, e
	}
	return uint16(vol), nil
}

func (rend *Renderer) SetMute(instanceId uint32, channel string, desiredMute bool) error {
	return rend.renderControl.SendAction("SetMute", "Channel", "Master", "DesiredMute", desiredMute)
}

//-------------------------------------------------------------[ AVTRANSPORT ]--

func (rend *Renderer) Play(instanceId uint32, speed string) error {
	return rend.avTransport.SendAction("Play", "Speed", speed)
}

func (rend *Renderer) Pause(instanceId uint32) error {
	return rend.avTransport.SendAction("Pause")
}

func (rend *Renderer) PlayPause(instanceId uint32, speed string) error {
	switch rend.state {
	case upnptype.PlaybackStatePaused, upnptype.PlaybackStateStopped:
		return rend.Play(instanceId, speed)

	case upnptype.PlaybackStatePlaying:
		return rend.Pause(instanceId)
	}
	return nil
}

func (rend *Renderer) Stop(instanceId uint32) error {
	return rend.avTransport.SendAction("Stop")
}

func (rend *Renderer) SetAVTransportURI(instanceId uint32, currentURI, currentURIMetaData string) error {
	rend.Stop(0)
	e := rend.avTransport.SendAction("SetAVTransportURI", "CurrentURI", currentURI, "CurrentURIMetaData", currentURIMetaData)
	if e != nil {
		return e
	}
	return rend.Play(0, "1")
}

// AddURIToQueue is TODO.
func (rend *Renderer) AddURIToQueue(instanceId uint32, req *upnptype.AddURIToQueueIn) (*upnptype.AddURIToQueueOut, error) {
	return nil, nil
}

// unit: ABS_TIME   (REL_TIME don't work on my TV)
func (rend *Renderer) Seek(instanceId uint32, unit, target string) error {
	e := rend.avTransport.SendAction("Seek", "Unit", unit, "Target", target)
	if e != nil {
		return e
	}
	rend.GetCurrentTime()
	rend.DisplayCurrentTime()
	return nil
}

func (rend *Renderer) GetCurrentTime() int {
	current := ""
	if !log.Err(rend.avTransport.SendAction("GetPositionInfo", nil, "AbsTime", &current), "AbsTime") {
		// log.Info("AbsTime", current)
		rend.Current = upnptype.TimeToSecond(current)
	}
	return rend.Current
}

func (rend *Renderer) DisplayCurrentTime() {
	rend.events.OnCurrentTime(rend, rend.Current, float64(rend.Current)/float64(rend.duration))
}

//
//-------------------------------------[ BACKEND INTERFACE COMPLIANCE - TODO ]--

func (rend *Renderer) AddMultipleURIsToQueue(instanceID uint32, req *upnptype.AddMultipleURIsToQueueIn) (*upnptype.AddMultipleURIsToQueueOut, error) {
	return nil, nil
}

func (rend *Renderer) GetCurrentTransportActions(instanceID uint32) ([]string, error) {
	return []string{}, nil
}

func (rend *Renderer) GetMediaInfo(instanceID uint32) (*upnptype.MediaInfo, error) {
	return &upnptype.MediaInfo{}, nil
}

func (rend *Renderer) GetTransportInfo(instanceID uint32) (*upnptype.TransportInfo, error) {
	return &upnptype.TransportInfo{}, nil
}

// TODO: need to complete it.
func (rend *Renderer) GetPositionInfo(instanceID uint32) (*upnptype.PositionInfo, error) {
	pos := &upnptype.PositionInfo{}
	rend.avTransport.SendAction("GetPositionInfo", nil, "TrackDuration", &pos.TrackDuration)
	// rend.avTransport.SendAction("GetMediaInfo", nil, "MediaDuration", &pos.TrackDuration) // works too, but may not point exactly to the same thing...

	return pos, nil
}

func (rend *Renderer) Next(instanceID uint32) error     { return nil }
func (rend *Renderer) Previous(instanceID uint32) error { return nil }
func (rend *Renderer) SetNextAVTransportURI(instanceID uint32, nextURI, nextURIMetaData string) error {
	return nil
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
			rend.state = upnptype.PlaybackStateFromName(v.(string))
			rend.events.OnTransportState(rend, rend.state)
			rend.tick = upnptype.Clock(rend, rend.tick, rend.state)

		case "CurrentTrackDuration":
			rend.duration = upnptype.TimeToSecond(v.(string))
			rend.events.OnCurrentTrackDuration(rend, upnptype.TimeToSecond(v.(string)))

		case "CurrentTrackMetaData":
			rend.events.OnCurrentTrackMetaData(rend, unmarshalDidl(v.(string)))

			// log.DETAIL(item)
			// log.Info(k, v)

			// case "CurrentTransportActions":

			// default:
			//  log.Info("AVT Unused", k, v)
		}
	}
}

func (rend *Renderer) onMsgRCS(str string) {
	// log.Info("onMsgRCS", str)
	// log.DEV("RCS")

	for k, v := range unmarshalMessage(str) {
		_ = v
		switch k {
		case "Mute":
			if i, e := strconv.Atoi(v.(string)); e == nil {
				rend.events.OnMute(rend, i == 1)
			}

		case "Volume":
			if i, e := strconv.Atoi(v.(string)); e == nil {
				rend.events.OnVolume(rend, uint(i))
			}

			// default:
			//  log.Info("RCS Unused", k, v)

		}

	}
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
//-------------------------------------------------------------[ XML PARSING ]--

func unmarshalDidl(str string) *upnptype.Item {
	item := didl{}
	xml.Unmarshal([]byte(str), &item)
	return &item.Item
}

type didlContainer struct {
	Container []upnptype.Container `xml:"container"`
	Item      []upnptype.Item      `xml:"item"`
}

type didl struct {
	Item upnptype.Item
}
