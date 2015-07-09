// Package upnptype defines common types for the upnp package.
package upnptype

import (
	"encoding/xml"
	"fmt"
	"time"
)

// type ControlPoint interface {
// }

// Action defines an UPnP simple renderer action.
//
type Action int

// Actions defined by the media controler.
//
const (
	ActionNone Action = iota
	ActionToggleMute
	ActionVolumeDown
	ActionVolumeUp
	ActionPlayPause
	ActionStop
	ActionSeekBackward
	ActionSeekForward
)

// Common settings.
//
const (
	MaxBrowse = 64
)

// Logger defines logging forwarding for the API.
//
type Logger interface {
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
}

// MediaControl defines actions provided by the selected server and renderer.
//
type MediaControl interface {

	// DefineEvents returns pointers to control events callbacks.
	//
	DefineEvents() ControlPointEvents

	// Action sends an action message to the selected renderer.
	//
	Action(Action) error

	//
	//---------------------------------------------------------------[ RENDERERS ]--

	// Renderer return the current renderer if any.
	//
	Renderer() Renderer

	// RendererExists return true if a renderer is selected.
	//
	RendererExists() bool

	// SetRenderer sets the current renderer by its udn reference.
	//
	SetRenderer(udn string)

	// Renderers return the list of all known renderers.
	//
	Renderers() Renderers

	//
	//-----------------------------------------------------------------[ SERVERS ]--

	// Server return the current server if any.
	//
	Server() Server

	// ServerExists return true if a server is selected.
	//
	ServerExists() bool

	// SetServer sets the current server by its udn reference.
	//
	SetServer(udn string)

	// Servers return the list of all known servers.
	//
	Servers() map[string]Server

	//
	//-----------------------------------------------------------------[ ?? ]--

	// SetVolumeDelta configures the default volume delta for user actions.
	//
	SetVolumeDelta(delta int)

	// SetSeekDelta configures the default volume delta for user actions.
	//
	SetSeekDelta(delta int)

	// SetPreferredRenderer sets the renderer that will be selected if found (unless anoter is selected).
	//
	SetPreferredRenderer(name string)

	// SetPreferredServer sets the server that will be selected if found (unless anoter is selected)..
	//
	SetPreferredServer(name string)

	//
	//--------------------------------------------------------------------[ TIME ]--

	// SetSeek seeks to new time in track.
	//
	Seek(unit, target string) error

	// SetSeekPercent seeks to new time in track. Input is the percent position in track. Range 0 to 100.
	//
	SeekPercent(value float64) error

	// target = "%d:%02d:%02d" % (hours,minutes,seconds)

	// GetCurrentTime returns the current track position on selected renderer in seconds.
	//
	GetCurrentTime() int

	//
	//------------------------------------------------------------------[ BROWSE ]--

	// Browse lists files on a server.
	//
	Browse(container string, startingIndex uint32) ([]Container, []Object, uint, uint)

	// BrowseMetadata starts the playback of the given file on the selected renderer.
	//
	BrowseMetadata(container string, startingIndex uint) error

	// AddURIToQueue(req *AddURIToQueueIn) (*AddURIToQueueOut, error)

	// SetNextAVTransportURI sets the next playback URI,
	//
	SetNextAVTransportURI(nextURI, nextURIMetaData string) error

	//
	//---------------------------------------------------------------[ HOOKS ]--

	// SubscribeHook register a new hook client and returns the MediaHook to connect to.
	//
	SubscribeHook(id string) *MediaHook

	// UnsubscribeHook removes a client hook.
	//
	UnsubscribeHook(id string)
}

//
//-------------------------------------------------------[ RENDERING CONTROL ]--

//
//-------------------------------------------------------[ ... ]--

// serviceID defines common actions provided by an UPnP device.
//
type serviceID interface {
	UDNer
	Name() string
	Icon() string
	SetIcon(icon string)
}

// UDNer defines an object that returns its UPnP ID.
//
type UDNer interface {
	UDN() string
}

// Server defines actions provided by an UPnP file server.
//
type Server interface {
	serviceID
	ServiceContentDirectory

	// CompareProxy compares two devices to see if they points to the same object.
	//
	CompareProxy(devtest UDNer) bool

	// GetIconFile gets the device icon location.
	//
	GetIconFile(filename string) string
}

