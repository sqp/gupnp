package gupnp

/*
#include <libgupnp/gupnp-control-point.h>
#include <libgupnp-av/gupnp-av.h>
#include <libgssdp/gssdp-resource-browser.h>
#include <glib-2.0/glib.h>

static GUPnPContext*         toGUPnPContext(void *p)         { return (GUPNP_CONTEXT(p)); }
static GUPnPContextManager*  toGUPnPContextManager(void *p)  { return (GUPNP_CONTEXT_MANAGER(p)); }
static GUPnPControlPoint*    toGUPnPControlPoint(void *p)    { return (GUPNP_CONTROL_POINT(p)); }
static GUPnPDeviceInfo*      toGUPnPDeviceInfo(void *p)      { return (GUPNP_DEVICE_INFO(p)); }
static GUPnPDeviceProxy*     toGUPnPDeviceProxy(void *p)     { return (GUPNP_DEVICE_PROXY(p)); }
static GUPnPServiceProxy*    toGUPnPServiceProxy(void *p)    { return (GUPNP_SERVICE_PROXY(p)); }
static GSSDPResourceBrowser* toGSSDPResourceBrowser(void *p) { return (GSSDP_RESOURCE_BROWSER(p)); }
static GList*                toGlist(void* l)                { return (GList*)l; }
static gpointer              intToPointer(int i)             { return GINT_TO_POINTER(i); }

static gchar* error_get_message(GError *error) { return error->message; }



// golang functions declaration, prevents warning.
void onNotifyCallback(GValue*, int);

static void on_notify_callback (GUPnPServiceProxy *rendering_control, const char *variable_name, GValue *value, gpointer user_data) {
	onNotifyCallback(value, GPOINTER_TO_INT(user_data));
}

static gboolean service_proxy_add_notify (GUPnPServiceProxy *rendering_control, const char *variable_name, GType type, int callback_id) {
	return gupnp_service_proxy_add_notify(rendering_control, variable_name, type, on_notify_callback, GINT_TO_POINTER(callback_id));
}

*/
// #cgo pkg-config: glib-2.0 gupnp-1.0 gssdp-1.0 gupnp-av-1.0 gobject-introspection-1.0
import "C"

import (
	"github.com/conformal/gotk3/glib"

	"github.com/sqp/godock/libs/log"

	"errors"
	"runtime"
	"unsafe"

	"reflect"
)

/*
 * GUPnPContext
 */

// ContextManager is a representation of GUPnP's GUPnPContext.
type Context struct {
	*glib.Object
}

// Native() returns a pointer to the underlying GUPnPContext.
func (v *Context) Native() *C.GUPnPContext {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPContext(p)
}

func WrapContext(obj *glib.Object) *Context {
	return &Context{obj}
}

/*
 * GUPnPContextManager
 */

// ContextManager is a representation of GUPnP's GUPnPContextManager.
type ContextManager struct {
	*glib.Object
}

// Native() returns a pointer to the underlying GUPnPContextManager.
func (v *ContextManager) Native() *C.GUPnPContextManager {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPContextManager(p)
}

func WrapContextManager(obj *glib.Object) *ContextManager {
	return &ContextManager{obj}
}

/*
// ContextManagerNew is a wrapper around gtk_check_menu_item_new().
func ContextManagerNew(model ITreeModel, root *TreePath) (*ContextManager, error) {
	// cfont := C.CString("1")
	// defer C.free(unsafe.Pointer(cfont))
	// grr := C.gtk_tree_path_new_from_string((*C.gchar)(cfont))
	// c := C.gtk_tree_model_filter_new(model.toTreeModel(), grr)
	c := C.gtk_tree_model_filter_new(model.toTreeModel(), root.Native())
	if c == nil {
		return nil, nilPtrErr
	}
	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c))}
	s := WrapContextManager(obj)

	// s.TreeModel = model

	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return s, nil
}
*/

func ContextManagerCreate(port uint) *ContextManager {
	return ToContextManager(C.gupnp_context_manager_create(C.guint(port)))
	// c := C.gupnp_context_manager_create(C.guint(port))
	// if c == nil {
	// 	log.DEV("NIL")
	// 	return nil
	// }
	// obj := wrapObject(unsafe.Pointer(c))
	// s := WrapContextManager(obj)
	// return s
}

func ToContextManager(cm *C.GUPnPContextManager) *ContextManager {
	if cm == nil {
		return nil
	}
	return WrapContextManager(wrapObject(unsafe.Pointer(cm)))
}

