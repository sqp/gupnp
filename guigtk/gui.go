// Package guigtk provides a simple GTK3 GUI to interact with UPnP media players.
//
package guigtk

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/widgets/common"
	"github.com/sqp/godock/widgets/gtk/buildhelp"

	"github.com/sqp/godock/libs/log"

	"github.com/sqp/gupnp/upnptype"
)

// Rows for renderer and server comboboxes.
const (
	RowIcon = iota
	RowText
	RowRenderer // Object
	RowUDN      // ID
	RowSeriesID
	RowVisible
)

// Window settings.
const (
	WindowWidth  = 250
	WindowHeight = 400
)

var (
	// WindowTitle defines the title of the window.
	WindowTitle = "TVPlay"
)

// NewGui creates a window with a TVGui widget.
//
func NewGui(control upnptype.MediaControl) (*TVGui, *gtk.Window) {
	gui := NewTVGui(control)
	if gui == nil {
		return nil, nil
	}
	gui.Connect("destroy", func() {
		gui.Destroy()
		gui.DisconnectControl()
		gui = nil
	})

	window, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	window.SetDefaultSize(WindowWidth, WindowHeight)
	window.Add(gui)

	window.SetTitle(WindowTitle)
	window.SetWMClass(WindowTitle, WindowTitle)
	window.ShowAll()

	return gui, window
}

// TVGui is a media renderer selector widget using ComboBox.
//
type TVGui struct {
	gtk.Box // Container is first level. Act as (at least) a GtkWidget.

	// private widgets.
	model       *gtk.ListStore
	serverModel *gtk.ListStore
	filesModel  *gtk.TreeStore

	renderer *gtk.ComboBox
	server   *gtk.ComboBox
	files    *gtk.TreeView

	play     *gtk.Button
	stop     *gtk.Button
	backward *gtk.Button
	forward  *gtk.Button

	seekScale *gtk.Scale
	seekAdj   *gtk.Adjustment
	muted     *gtk.CheckButton
	volume    *gtk.Adjustment

	abstime  *gtk.Label
	duration *gtk.Label
	title    *gtk.Label

	callRenderer glib.SignalHandle // renderer changed callback reference.
	callServer   glib.SignalHandle // server changed callback reference.
	callSeekAdj  glib.SignalHandle // seekAdj value-changed callback reference.
	callVolume   glib.SignalHandle // Volume value-changed callback reference.
	callMuted    glib.SignalHandle // Muted changed callback reference.

	rendererIters map[string]*gtk.TreeIter
	serverIters   map[string]*gtk.TreeIter

	control upnptype.MediaControl
}

// NewTVGui creates a new TVGui widget.
//
// Parameters:
//   control   *MediaControl     The UPnP media control point interface.
//
func NewTVGui(control upnptype.MediaControl) *TVGui {

	builder := buildhelp.New()
	builder.AddFromString(string(guigtkXML()))
	// builder.AddFromFile("src/guigtk.xml")

	box := builder.GetBox("box")
	if box == nil || control == nil {
		log.Info("nobox")
		return nil
	}

	gui := &TVGui{
		Box:         *box,
		model:       builder.GetListStore("rendererModel"),
		serverModel: builder.GetListStore("serverModel"),
		filesModel:  builder.GetTreeStore("filesModel"),

		renderer: builder.GetComboBox("renderer"),
		server:   builder.GetComboBox("server"),
		files:    builder.GetTreeView("files"),

		play:     builder.GetButton("play"),
		stop:     builder.GetButton("stop"),
		backward: builder.GetButton("backward"),
		forward:  builder.GetButton("forward"),

		muted:     builder.GetCheckButton("muted"),
		seekScale: builder.GetScale("seekScale"),
		seekAdj:   builder.GetAdjustment("seekAdj"),
		volume:    builder.GetAdjustment("volumeAdj"),

		abstime:  builder.GetLabel("abstime"),
		duration: builder.GetLabel("duration"),
		title:    builder.GetLabel("title"),

		rendererIters: make(map[string]*gtk.TreeIter),
		serverIters:   make(map[string]*gtk.TreeIter),

		control: control,
	}

	if len(builder.Errors) > 0 {
		for _, e := range builder.Errors {
			log.DEV("build tvgui", e)
		}
		return nil
	}

	// Connect events.
	gui.callRenderer, _ = gui.renderer.Connect("changed", gui.onRendererChanged)
	gui.callServer, _ = gui.server.Connect("changed", gui.onServerChanged)
	gui.files.Connect("row-activated", gui.onFilesSelected)

	gui.callVolume, _ = gui.volume.Connect("value-changed", func() { gui.onVolumeSelected(gui.volume.GetValue()) })
	gui.callSeekAdj, _ = gui.seekAdj.Connect("value-changed", func() { gui.control.SeekPercent(gui.seekAdj.GetValue()) })
	gui.callMuted, _ = gui.muted.Connect("toggled", func() { gui.control.Action(upnptype.ActionToggleMute) })
	gui.play.Connect("clicked", func() { gui.control.Action(upnptype.ActionPlayPause) })
	gui.stop.Connect("clicked", func() { gui.control.Action(upnptype.ActionStop) })

	gui.backward.Connect("clicked", func() { gui.control.Action(upnptype.ActionSeekBackward) })
	gui.forward.Connect("clicked", func() { gui.control.Action(upnptype.ActionSeekForward) })

	gui.ConnectControl()

	return gui
}