// ServiceContentDirectory defines actions provided by an UPnP file server.
//
type ServiceContentDirectory interface {
	// GetSearchCapabilities() (searchCaps string, err error)
	// GetSortCapabilities() (sortCaps string, err error)
	// GetSystemUpdateID() (id uint32, err error)
	// GetAlbumArtistDisplayOption() (albumArtistDisplayOption string, err error)
	// GetLastIndexChange() (lastIndexChange string, err error)

	// Browse lists files on a server.
	//
	Browse(req *BrowseRequest) (browseResult *BrowseResult, err error)

	BrowseMetadata(container string, startingIndex, requestedCount uint) ([]Container, []Item, string)

	// FindPrefix(objectId, prefix string) (startingIndex, updateId uint32, err error)
	// GetAllPrefixLocations(objectId string) (prefixLocations *PrefixLocations, err error)
	// CreateObject(container, elements string) (objectId, result string, err error)
	// UpdateObject(objectId, currentTagValue, newTagValue string) (err error)

	//
	// Remove the directory object given by @objectId (e.g. "SQ:11", to
	// remove a saved queue). A 701 error is returned if an invalid @objectId
	// is specified.
	//
	// DestroyObject(objectId string) error
	// RefreshShareList() (err error)
	// RefreshShareIndex(albumArtistDisplayOption string) (err error)
	// RequestResort(sortOrder string) (err error)
	// GetShareIndexInProgress() (isIndexing bool, err error)
	// GetBrowseable() (isBrowseable bool, err error)
	// SetBrowseable(browseable bool) (err error)
}

// Browse common options.
//
const (
	BrowseObjectIDRoot = "0"

	BrowseFlagBrowseMetadata       = "BrowseMetadata"
	BrowseFlagBrowseDirectChildren = "BrowseDirectChildren"

	BrowseFilterAll = "*"

	BrowseSortCriteriaNone = ""
)

// BrowseRequest defines input parameters for Browse.
//
type BrowseRequest struct {
	ObjectID      string
	BrowseFlag    string
	Filter        string
	StartingIndex uint32
	RequestCount  uint32
	SortCriteria  string
}

// BrowseResult defines output value for Browse.
//
type BrowseResult struct {
	NumberReturned int32
	TotalMatches   int32
	UpdateID       int32
	// Doc            *didl.Lite
	Container []Container
	Item      []Object
}

// Renderer defines actions provided by an UPnP renderer.
//
type Renderer interface {
	serviceID

	// Events returns the renderer events callbacks.
	//
	Events() *RendererEvents

	// Duration() int

	// CompareProxy compares two devices to see if they points to the same object.
	//
	CompareProxy(devtest UDNer) bool

	// GetIconFile gets the device icon location.
	//
	GetIconFile(filename string) string

	ServiceRenderingControl
	ServiceAVTransport

	//-------------------------------------------------------------[ AVTRANSPORT ]--

	// PlayPause toggles the play / pause action on the renderer.
	//
	PlayPause(instanceID uint32, speed string) error

	// SetAVTransportURI(instanceID uint32, currentURI, currentURIMetaData string) error

	// AddURIToQueue(instanceID uint32, req *AddURIToQueueIn) (*AddURIToQueueOut, error)

	// GetCurrentTime returns the current track position on selected renderer in seconds.
	//
	GetCurrentTime() int

	// DisplayCurrentTime forwards the current track position with OnCurrentTime event.
	//
	DisplayCurrentTime()
}

// ServiceRenderingControl defines actions provided by the renderer control.
//
//   @instanceID will always be 0 for most devices.
//   @channel is one of the constants: ChannelMaster, ChannelRF, ChannelLF.
//
type ServiceRenderingControl interface {
	GetMute(instanceID uint32, channel string) (currentMute bool, e error)
	SetMute(instanceID uint32, channel string, desiredMute bool) error

	GetVolume(instanceID uint32, channel string) (currentVolume uint16, e error)

	// Set the playback volume.
	// @volume is an integer between 0 and 100, where 100 is the loudest.
	//
	SetVolume(instanceID uint32, channel string, volume uint16) error
	SetRelativeVolume(instanceID uint32, channel string, adjustment int32) (newVolume uint16, e error)

	// ResetBasicEQ(instanceID uint32) (basicEQ *BasicEQ, err error)
	// ResetExtEQ(instanceID uint32, eqType string) (err error)

	// GetVolumeDB(instanceID uint32, channel string) (currentVolume int16, err error)
	// SetVolumeDB(instanceID uint32, channel string, volume int16) (err error)
	// GetVolumeDBRange(instanceID uint32, channel string) (min, max int16, err error)
	// GetBass(instanceID uint32) (currentBass int16, err error)
	// SetBass(instanceID, desiredBass int16) (err error)
	// GetTreble(instanceID uint32) (currentTreble int16, err error)
	// SetTreble(instanceID, desiredTreble int16) (err error)
	// GetEQ(instanceID uint32, eqType string) (currentValue int16, err error)
	// SetEQ(instanceID uint32, eqType string, desiredValue int16) (err error)
	// GetLoudness(instanceID uint32, channel string) (loudness bool, err error)
	// SetLoudness(instanceID uint32, channel string, loudness bool) (err error)
	// GetSupportsOutputFixed(instanceID uint32) (currentSupportsFixed bool, err error)
	// GetOutputFixed(instanceID uint32) (currentFixed bool, err error)
	// SetOutputFixed(instanceID uint32, desiredFixed bool) (err error)
	// GetHeadphoneConnected(instanceID uint32) (currentHeadphoneConnected bool, err error)

	// RampToVolume(instanceID uint32, channel, req RampRequest) (rampTime uint32, err error)
	// RestoreVolumePriorToRamp(instanceID uint32, channel string) (err error)
	// SetChannelMap(instanceID uint32, channelMap string) (err error)

	// /* Reciva */
	// ListPresets(instanceID uint32) (presets string, err error)
	// SelectPreset(instanceID uint32, presetName string) erro
}