// GetActive() is a wrapper around gtk_combo_box_get_active_iter().
func (v *ContextManager) ManageControlPoint(cp *ControlPoint) {
	C.gupnp_context_manager_manage_control_point(v.Native(), cp.Native())
}

/*
 * GSSDPResourceBrowser
 */

// SSDPResourceBrowser is a representation of GUPnP's GSSDPResourceBrowser.
type SSDPResourceBrowser struct {
	*glib.Object
}

// Native() returns a pointer to the underlying GSSDPResourceBrowser.
func (v *SSDPResourceBrowser) Native() *C.GSSDPResourceBrowser {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGSSDPResourceBrowser(p)
}

func (v *SSDPResourceBrowser) SetActive(active bool) {
	C.gssdp_resource_browser_set_active(v.Native(), gbool(active))
}

func (v *SSDPResourceBrowser) Rescan() {
	C.gssdp_resource_browser_rescan(v.Native())
}

/*
 * GUPnPControlPoint
 */

// ControlPoint is a representation of GUPnP's GUPnPControlPoint.
type ControlPoint struct {
	SSDPResourceBrowser
}

// Native() returns a pointer to the underlying GUPnPControlPoint.
func (v *ControlPoint) Native() *C.GUPnPControlPoint {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPControlPoint(p)
}

func wrapControlPoint(obj *glib.Object) *ControlPoint {
	return &ControlPoint{SSDPResourceBrowser{obj}}
}

func ControlPointNew(context *Context, target string) *ControlPoint {
	cstr := C.CString(target)
	defer C.free(unsafe.Pointer(cstr))
	c := C.gupnp_control_point_new(context.Native(), cstr)
	if c == nil {
		return nil
	}
	return wrapControlPoint(wrapObject(unsafe.Pointer(c)))
}

/*
 * GUPnPDeviceInfo
 */

type DeviceInfo struct {
	*glib.Object
}

// Native() returns a pointer to the underlying GUPnPDeviceProxy.
func (v *DeviceInfo) Native() *C.GUPnPDeviceInfo {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPDeviceInfo(p)
}

func WrapDeviceInfo(obj *glib.Object) *DeviceInfo {
	return &DeviceInfo{obj}
}

func toDeviceInfo(c *C.GUPnPDeviceInfo) *DeviceInfo {
	if c == nil {
		return nil
	}
	return WrapDeviceInfo(wrapObject(unsafe.Pointer(c)))
}

func (v *DeviceInfo) GetService(typ string) *ServiceInfo {
	cstr := C.CString(typ)
	defer C.free(unsafe.Pointer(cstr))
	return toServiceInfo(C.gupnp_device_info_get_service(v.Native(), cstr))
}

func (v *DeviceInfo) GetUdn() string {
	return C.GoString(C.gupnp_device_info_get_udn(v.Native()))
}