// Load fills renderers and servers combo boxes.
//
func (gui *TVGui) Load() {
	gui.SetPlaybackState(upnptype.PlaybackStateUnknown)

	for _, rend := range gui.control.Renderers() {
		gui.AddRenderer(rend)
	}

	for _, srv := range gui.control.Servers() {
		gui.AddServer(srv)
	}
}

//
//-------------------------------------------------[ CALLBACKS CONTROL POINT ]--

// ConnectControl connects media callbacks.
//
func (gui *TVGui) ConnectControl() {
	hook := gui.control.SubscribeHook("gui")
	hook.OnRendererFound = gui.AddRenderer
	hook.OnServerFound = gui.AddServer
	hook.OnRendererSelected = gui.SetRenderer
	hook.OnServerSelected = gui.SetServer
	hook.OnRendererLost = gui.RemoveRenderer
	hook.OnServerLost = gui.RemoveServer

	hook.OnTransportState = func(r upnptype.Renderer, state upnptype.PlaybackState) { gui.SetPlaybackState(state) }
	hook.OnCurrentTrackDuration = func(r upnptype.Renderer, dur int) { gui.SetDuration(upnptype.TimeToString(dur)) }
	hook.OnCurrentTrackMetaData = func(r upnptype.Renderer, item *upnptype.Item) { gui.SetTitle(item.Title) }
	hook.OnMute = func(r upnptype.Renderer, muted bool) { gui.SetMuted(muted) }
	hook.OnVolume = func(r upnptype.Renderer, vol uint) { gui.SetVolume(int(vol)) }
	hook.OnCurrentTime = func(r upnptype.Renderer, secs int, f float64) { gui.SetCurrentTime(secs, f*100) }
	hook.OnSetVolumeDelta = func(delta int) { gui.SetVolumeDelta(delta) }
	// hook.OnSetSeekDelta = func(delta int) { gui.SetSeekDelta(delta) }
}

// DisconnectControl removes media callbacks.
//
func (gui *TVGui) DisconnectControl() {
	gui.control.UnsubscribeHook("gui")
}

// AddRenderer adds a media renderer to the gui.
//
func (gui *TVGui) AddRenderer(rend upnptype.Renderer) {
	glib.IdleAdd(func() {
		iter := gui.model.Append()
		gui.rendererIters[rend.UDN()] = iter

		gui.model.SetValue(iter, RowText, rend.Name())
		gui.model.SetValue(iter, RowUDN, rend.UDN())

		if rend.Icon() != "" {
			if pix, e := common.PixbufAtSize(rend.Icon(), 24, 24); !log.Err(e, "pixbuf icon") {
				gui.model.SetValue(iter, RowIcon, pix)
			}
		}
	})
}