// ServiceAVTransport defines actions provided by the renderer transport.
//
type ServiceAVTransport interface {

	// SetAVTransportURI sets the current playback URI,
	//
	//   @currentURI will be a valid URI as given by the Res() attribute of a
	//               ContentDirectory object.
	//   @currentURIMetaData can be empty.
	//
	// Use this method to, for example, resume playback from a queue after
	// playback from a radio station or other source.
	//
	SetAVTransportURI(instanceID uint32, currentURI, currentURIMetaData string) error

	//
	// AddURIToQueue adds a single track to the queue (Q:0).
	// See AddURIToQueueIn for a discussion of the input
	// parameters and AddURIToQueueOut for a discussion of the output
	// parameters.
	//
	AddURIToQueue(instanceID uint32, req *AddURIToQueueIn) (*AddURIToQueueOut, error)

	//
	// Add multiple tracks to the queue (Q:0).  This method does not seem
	// to be a standard part of AVTransport:1, but rather a Sonos extension.
	// As such it is not entirely clear how it should be used.
	//
	// For Sonos @instanceID should always be 0; @UpdateID should be 0;
	// @NumberOfURIs should be the number of tracks to be added by the
	// request; @EnqueuedURIs is a space-separated list of URIs (as given by
	// the Res() method of the model.Object class); @EnqueuedURIMetData is a
	// space-separated list of DIDL-Lite documents describing the resources
	// to be added; @ContainerURI should be the ContentDirectory URI for
	// A:TRACK, when adding tracks; @ContainerMetaData should be a DIDL-Lite
	// document describing A:TRACK. Other arguments have the same meaning as
	// in @AddURIToQueue.
	//
	// Note that the number of DIDL-Lite documents in @EnqueuedURIsMetaData
	// must match the number of URIs in @EnqueuedURIs.  These DIDL-Lite documents
	// can be empty, but must be present.  @ContainerMetaData must be a string
	// of non-zero length, but need not be a valid DIDL-Lite document.
	//
	AddMultipleURIsToQueue(instanceID uint32, req *AddMultipleURIsToQueueIn) (*AddMultipleURIsToQueueOut, error)

	//
	// // Move a contiguous range of tracks to a given point in the queue.
	// // For Sonos @instanceID will always be 0; @startingIndex is the first
	// // track in the range to be moved, where the first track in the queue is
	// // track 1; @numberOfTracks is the length of the range; @insertBefore set
	// // the destination position in the queue; @updateId should be 0.
	// //
	// // Note that to move tracks to the end of the queue @insertBefore must be
	// // set to the number of tracks in the queue plus 1.  This method fails with
	// // 402 if @startingndex, @numberOfTracks, or @insertBefore are out of range.
	// //
	// ReorderTracksInQueue(instanceID, startingIndex, numberOfTracks, insertBefore, updateId uint32) error

	// //
	// // Remove a single track from the queue (Q:0).  For Sonos @instanceID
	// // will always be 0; @objectId will be the identifier of the item to be
	// // removed from the queue (e.g. "Q:0/5" for the fifth element in the queue);
	// // @updateId will always be 0.
	// //
	// RemoveTrackFromQueue(instanceID uint32, objectId string, updateId uint32) error

	// //
	// // Remove a continguous range of tracks from the queue (Q:0).  For Sonos
	// // @instanceID will always be 0; @updateId should be 0; @startingIndex is
	// // the first track to remove where the first track is 1; @numberOfTracks
	// // is the number of tracks to remove.  Returns the new @updateId.
	// //
	// // This method fails with 402 if either @startingIndex or @numberOfTracks
	// // is out of range.
	// //
	// RemoveTrackRangeFromQueue(instanceID, updateId, startingIndex, numberOfTracks uint32) (uint32, error)

	// //
	// // Remove all tracks from the queue (Q:0).  For Sonos @instanceID will
	// // always be 0.  Emptying an already empty queue is not an error.
	// //
	// RemoveAllTracksFromQueue(instanceID uint32) error

	// //
	// // Create a new named queue (SQ:n) from the contents of the current
	// // queue (Q:0).  For Sonos @instanceID should always be 0; @title is the
	// // display name of the new named queue; @objectId should be left blank.
	// // This method returns the objectId of the newly created queue.
	// //
	// SaveQueue(instanceID uint32, title, objectId string) (string, error)
	// BackupQueue(instanceID uint32) (err error)

	// GetMediaInfo gets information about the currently selected media, its URI,
	// length in tracks, and recording status, if any.
	//
	// Many renderers will have a lot of the fields unsupported.
	//
	GetMediaInfo(instanceID uint32) (*MediaInfo, error)

	// GetTransportInfo returns the current state of the transport (playing,
	// stopped, paused), its status, and playback speed.
	//
	GetTransportInfo(instanceID uint32) (*TransportInfo, error)

	// GetPositionInfo returns information about the track that is currently
	// playing, its URI, position in the queue (Q:0), and elapsed time.
	//
	GetPositionInfo(instanceID uint32) (*PositionInfo, error)

	// //
	// // Return the device capabilities, sources of input media, recording
	// // media, and recoding quality modes.  For Sonos @instanceID should always
	// // be 0, and the record-related fields are unsupported.
	// //
	// GetDeviceCapabilities(instanceID uint32) (*DeviceCapabilities, error)

	// //
	// // Return the current transport settings; the playback mode (NORMAL,
	// // REPEAT_ALL, SHUFFLE, etc.); and the recoding quality (not support
	// // on Sonos).  For Sonos @instanceID will always with 0.
	// //
	// GetTransportSettings(instanceID uint32) (*TransportSettings, error)

	// //
	// // Returns true if crossfade mode is active; false otherwise.  For Sonos
	// // @instanceID should always be 0.
	// //
	// GetCrossfadeMode(instanceID uint32) (bool, error)

	// Stops playback and return to the beginning of the queue (Q:1).
	//
	Stop(instanceID uint32) error

	// Starts or resumes playback at the given speed.
	//
	//   @speed is a fraction relative to normal speed (e.g. "1" or "1/2").
	//
	Play(instanceID uint32, speed string) error

	// Pause playback, prepared to resume at the current position.  For Sonos
	// @instanceID should always be 0.
	//
	Pause(instanceID uint32) error

	//
	// A general function to seek within the playback queue (Q:0).  For Sonos
	// @instanceID should always be 0; @unit should be one of the values given
	// for seek mode (TRACK_NR, REL_TIME, or SECTION); and @target should
	// give the track, time offset, or section where playback should resume.
	//
	// For TRACK_NR the integer track number relative to the start of the queue
	// is supplied to @target.  For REL_TIME a duration in the format HH:MM:SS
	// is given as @target.  SECTION is not tested.
	//
	Seek(instanceID uint32, unit, target string) error

	//
	// Skip ahead to the next track in the queue (Q:).  For Sonos @instanceID
	// should always be 0.  This method returns an error 711 if the current
	// track is the last track in the queue.
	//
	Next(instanceID uint32) error

	// NextProgrammedRadioTracks(instanceID uint32) (err error)

	//
	// Move to a previous track in the queue (Q:0).  For Sonos @instanceID
	// should always be 0.  This method returns error 711 if the current track
	// is the first track in the queue.
	//
	Previous(instanceID uint32) error

	// //
	// // Advance one section in the current track.  For Sonos @instanceID will
	// // always be zero.  Experimentally this method returns 711 if the current
	// // track does not contain multiple sections.
	// //
	// NextSection(instanceID uint32) error

	// //
	// // Retreat one section in the current track.  For Sonos @instanceID will
	// // always be zero.  Experimentally this method returns 711 if the current
	// // track does not contain multiple sections.
	// //
	// PreviousSection(instanceID int) error

	// //
	// // Set the current playback mode where @newPlayMode is one of the values
	// // given for PlayMode above.  For Sonos @instanceID should always be 0.
	// // This method returns 712 if an invalid @newPlayMode is supplied.
	// //
	// SetPlayMode(instanceID uint32, newPlayMode string) error

	// //
	// // Toggle crossfade mode on or off.  For Sonos @instanceID should always
	// // be 0.  If @crossfadeMode is true crossfade mode will be enabled; otherwise
	// // disabled.
	// //
	// SetCrossfadeMode(instanceID uint32, crossfadeMode bool) error

	// NotifyDeletedURI(instanceID uint32, deletedURI string) (err error)

	//
	// Returns a list of the actions that are valid at this time.  The list
	// consists of human-readable strings, such as "Play", and "Stop".  For Sonos
	// @instanceID will always be 0.
	//
	GetCurrentTransportActions(instanceID uint32) ([]string, error)

	// BecomeCoordinatorOfStandaloneGroup(instanceID uint32) (err error)
	// BecomeGroupCoordinator(instanceID uint32, req *BecomeGroupCoordinatorRequest) (err error)
	// BecomeGroupCoordinatorAndSource(instanceID uint32, req *BecomeGroupCoordinatorAndSourceRequest) (err error)
	// ChangeCoordinator(instanceID uint32, req *ChangeCoordinatorRequest) (err error)
	// ChangeTransportSettings(instanceID uint32, newTransportSettings, currentAVTransportURI string) (err error)
	// ConfigureSleepTimer(instanceID uint32, newSleepTimerDuration string) (err error)
	// GetRemainingSleepTimerDuration(instanceID uint32) (remainingSleepTimerDuration string)
	// RunAlarm(instanceID uint32, req *RunAlarmRequest) (err error)
	// StartAutoplay(instanceID uint32, req *StartAutoplayRequest) (err error)
	// GetRunningAlarmProperties(instanceID uint32) (alarmId uint32, groupId, loggedStartTime string, err error)
	// SnoozeAlarm(instanceID uint32, duration string) (err error)
	// DelegateGroupCoordinationTo(instanceID uint32, newCoordinator string, rejoinGroup bool) error

	// SetNextAVTransportURI sets the next playback URI,
	//
	SetNextAVTransportURI(instanceID uint32, nextURI, nextURIMetaData string) error

	// CreateSavedQueue(instanceID uint32, req *CreateSavedQueueIn) (*CreateSavedQueueOut, error)
	// AddURIToSavedQueue(instanceID uint32, req *AddURIToSavedQueueIn) (*AddURIToSavedQueueOut, error)
	// ReorderTracksInSavedQueue(instanceID uint32, req *ReorderTracksInSavedQueueIn) (*ReorderTracksInSavedQueueOut, error)
}