func (v *DeviceInfo) GetFriendlyName() string {
	cstr := C.gupnp_device_info_get_friendly_name(v.Native())
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func (v *DeviceInfo) ListDevices() []*DeviceInfo {
	c := C.gupnp_device_info_list_devices(v.Native())
	glist := glib.ListFromNative(unsafe.Pointer(c))
	// log.Info("len", C.g_list_length(c))
	devices := make([]*DeviceInfo, glist.Length())
	// The returned list should be g_list_free()'d and the elements should be g_object_unref()'d.
	for i := uint(0); i < uint(len(devices)); i++ {
		devices[i] = toDeviceInfo(glist.NthData(i).(*C.GUPnPDeviceInfo))
	}
	return devices
}

func (v *DeviceInfo) GetIconUrl(requestedMimeType string, requestedDepth, requestedWidth, requestedHeight int, preferBigger bool) (string, string, int, int, int) {
	var cMT *C.char = nil
	if requestedMimeType != "" {
		cMT := C.CString(requestedMimeType)
		defer C.free(unsafe.Pointer(cMT))
	}
	var depth, width, height C.int
	var cMimeType *C.char // := C.CString(typ)

	cUrl := C.gupnp_device_info_get_icon_url(v.Native(), cMT, C.int(requestedDepth), C.int(requestedWidth), C.int(requestedHeight),
		gbool(preferBigger), &cMimeType, &depth, &width, &height)
	defer C.free(unsafe.Pointer(cUrl))
	defer C.free(unsafe.Pointer(cMimeType))

	return C.GoString(cUrl), C.GoString(cMimeType), int(depth), int(width), int(height)
}

/*
 * GUPnPDeviceProxy
 */

// DeviceProxy is a representation of GUPnP's GUPnPDeviceProxy.
type DeviceProxy struct {
	DeviceInfo
}

// Native() returns a pointer to the underlying GUPnPDeviceProxy.
func (v *DeviceProxy) Native() *C.GUPnPDeviceProxy {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPDeviceProxy(p)
}

func WrapDeviceProxy(obj *glib.Object) *DeviceProxy {
	return &DeviceProxy{DeviceInfo{obj}}
}

// func toDeviceProxy(c *C.GUPnPDeviceProxy) *DeviceProxy {
// 	if c == nil {
// 		return nil
// 	}
// 	return WrapDeviceProxy(wrapObject(unsafe.Pointer(c)))
// }

// func DeviceProxyNew(context *Context, target string) *DeviceProxy {
// 	cstr := C.CString(target)
// 	defer C.free(unsafe.Pointer(cstr))
// 	c := C.gupnp_control_point_new(context.Native(), cstr)
// 	if c == nil {
// 		return nil
// 	}
// 	s := WrapDeviceProxy(wrapObject(unsafe.Pointer(c)))
// 	return s
// }

/*
 * GUPnPServiceInfo
 */

// ServiceInfo is a representation of GUPnP's GUPnPServiceInfo.
type ServiceInfo struct {
	*glib.Object
}

func WrapServiceInfo(obj *glib.Object) *ServiceInfo {
	return &ServiceInfo{obj}
}

func toServiceInfo(c *C.GUPnPServiceInfo) *ServiceInfo {
	if c == nil {
		return nil
	}
	return WrapServiceInfo(wrapObject(unsafe.Pointer(c)))
}

/*
 * GUPnPServiceProxy
 */

// ServiceProxy is a representation of GUPnP's GUPnPServiceProxy.
type ServiceProxy struct {
	ServiceInfo
}

// Native() returns a pointer to the underlying GUPnPServiceProxy.
func (v *ServiceProxy) Native() *C.GUPnPServiceProxy {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGUPnPServiceProxy(p)
}

//
//-----------------------------------------------------------[ NOTIFICATIONS ]--

var callbacksFunc = []func(string){}

func (v *ServiceProxy) AddNotify(variable string, typ glib.Type, callback func(string)) bool {
	callbackID := len(callbacksFunc)
	callbacksFunc = append(callbacksFunc, callback)

	cstr := C.CString(variable)
	defer C.free(unsafe.Pointer(cstr))
	C.service_proxy_add_notify(v.Native(), cstr, C.GType(typ), C.int(callbackID))
	return false
}

//export onNotifyCallback
func onNotifyCallback(cGValue *C.GValue, callbackID C.int) {
	gv := glib.ValueFromNative(unsafe.Pointer(cGValue))
	str, e := gv.GetString()
	if !log.Err(e, "onNotifyCallback get GValue") {
		callbacksFunc[int(callbackID)](str)
	}
}

//
func (v *ServiceProxy) SetSubscribed(subscribed bool) {
	C.gupnp_service_proxy_set_subscribed(v.Native(), gbool(subscribed))
}

// Note that outvalues will be changed. Default GValues will be freed and new ones seem to get allocated.
// It's up to you to free everything.
//
func (v *ServiceProxy) SendActionList(action string, innames, invalues, outnames, outtypes, outvalues *List) error {
	cAction := C.CString(action)
	defer C.free(unsafe.Pointer(cAction))

	var err *C.GError = nil
	res := C.gupnp_service_proxy_send_action_list(v.Native(), cAction, &err, innames.GList, invalues.GList, outnames.GList, outtypes.GList, &outvalues.GList)
	if res == 0 {
		defer C.g_error_free(err)
		return errors.New(C.GoString((*C.char)(C.error_get_message(err))))
	}

	return nil
}

//
//-----------------------------------------------------------------[ ACTIONS ]--

func (v *ServiceProxy) SendAction(action string, args ...interface{}) error {
	instance := []interface{}{"InstanceID", uint(0)}
	args = append(instance, args...)

	innames, invalues, argsOut := v.addArgumentsIn(args...)
	outnames, outtypes, outvalues := v.addArgumentsOut(argsOut)

	e := v.SendActionList(action, innames, invalues, outnames, outtypes, outvalues)

	for i := uint(0); i < outvalues.Length(); i++ {
		ret := outvalues.NthData(i).(C.gpointer)
		gv := glib.ValueFromNative(unsafe.Pointer(ret))
		goval, e := gv.GoValue()
		if e != nil {
			return errors.New("send action parse return: " + e.Error())
		}
		// if log.Err(e, "send action parse return") {
		// 	return e
		// }
		switch argsOut[2*i+1].(type) {
		case *bool:
			*argsOut[2*i+1].(*bool) = goval.(bool)

		case *string:
			*argsOut[2*i+1].(*string) = goval.(string)

		case *uint:
			*argsOut[2*i+1].(*uint) = goval.(uint)
		}
	}

	// TODO: need to free GLists and content.

	return e
}

func (v *ServiceProxy) addArgumentsIn(args ...interface{}) (names, values *List, argsOut []interface{}) {

	names = &List{}
	values = &List{}

	// var argsOut []interface{}

	if len(args) >= 2 {
		var gval *glib.Value
		for i := 0; i < len(args); i += 2 {

			if args[i] == nil { // Separator between in and out args.
				argsOut = args[i+1:]
				break
			}

			gstr := (*C.gchar)(C.CString(args[i].(string)))
			defer C.free(unsafe.Pointer(gstr))
			names = names.Append(unsafe.Pointer(C.g_strdup(gstr)))

			switch args[i+1].(type) {
			case bool:
				gval, _ = glib.ValueInit(glib.TYPE_BOOLEAN)
				gval.SetBool(args[i+1].(bool))

			case string:
				gval, _ = glib.ValueInit(glib.TYPE_STRING)
				gval.SetString(args[i+1].(string))

			case uint:
				gval, _ = glib.ValueInit(glib.TYPE_UINT)
				gval.SetUInt(args[i+1].(uint))

			default:
				log.Info("addArguments unknown type", reflect.TypeOf(args[i+1]))
			}

			values = values.Append(unsafe.Pointer(gval.Native()))
		}
	}

	return names, values, argsOut
}

func (v *ServiceProxy) addArgumentsOut(args []interface{}) (names, types, values *List) {

	lNames := &List{}
	lTypes := &List{}
	lValues := &List{}

	if len(args) >= 2 {

		// log.DEV("args out", len(args), args)

		var gval *glib.Value
		for i := 0; i < len(args); i += 2 {

			gstr := (*C.gchar)(C.CString(args[i].(string)))
			defer C.free(unsafe.Pointer(gstr))
			lNames = lNames.Append(unsafe.Pointer(C.g_strdup(gstr)))

			switch args[i+1].(type) {
			case *bool:
				gval, _ = glib.ValueInit(glib.TYPE_BOOLEAN)
				gval.SetBool(*args[i+1].(*bool))

			case *string:
				gval, _ = glib.ValueInit(glib.TYPE_STRING)
				gval.SetString(*args[i+1].(*string))

			case *uint:
				gval, _ = glib.ValueInit(glib.TYPE_UINT)
				gval.SetUInt(*args[i+1].(*uint))

			default:
				log.Info("addArguments unknown type", reflect.TypeOf(args[i+1]))
			}

			_, gtype, _ := gval.Type()
			lTypes = lTypes.Append(unsafe.Pointer(C.intToPointer(C.int(gtype))))

			lValues = lValues.Append(unsafe.Pointer(gval.Native()))

			// log.Info("type", gtype)
		}
	}
	return lNames, lTypes, lValues
}

//

//
//-----------------------------------------------------------------[ HELPERS ]--

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func wrapObject(ptr unsafe.Pointer) *glib.Object {
	obj := &glib.Object{glib.ToGObject(ptr)}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)
	return obj
}