// AddServer adds a media server to the gui.
//
func (gui *TVGui) AddServer(srv upnptype.Server) {
	glib.IdleAdd(func() {
		iter := gui.serverModel.Append()
		gui.serverIters[srv.UDN()] = iter

		gui.serverModel.SetValue(iter, RowText, srv.Name())
		gui.serverModel.SetValue(iter, RowUDN, srv.UDN())

		if srv.Icon() != "" {
			if pix, e := common.PixbufAtSize(srv.Icon(), 24, 24); !log.Err(e, "pixbuf icon") {
				gui.serverModel.SetValue(iter, RowIcon, pix)
			}
		}
	})
}

// RemoveRenderer removes a media renderer from the gui.
//
func (gui *TVGui) RemoveRenderer(rend upnptype.Renderer) {
	glib.IdleAdd(func() {
		gui.model.Remove(gui.rendererIters[rend.UDN()])
	})
	delete(gui.rendererIters, rend.UDN())
}

// RemoveServer removes a media server from the gui.
//
func (gui *TVGui) RemoveServer(srv upnptype.Server) {

	// log.Info("DEL", srv.Name())

	// if server was selected, clear list.
	// iter, e := gui.server.GetActiveIter()
	// if e == nil {
	// 	v, _ := gui.serverModel.GetValue(iter, RowUDN)
	// 	udn, _ := v.GetString()
	// 	if udn == srv.UDN() {
	// 		gui.filesModel.Clear()
	// 	}
	// }
	glib.IdleAdd(func() {
		gui.serverModel.Remove(gui.serverIters[srv.UDN()])
	})
	delete(gui.serverIters, srv.UDN())
}

// SetRenderer selects a renderer in the combo. Don't propagate event.
//
func (gui *TVGui) SetRenderer(rend upnptype.Renderer) {
	glib.IdleAdd(func() {
		var iter *gtk.TreeIter
		if rend != nil {
			if _, ok := gui.rendererIters[rend.UDN()]; ok {
				iter = gui.rendererIters[rend.UDN()]
				// if was selected, blank everything
			}
		}
		gui.renderer.HandlerBlock(gui.callRenderer)
		gui.renderer.SetActiveIter(iter)
		gui.renderer.HandlerUnblock(gui.callRenderer)
	})
}

// SetServer selects a server in the combo. Don't propagate event.
//
func (gui *TVGui) SetServer(srv upnptype.Server) {
	gui.filesModel.Clear()

	glib.IdleAdd(func() {
		var iter *gtk.TreeIter
		if srv != nil {
			if _, ok := gui.serverIters[srv.UDN()]; ok {
				iter = gui.serverIters[srv.UDN()]
				gui.browseDirectory("0", nil)
			}
		}
		gui.server.HandlerBlock(gui.callServer)
		gui.server.SetActiveIter(iter)
		gui.server.HandlerUnblock(gui.callServer)
	})
}

// func (gui *TVGui) SelectedRenderer() (*Renderer, error) {
// 	iter, e := gui.renderer.GetActiveIter()
// 	if e == nil {
// 		v, _ := gui.model.GetValue(iter, RowUDN)
// 		udn, _ := v.GetString()

// 		log.Info("Selected", udn)
// 		return gui.control.GetRenderer(udn), nil
// 	}
// 	return nil, e
// }

// SetVolumeDelta sets the volume interval for volume changes.
//
func (gui *TVGui) SetVolumeDelta(delta int) {
	glib.IdleAdd(func() {
		gui.volume.SetPageIncrement(float64(delta))
	})
}

// SetSeekDelta configures the default seek delta for user actions.
//
func (gui *TVGui) SetSeekDelta(delta int) {
	glib.IdleAdd(func() {
		gui.seekAdj.SetPageIncrement(float64(delta))
	})
}

//
//----------------------------------------------------------------[ SET INFO ]--

