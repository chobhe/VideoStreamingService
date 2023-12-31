package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

var _ rtmp.Handler = (*Handler)(nil)

// Handler is an implementation of the rtmp.Handler interface.
// It provides methods for handling RTMP connection events, such as OnConnect, OnPublish, and OnPlay.
type Handler struct {
	rtmp.DefaultHandler               // Inherits the default implementations of the rtmp.Handler interface
	relayService        *RelayService // The relay service used to manage streams

	conn *rtmp.Conn // The RTMP connection associated with this Handler

	pub *Pub // The publisher associated with this Handler
	sub *Sub // The subscriber associated with this Handler
}

// OnServe is called when the RTMP connection is ready to serve.
// It sets the conn field of the Handler to the given connection.
func (h *Handler) OnServe(conn *rtmp.Conn) {
	h.conn = conn
}

// OnConnect is called when the RTMP connection is established.
// It logs the connection command details.
func (h *Handler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Printf("OnConnect: %#v", cmd)
	// Like if it's a play command or a publish command

	// TODO: check app name to distinguish stream names per apps
	// cmd.Command.App

	return nil
}

// OnCreateStream is called when a new RTMP stream is created.
// It logs the stream creation command details.
func (h *Handler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Printf("OnCreateStream: %#v", cmd)
	return nil
}

// OnPublish is called when a RTMP publish command is received.
// It creates a new Pubsub in the RelayService with the stream name provided in the command.
// The method will error out if a stream is already published or the stream name is empty.
func (h *Handler) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Printf("OnPublish: %#v", cmd)

	if h.sub != nil {
		return errors.New("Cannot publish to this stream")
	}

	// (example) Reject a connection when PublishingName is empty
	if cmd.PublishingName == "" {
		return errors.New("PublishingName is empty")
	}

	pubsub, err := h.relayService.NewPubsub(cmd.PublishingName)
	if err != nil {
		return errors.Wrap(err, "Failed to create pubsub")
	}

	pub := pubsub.Pub()

	h.pub = pub

	return nil
}

// OnPlay is called when a RTMP play command is received.
// It retrieves the Pubsub associated with the stream name from the RelayService and initializes a subscriber for it.
func (h *Handler) OnPlay(ctx *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPlay) error {
	if h.sub != nil {
		return errors.New("Cannot play on this stream")
	}

	pubsub, err := h.relayService.GetPubsub(cmd.StreamName)
	if err != nil {
		return errors.Wrap(err, "Failed to get pubsub")
	}

	sub := pubsub.Sub()
	sub.eventCallback = onEventCallback(h.conn, ctx.StreamID)

	h.sub = sub

	return nil
}

// OnSetDataFrame is called when a RTMP set data frame command is received.
// It decodes the data frame payload into a ScriptData object and publishes it to the stream.
func (h *Handler) OnSetDataFrame(timestamp uint32, data *rtmpmsg.NetStreamSetDataFrame) error {
	r := bytes.NewReader(data.Payload)

	var script flvtag.ScriptData
	if err := flvtag.DecodeScriptData(r, &script); err != nil {
		log.Printf("Failed to decode script data: Err = %+v", err)
		return nil // ignore
	}

	log.Printf("SetDataFrame: Script = %#v", script)

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeScriptData,
		Timestamp: timestamp,
		Data:      &script,
	})

	return nil
}

// OnAudio is called when an RTMP audio data packet is received.
// It decodes the audio data and publishes it to the stream.
func (h *Handler) OnAudio(timestamp uint32, payload io.Reader) error {
	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	flvBody := new(bytes.Buffer)
	if _, err := io.Copy(flvBody, audio.Data); err != nil {
		return err
	}
	audio.Data = flvBody

	fmt.Printf("%v", flvBody)

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeAudio,
		Timestamp: timestamp,
		Data:      &audio,
	})

	return nil
}

// OnVideo is called when an RTMP video data packet is received.
// It decodes the video data and publishes it to the stream.
func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {
	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	// Need deep copy because payload will be recycled
	flvBody := new(bytes.Buffer)
	written, err := io.Copy(flvBody, video.Data)
	if err != nil {
		return err
	}

	fmt.Printf("Copied %v bytes", written)

	// Send flv file to python webserver microservice
	SendByteStreamPost(h, "INSERT PYTHON MICROSERVICE URL", flvBody.Bytes())

	video.Data = flvBody
	fmt.Printf("%v", flvBody)

	_ = h.pub.Publish(&flvtag.FlvTag{
		TagType:   flvtag.TagTypeVideo,
		Timestamp: timestamp,
		Data:      &video,
	})

	return nil
}

// OnClose is called when the RTMP connection is closed.
// It closes the publisher and subscriber associated with this handler.
func (h *Handler) OnClose() {
	log.Printf("OnClose")

	if h.pub != nil {
		_ = h.pub.Close()
	}

	if h.sub != nil {
		_ = h.sub.Close()
	}
}

// Stream byte data to the microservice
func SendByteStreamPost(h *Handler, url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/octet-stream")

	res, err := h.relayService.httpClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return []byte{}, err
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

// Send flv file to the microservice
// func sendMultiPartPost(h *Handler, url string, data []byte) ([]byte, error) {
// 	var (
// 		buf = new(bytes.Buffer)
// 		w   = multipart.NewWriter(buf)
// 	)
// 	part, err := w.CreateFormFile("flv", filepath.Base("/flv_data.flv"))
// 	if err != nil {
// 		return []byte{}, err
// 	}

// 	_, err = part.Write(data)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	err = w.Close()
// 	if err != nil {
// 		return []byte{}, err
// 	}

// 	req, err := http.NewRequest("POST", url, buf)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	req.Header.Add("Content-Type", w.FormDataContentType())

// 	res, err := h.relayService.httpClient.Do(req)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	defer res.Body.Close()

// 	content, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	return content, nil
// }