// AddURIToQueueIn defines input parameters for AddURIToQueue.
//
type AddURIToQueueIn struct {
	// The URI of the track to be added to the queue, corresponding
	// the to <res> tag in a DIDL-Lite description (@see dldl,
	// @ContentDirectory, etc) e.g.:
	//     "x-file-cifs://servername/path/to/track.mp3"
	EnqueuedURI string
	// A DIDL-Lite document describing the the resource given by @EnqueuedURI
	EnqueuedURIMetaData string
	// This field should be 0 to insert the new item at the end
	// of the queue.  If non-zero the new track will be inserted at
	// this location, and later tracks will see their track numbers
	// incremented.
	DesiredFirstTrackNumberEnqueued uint32
	// ???? (possibly unsupported)
	EnqueueAsNext bool
}

// AddURIToQueueOut defines output value for AddURIToQueue
//
type AddURIToQueueOut struct {
	// The track number of the newly added track.
	FirstTrackNumberEnqueued uint32
	// The number of tracks added by this request (always 1).
	NumTracksAdded uint32
	// The length of the queue now that this track has been added
	NewQueueLength uint32
}

// AddMultipleURIsToQueueIn defines input parameters for AddMultipleURIsToQueue.
//
type AddMultipleURIsToQueueIn struct {
	// UpdateID (in), can be 0
	UpdateID uint32
	// The number of URIs to be added in this request
	NumberOfURIs uint32
	// A list of @NumberOfURIs URIs, separated by a space
	EnqueuedURIs string
	// A list of @NumberOfURIs DIDL-Lite documents, separated by a space
	EnqueuedURIsMetaData string
	// URI of a container in the ContentDirectory containing the
	// URIs to be added.  If adding tracks this should be the URI for
	// the A:TRACK entry in the directory.
	ContainerURI string
	// A DIDL-Lite document describing the resource given by @ContainerURI
	ContainerMetaData string
	// This field should be 0 to insert the new item at the end
	// of the queue.  If non-zero the new track will be inserted at
	// this location, and later tracks will see their track numbers
	// incremented.
	DesiredFirstTrackNumberEnqueued uint32
	// ???? (possibly unsupported)
	EnqueueAsNext bool
}