// SetCurrentTime sets the position of the timer slider.
//
func (gui *TVGui) SetCurrentTime(secs int, percent float64) {
	glib.IdleAdd(func() {
		gui.abstime.SetLabel(upnptype.TimeToString(secs))

		gui.seekAdj.HandlerBlock(gui.callSeekAdj)
		gui.seekAdj.SetValue(float64(percent))
		gui.seekAdj.HandlerUnblock(gui.callSeekAdj)
	})
}

// SetDuration sets the content of the duration label.
//
func (gui *TVGui) SetDuration(dur string) {
	glib.IdleAdd(func() {
		gui.duration.SetLabel(dur)
	})
}

// SetMuted sets the position of the muted button.
//
func (gui *TVGui) SetMuted(muted bool) {
	glib.IdleAdd(func() {
		gui.muted.HandlerBlock(gui.callMuted)
		gui.muted.SetActive(!muted)
		gui.muted.HandlerUnblock(gui.callMuted)
	})
}

// SetTitle sets the content of the title label.
//
func (gui *TVGui) SetTitle(title string) {
	glib.IdleAdd(func() {
		gui.title.SetLabel(title)
	})
}

// SetVolume sets the position of the volume slider.
//
func (gui *TVGui) SetVolume(vol int) {
	glib.IdleAdd(func() {
		gui.volume.HandlerBlock(gui.callVolume)
		gui.volume.SetValue(float64(vol))
		gui.volume.HandlerUnblock(gui.callVolume)
	})
}

// SetPlaybackState update controls according to the new state.
//
func (gui *TVGui) SetPlaybackState(state upnptype.PlaybackState) {
	glib.IdleAdd(func() {

		// println("state GUI", state)

		switch state {
		case upnptype.PlaybackStateUnknown, upnptype.PlaybackStateTransitioning: // not sure about the disable when unknown. could prevent from returning to a good state.
			gui.setControlsActive(false)
			gui.seekScale.SetSensitive(false)
			gui.backward.SetSensitive(false)
			gui.forward.SetSensitive(false)

		case upnptype.PlaybackStatePaused, upnptype.PlaybackStateStopped:
			imgplay, _ := gtk.ImageNewFromIconName("media-playback-start", gtk.ICON_SIZE_SMALL_TOOLBAR)
			gui.play.SetImage(imgplay)
			gui.setControlsActive(true)
			gui.seekScale.SetSensitive(false)
			gui.backward.SetSensitive(false)
			gui.forward.SetSensitive(false)

		case upnptype.PlaybackStatePlaying:
			imgplay, _ := gtk.ImageNewFromIconName("media-playback-pause", gtk.ICON_SIZE_SMALL_TOOLBAR)
			gui.play.SetImage(imgplay)
			gui.setControlsActive(true)
			gui.seekScale.SetSensitive(true)
			gui.backward.SetSensitive(true)
			gui.forward.SetSensitive(true)
		}
	})
}

func (gui *TVGui) setControlsActive(active bool) {
	glib.IdleAdd(func() {
		gui.play.SetSensitive(active)
		gui.stop.SetSensitive(active)
	})
}

//
//-----------------------------------------------------------[ CALLBACKS GUI ]--

func (gui *TVGui) onRendererChanged() {
	if iter, e := gui.renderer.GetActiveIter(); e == nil {
		v, _ := gui.model.GetValue(iter, RowUDN)
		udn, _ := v.GetString()

		gui.control.SetRenderer(udn)

		// TODO: disable actions until we have a valid state.

		// need to pull PlaybackState , volume and others.
	}
}

func (gui *TVGui) onServerChanged() {
	if iter, e := gui.server.GetActiveIter(); e == nil {
		v, _ := gui.serverModel.GetValue(iter, RowUDN)
		udn, _ := v.GetString()

		gui.control.SetServer(udn)
	}
}

func (gui *TVGui) onFilesSelected() {
	filesSelection, _ := gui.files.GetSelection()
	_, iter, ok := filesSelection.GetSelected()
	if ok {
		v, _ := gui.filesModel.GetValue(iter, RowUDN)
		ID, _ := v.GetString()
		gui.control.BrowseMetadata(ID, 0)
	}
}