//
//-------------------------------------------------------------------[ GList ]--

type List struct {
	GList *C.GList
}

func ListFromNative(l unsafe.Pointer) *List {
	return &List{
		C.toGlist(l)}
}

// func (v List) Data() interface{} {
//     return v.GList.data
// }
func (v List) Append(data unsafe.Pointer) *List {
	return &List{C.g_list_append(v.GList, C.gpointer(data))}
}

// func (v List) Prepend(data unsafe.Pointer) *List {
//     return &List{C.g_list_prepend(v.GList, C.gpointer(data))}
// }
// func (v List) Insert(data unsafe.Pointer, pos int) *List {
//     return &List{C.g_list_insert(v.GList, C.gpointer(data), C.gint(pos))}
// }
// func (v List) InsertBefore(sib List, data unsafe.Pointer) *List {
//     return &List{C.g_list_insert_before(v.GList, sib.GList, C.gpointer(data))}
// }

// //GList*              g_list_insert_sorted                (GList *list,
// //                                                         gpointer data,
// //                                                         GCompareFunc func);
// func (v List) Remove(data unsafe.Pointer) *List {
//     return &List{C.g_list_remove(v.GList, C.gconstpointer(data))}
// }
// func (v List) RemoveLink(link List) *List {
//     return &List{C.g_list_remove_link(v.GList, link.GList)}
// }
// func (v List) DeleteLink(link List) *List {
//     return &List{C.g_list_delete_link(v.GList, link.GList)}
// }
// func (v List) RemoveAll(data unsafe.Pointer) *List {
//     return &List{C.g_list_remove_all(v.GList, C.gconstpointer(data))}
// }
func (v List) Free() {
	C.g_list_free(v.GList)
}