// AddMultipleURIsToQueueOut defines output value for AddMultipleURIsToQueue.
//
type AddMultipleURIsToQueueOut struct {
	FirstTrackNumberEnqueued uint32 // The starting position int the queue (Q:0) of the newly added tracks
	NumTracksAdded           uint32 // The number of tracks added by the request
	NewQueueLength           uint32 // The length of the queue after the request was complete
	NewUpdateID              uint32 // The new UpdateID
}

// MediaInfo defines information about the currently selected media.
//
type MediaInfo struct {
	NrTracks           uint32 // The number of tracks in the queue (Q:0)
	MediaDuration      string // ???? (possibly not supported)
	CurrentURI         string // The URI of the queue
	CurrentURIMetaData string // ????
	NextURI            string // ???? (possibly not supported)
	NextURIMetaData    string // ???? (possibly not supported)
	PlayMedium         string // ????
	RecordMedium       string // ???? (possibly not supported)
	WriteStatus        string // ???? (possibly not supported)
}

// TransportInfo defines the current state of the transport.
//
type TransportInfo struct {
	// Indicates whether the device is playing, paused, or stopped
	CurrentTransportState string
	// Indicates if an error condition exists ("OK" otherwise)
	CurrentTransportStatus string
	// Playback speed relative to normal playback speed (e.g. "1" or "1/2")
	CurrentSpeed string
}