func (gui *TVGui) onVolumeSelected(value float64) {
	if gui.control.RendererExists() {
		gui.control.Renderer().SetVolume(0, upnptype.ChannelMaster, uint16(value))
	}
}

//
//---------------------------------------------------------------[  ]--

func (gui *TVGui) browseDirectory(container string, parent *gtk.TreeIter) {

	containers, items, _, _ := gui.control.Browse(container, 0)

	for _, item := range containers {
		iter := gui.filesModel.Append(parent)
		gui.filesModel.SetValue(iter, RowText, item.Title)
		gui.filesModel.SetValue(iter, RowRenderer, item)
		gui.filesModel.SetValue(iter, RowUDN, item.ID)

		// log.Info(item.ID, item.Title)
		// 	go func() {
		if item.ChildCount > 0 {
			gui.browseDirectory(item.ID, iter)
		}
		// 	}()

		// gtk_tree_store_insert_with_values
		//                 (GTK_TREE_STORE (model),
		//                  &device_iter, parent_iter, -1,
		//                  0, get_icon_by_id (ICON_DEVICE),
		//                  1, friendly_name,
		//                  2, info,
		//                  3, content_dir,
		//                  4, "0",
		//                  5, TRUE,
		//                  -1);

		// log.Info(item.Title)
		// DEB(item)
	}

	for _, item := range items {
		iter := gui.filesModel.Append(parent)
		gui.filesModel.SetValue(iter, RowText, item.Title)
		gui.filesModel.SetValue(iter, RowRenderer, item)
		gui.filesModel.SetValue(iter, RowUDN, item.ID)

		// gtk_tree_store_insert_with_values
		//                 (GTK_TREE_STORE (model),
		//                  &device_iter, parent_iter, -1,
		//                  0, get_icon_by_id (ICON_DEVICE),
		//                  1, friendly_name,
		//                  2, info,
		//                  3, content_dir,
		//                  4, "0",
		//                  5, TRUE,
		//                  -1);

		// log.Info(item.Title)
		// DEB(item)
	}

	// for _, cont := range containers {
	// 	log.DEV(cont.Title, cont.ChildCount)
	// 	// if cont.ID == "64$2F" {
	// 	// 	it, _, _ := cp.server.Browse(cont.ID, 0, requestedCount)
	// 	// 	for _, cont := range it {
	// 	// 		log.DEV(cont.Title, cont.ChildCount)
	// 	// 	}
	// 	// }
	// }

}

// GtkTreeIter parent_iter;
// const char *id = NULL;
// const char *parent_id = NULL;
// const char *title = NULL;
// gboolean    is_container;
// gint        child_count;
// GdkPixbuf  *icon;
// gint        position;

// id = gupnp_didl_lite_object_get_id (object);
// title = gupnp_didl_lite_object_get_title (object);
// parent_id = gupnp_didl_lite_object_get_parent_id (object);
// if (id == NULL || title == NULL || parent_id == NULL)
//         return;

// is_container = GUPNP_IS_DIDL_LITE_CONTAINER (object);

// if (is_container) {
//         GUPnPDIDLLiteContainer *container;
//         position = 0;
//         icon = get_icon_by_id (ICON_CONTAINER);

//         container = GUPNP_DIDL_LITE_CONTAINER (object);
//         child_count = gupnp_didl_lite_container_get_child_count
//                                                         (container);
// } else {
//         position = -1;
//         child_count = 0;
//         icon = get_item_icon (object);
// }

// /* Check if we browsed the root container. */
// if (strcmp (browse_data->id, "0") == 0) {
//         parent_iter = *server_iter;
// } else if (!find_row (model,
//                       server_iter,
//                       &parent_iter,
//                       compare_container,
//                       (gpointer) parent_id,
//                       TRUE))
//         return;

// gtk_tree_store_insert_with_values (GTK_TREE_STORE (model),
//                                    NULL, &parent_iter, position,
//                                    0, icon,
//                                    1, title,
//                                    3, browse_data->content_dir,
//                                    4, id,
//                                    5, is_container,
//                                    6, child_count,
//                                    -1);