// func GListAlloc() *List {
//     return &List{C.g_list_alloc()}
// }
func (v List) Free1() {
	C.g_list_free_1(v.GList)
}
func (v List) Length() uint {
	return uint(C.g_list_length(v.GList))
}

// func (v List) Copy() *List {
//     return &List{C.g_list_copy(v.GList)}
// }
// func (v List) Reverse() *List {
//     return &List{C.g_list_reverse(v.GList)}
// }

// //GList*              g_list_sort                         (GList *list,
// //                                                         GCompareFunc compare_func);
// //gint                (*GCompareFunc)                     (gconstpointer a,
// //                                                         gconstpointer b);
// //GList*              g_list_insert_sorted_with_data      (GList *list,
// //                                                         gpointer data,
// //                                                         GCompareDataFunc func,
// //                                                         gpointer user_data);
// //GList*              g_list_sort_with_data               (GList *list,
// //                                                         GCompareDataFunc compare_func,
// //                                                         gpointer user_data);
// //gint                (*GCompareDataFunc)                 (gconstpointer a,
// //                                                         gconstpointer b,
// //                                                         gpointer user_data);
// func (v List) Concat(link List) *List {
//     return &List{C.g_list_concat(v.GList, link.GList)}
// }
// func (v List) ForEach(callback func(interface{}, interface{}), user_datas ...interface{}) {
//     var user_data interface{}
//     if len(user_datas) > 0 {
//         user_data = user_datas[0]
//     }
//     l := v.First()
//     for n := uint(0); n < l.Length(); n++ {
//         callback(l.NthData(n), user_data)
//     }
// }
// func (v List) First() *List {
//     return &List{C.g_list_first(v.GList)}
// }
// func (v List) Last() *List {
//     return &List{C.g_list_last(v.GList)}
// }
// func (v List) Nth(n uint) *List {
//     return &List{C.g_list_nth(v.GList, C.guint(n))}
// }
func (v List) NthData(n uint) interface{} {
	return C.g_list_nth_data(v.GList, C.guint(n))
}

// func (v List) NthPrev(n uint) *List {
//     return &List{C.g_list_nth_prev(v.GList, C.guint(n))}
// }
// func (v List) Find(data unsafe.Pointer) *List {
//     return &List{C.g_list_find(v.GList, C.gconstpointer(data))}
// }

// //GList*              g_list_find_custom                  (GList *list,
// //                                                         gconstpointer data,
// //                                                         GCompareFunc func);
// func (v List) Position(link List) int {
//     return int(C.g_list_position(v.GList, link.GList))
// }
// func (v List) Index(data unsafe.Pointer) int {
//     return int(C.g_list_index(v.GList, C.gconstpointer(data)))
// }

/*


#define MAX_BROWSE 64

        if (didl_xml) {
                GUPnPDIDLLiteParser *parser;
                gint32              remaining;
                gint32              batch_size;
                GError              *error;

                error = NULL;
                parser = gupnp_didl_lite_parser_new ();

                g_signal_connect (parser,
                                  "object-available",
                                  G_CALLBACK (on_didl_object_available),
                                  data);

                // Only try to parse DIDL if server claims that there was a result
                if (number_returned > 0)
                        if (!gupnp_didl_lite_parser_parse_didl (parser,
                                                                didl_xml,
                                                                &error)) {
                                g_warning ("Error while browsing %s: %s",
                                           data->id,
                                           error->message);
                                g_error_free (error);
                        }

                g_object_unref (parser);
                g_free (didl_xml);

                data->starting_index += number_returned;

                // See if we have more objects to get
                remaining = total_matches - data->starting_index;

                // Keep browsing till we get each and every object
                if ((remaining > 0 || total_matches == 0) && number_returned != 0) {
                        if (remaining > 0)
                                batch_size = MIN (remaining, MAX_BROWSE);
                        else
                                batch_size = MAX_BROWSE;

                        browse (content_dir,
                                data->id,
                                data->starting_index,
                                batch_size);
                } else
                        update_container_child_count (content_dir, data->id);
        } else if (error) {
                GUPnPServiceInfo *info;

                info = GUPNP_SERVICE_INFO (content_dir);
                g_warning ("Failed to browse '%s': %s",
                           gupnp_service_info_get_location (info),
                           error->message);

                g_error_free (error);
        }
*/