// PositionInfo defines information about the track.
//
type PositionInfo struct {
	// Track number relative to the beginning of the queue (not the containing album).
	Track uint32
	// Total length of the track in HH:MM:SS format
	TrackDuration string
	// The DIDL-Lite document describing this item in the ContentDirectory
	TrackMetaData string
	// The URI of the track, corresponding // the to <res> tag in
	// a DIDL-Lite description (@see dldl, @ContentDirectory, etc) e.g.:
	//     "x-file-cifs://servername/path/to/track.mp3"
	TrackURI string
	// The number of elapsed seconds into the track in HH:MM:SS format
	RelTime string
	// ???? (possibly unsupported)
	AbsTime string
	// ???? (possibly unsupported)
	RelCount uint32
	// ???? (possibly unsupported)
	AbsCount uint32
}

// // The return type of the GetDeviceCapabilities method
// //
// type DeviceCapabilities struct {
// 	// Configured sources of media
// 	PlayMedia string
// 	// ???? (possibly unsupported)
// 	RecMedia string
// 	// ???? (possibly unsupported)
// 	RecQualityModes string
// }

//
// TransportSettings defines output value for GetTransportSettings
//
type TransportSettings struct {
	// The current play mode (NORMAL, REPEAT_ALL, SHUFFLE, etc.)
	PlayMode string

	// The record quality (not supported in Sonos)
	RecQualityMode string
}

// Channels names.
//
const (
	ChannelMaster = "Master"
	ChannelRF     = "RF"
	ChannelLF     = "LF"
)

// Playback speed.
//
const PlaySpeedNormal = "1"

// Legal values for @unit in calls to Seek.
//
const (
	SeekModeTrackNR = "TRACK_NR" // Seek to the beginning of the specified track
	SeekModeAbsTime = "ABS_TIME" // Seek to the given absolute offset in the current track
	SeekModeRelTime = "REL_TIME" // Seek to the given relative offset in the current track
	SeekModeSection = "SECTION"  // Seek to the specified section (not tested)
)

// Valid values for PlayMode in SetPlayMode and TransportSettings.
//
const (
	PlayModeNormal          = "NORMAL"           // Play sequentially from the beginning of the queue to the end
	PlayModeRepeatAll       = "REPEAT_ALL"       // Begin again at the first track of the queue after reaching the last
	PlayModeShuffleNoRepeat = "SHUFFLE_NOREPEAT" // Play tracks out of order, with repeat
	PlayModeShuffle         = "SHUFFLE"          // Play through tracks out of order once
)

// type BasicEQ struct {
// 	Bass        int16
// 	Treble      int16
// 	Loudness    bool
// 	LeftVolume  uint16
// 	RightVolume uint16
// }

// const (
// 	RampType_SleepTimer = "SLEEP_TIMER_RAMP_TYPE"
// 	RampType_Alarm      = "ALARM_RAMP_TYPE"
// 	RampType_Autoplay   = "AUTOPLAY_RAMP_TYPE"
// )

// type RampRequest struct {
// 	RampType         string
// 	DesiredVolume    uint16
// 	ResetVolumeAfter bool
// 	ProgramURI       string
// }

// type BecomeGroupCoordinatorRequest struct {
// 	CurrentCoordinator    string
// 	CurrentGroupID        string
// 	OtherMembers          string
// 	TransportSettings     string
// 	CurrentURI            string
// 	CurrentURIMetaData    string
// 	SleepTimerState       string
// 	AlarmState            string
// 	StreamRestartState    string
// 	CurrentQueueTrackList string
// }

// type BecomeGroupCoordinatorAndSourceRequest struct {
// 	CurrentCoordinator    string
// 	CurrentGroupID        string
// 	OtherMembers          string
// 	CurrentURI            string
// 	CurrentURIMetaData    string
// 	SleepTimerState       string
// 	AlarmState            string
// 	StreamRestartState    string
// 	CurrentAVTTrackList   string
// 	CurrentQueueTrackList string
// 	CurrentSourceState    string
// 	ResumePlayback        bool
// }

// type ChangeCoordinatorRequest struct {
// 	CurrentCoordinator   string
// 	NewCoordinator       string
// 	NewTransportSettings string
// }

// type RunAlarmRequest struct {
// 	AlarmID            uint32
// 	LoggedStartTime    string
// 	Duration           string
// 	ProgramURI         string
// 	ProgramMetaData    string
// 	PlayMode           string
// 	Volume             uint32
// 	IncludeLinkedZones bool
// }

// type StartAutoplayRequest struct {
// 	ProgramURI         string
// 	ProgramMetaData    string
// 	Volume             uint32
// 	IncludeLinkedZones bool
// 	ResetVolumeAfter   bool
// }

// type CreateSavedQueueIn struct {
// 	Title               string
// 	EnqueuedURI         string
// 	EnqueuedURIMetaData string
// }

// type CreateSavedQueueOut struct {
// 	NumTracksAdded   uint32
// 	NewQueueLength   uint32
// 	AssignedObjectID string
// 	NewUpdateID      uint32
// }

// type AddURIToSavedQueueIn struct {
// 	ObjectID            string
// 	UpdateID            uint32
// 	EnqueuedURI         string
// 	EnqueuedURIMetaData string
// 	AddAtIndex          uint32
// }

// type AddURIToSavedQueueOut struct {
// 	NumTracksAdded uint32
// 	NewQueueLength uint32
// 	NewUpdateID    uint32
// }

// type ReorderTracksInSavedQueueIn struct {
// 	ObjectID        string
// 	UpdateID        uint32
// 	TrackList       string // dunno anything about A_ARG_TYPE_TrackList
// 	NewPositionList string // dunno anything about A_ARG_TYPE_TrackList
// }

// type ReorderTracksInSavedQueueOut struct {
// 	QueueLengthChange uint32
// 	NewQueueLength    uint32
// 	NewUpdateID       uint32
// }

//-------------------------------------------------------[ DIDL-Lite PARSING ]--

// Renderers defines a list of renderer indexed by udn.
//
type Renderers map[string]Renderer

//-----------------------------------------------------------[ DEVICE EVENTS ]--

// MediaHook provides a registration method to media events for multiple clients.
//
type MediaHook struct {
	ControlPointEvents
	RendererEvents

	// Local events.
	OnSetVolumeDelta   func(int)
	OnSetSeekDelta     func(int)
	OnRendererSelected func(Renderer)
	OnServerSelected   func(Server)
}

// ControlPointEvents defines events of a control point.
//
type ControlPointEvents struct {
	OnRendererFound func(Renderer)
	OnRendererLost  func(Renderer)
	OnServerFound   func(Server)
	OnServerLost    func(Server)
}

// RendererEvents defines events of a renderer.
//
type RendererEvents struct {
	OnTransportState       func(Renderer, PlaybackState)
	OnCurrentTrackDuration func(Renderer, int)
	OnCurrentTrackMetaData func(Renderer, *Item)

	OnMute   func(Renderer, bool)
	OnVolume func(Renderer, uint)

	OnCurrentTime func(r Renderer, secs int, percent float64)
}

//
//-----------------------------------------------------------[ PLAYBACKSTATE ]--

// DeviceBase provides a common device base to extend for backends.
//
type DeviceBase struct {
	udn  string
	name string
	icon string // path to icon file on disk.
}

// GetIconFile gets the device icon location.
//
func (db *DeviceBase) GetIconFile(filename string) string {
	// 	url, _, _, _, _ := db.proxy.DeviceInfo.GetIconUrl("", -1, 24, 24, true)
	// 	return getIconFile(url, filename)
	return ""
}

// UDN returns the UDN (UPnP ID) of the device.
//
func (db *DeviceBase) UDN() string { return db.udn }

// SetUDN sets the UDN (UPnP ID) of the device.
//
func (db *DeviceBase) SetUDN(udn string) { db.udn = udn }

// Name returns the friendly name of the device (if found).
//
func (db *DeviceBase) Name() string { return db.name }

// SetName sets the friendly name of the device.
//
func (db *DeviceBase) SetName(name string) { db.name = name }

// Icon returns the icon file location.
//
func (db *DeviceBase) Icon() string { return db.icon }

// SetIcon sets the icon file location.
//
func (db *DeviceBase) SetIcon(icon string) { db.icon = icon }

// CompareProxy compares two devices to see if they points to the same object.
//
func (db *DeviceBase) CompareProxy(devtest UDNer) bool {
	return db.UDN() == devtest.UDN()
}

// ServerBase provides a common server base to extend for backends.
//
type ServerBase struct {
	DeviceBase

	// events ServerEvents
}

// func (srv *ServerBase) Events() *ServerEvents { return &srv.events }

// RendererBase provides a common renderer base to extend for backends.
//
type RendererBase struct {
	DeviceBase

	events RendererEvents
}

// Events returns the renderer events callbacks.
//
func (rend *RendererBase) Events() *RendererEvents { return &rend.events }

//
//-----------------------------------------------------------[ PLAYBACKSTATE ]--

// PlaybackState defines the playback state of a renderer.
//
type PlaybackState int

// Renderer playback states.
const (
	PlaybackStateUnknown PlaybackState = iota
	PlaybackStateTransitioning
	PlaybackStateStopped
	PlaybackStatePaused
	PlaybackStatePlaying
)

// Legal values for TransportInfo.CurrentTransportState
//
const (
	StateTransitioning  = "TRANSITIONING"
	StateStopped        = "STOPPED"
	StatePausedPlayback = "PAUSED_PLAYBACK"
	StatePlaying        = "PLAYING"
)

// PlaybackStateFromName returns the PlaybackState for the given state string.
//
func PlaybackStateFromName(name string) PlaybackState {
	switch name {
	case StateStopped:
		return PlaybackStateStopped

	case StatePlaying:
		return PlaybackStatePlaying

	case StatePausedPlayback:
		return PlaybackStatePaused

	case StateTransitioning:
		return PlaybackStateTransitioning
	}

	return PlaybackStateUnknown
}

//
//--------------------------------------------------------------------[ TIME ]--

// TimeToString formats time input in seconds to HH:MM:SS string used by renderers.
//
// output as "15:04:05" format for seek requests.
//
func TimeToString(sec int) string {
	newtime := time.Time{}.Add(time.Duration(sec) * time.Second)
	return newtime.Format("15:04:05")
}

// TimeToSecond converts a time string to a duration in sceonds.
//
// input as "15:04:05" format as used by renderers.
//
func TimeToSecond(str string) int {
	var h, m, s int
	fmt.Sscanf(str, "%d:%d:%d", &h, &m, &s) // (n int, err error)
	return (h*60+m)*60 + s
}

// Clock ticks to refresh a renderer track position.
//
func Clock(rend Renderer, tick *time.Ticker, state PlaybackState) *time.Ticker {

	rend.GetCurrentTime()
	rend.DisplayCurrentTime()

	if tick != nil {
		tick.Stop()
	}

	switch state {
	case PlaybackStatePlaying:
		tick = time.NewTicker(time.Second)
		go func() {
			for _ = range tick.C {
				rend.GetCurrentTime()
				rend.DisplayCurrentTime()
			}
		}()
	}
	return tick
}

//
//-------------------------------------------------------[ DIDL-Lite PARSING ]--

// Resource defines a server file resource.
//
type Resource struct {
	// XMLName      xml.Name `xml:"res"`
	ProtocolInfo string `xml:"protocolInfo,attr"`
	URL          string `xml:",chardata"`
	Size         uint64 `xml:"size,attr,omitempty"`
	Bitrate      uint   `xml:"bitrate,attr,omitempty"`
	Duration     string `xml:"duration,attr,omitempty"`
	Resolution   string `xml:"resolution,attr,omitempty"`
}

// Container defines a server file container.
type Container struct {
	Object
	XMLName    xml.Name `xml:"container"`
	ChildCount int      `xml:"childCount,attr"`
}

// Item defines a server file object with resources.
//
type Item struct {
	Object
	XMLName xml.Name   `xml:"item"`
	Res     []Resource `xml:"res"`
}

// Object defines a server file object description.
//
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
