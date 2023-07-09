package main

import (
	"bytes"
	"context"

	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

// onEventCallback generates a callback function for handling FLV tags (events) in an RTMP stream.
func onEventCallback(conn *rtmp.Conn, streamID uint32) func(flv *flvtag.FlvTag) error {
	// The generated function takes an FLV tag and returns an error.
	return func(flv *flvtag.FlvTag) error {
		// A new buffer is created to hold the encoded data of the event.
		buf := new(bytes.Buffer)

		// The type of the event data is checked.
		switch flv.Data.(type) {

		// Case when the event data is of type AudioData.
		case *flvtag.AudioData:
			// The event data is cast to AudioData.
			d := flv.Data.(*flvtag.AudioData)

			// The AudioData is encoded into the buffer.
			if err := flvtag.EncodeAudioData(buf, d); err != nil {
				return err
			}

			// An AudioMessage is written to the RTMP connection.
			// The hardcoded chunk stream ID (5) may need to be replaced with a dynamic value.
			ctx := context.Background()
			chunkStreamID := 5
			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
				StreamID: streamID,
				Message: &rtmpmsg.AudioMessage{
					Payload: buf,
				},
			})

		// Case when the event data is of type VideoData.
		case *flvtag.VideoData:
			// The event data is cast to VideoData.
			d := flv.Data.(*flvtag.VideoData)

			// The VideoData is encoded into the buffer.
			if err := flvtag.EncodeVideoData(buf, d); err != nil {
				return err
			}

			// A VideoMessage is written to the RTMP connection.
			// The hardcoded chunk stream ID (6) may need to be replaced with a dynamic value.
			ctx := context.Background()
			chunkStreamID := 6
			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
				StreamID: streamID,
				Message: &rtmpmsg.VideoMessage{
					Payload: buf,
				},
			})

		// Case when the event data is of type ScriptData.
		case *flvtag.ScriptData:
			// The event data is cast to ScriptData.
			d := flv.Data.(*flvtag.ScriptData)

			// The ScriptData is encoded into the buffer.
			if err := flvtag.EncodeScriptData(buf, d); err != nil {
				return err
			}

			// A new buffer is created to hold the AMF-encoded DataMessage.
			amdBuf := new(bytes.Buffer)
			amfEnc := rtmpmsg.NewAMFEncoder(amdBuf, rtmpmsg.EncodingTypeAMF0)
			if err := rtmpmsg.EncodeBodyAnyValues(amfEnc, &rtmpmsg.NetStreamSetDataFrame{
				Payload: buf.Bytes(),
			}); err != nil {
				return err
			}

			// A "@setDataFrame" DataMessage is written to the RTMP connection.
			// The hardcoded chunk stream ID (8) and "@setDataFrame" may need to be replaced with dynamic values.
			ctx := context.Background()
			chunkStreamID := 8
			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
				StreamID: streamID,
				Message: &rtmpmsg.DataMessage{
					Name:     "@setDataFrame",
					Encoding: rtmpmsg.EncodingTypeAMF0,
					Body:     amdBuf,
				},
			})

		// If the event data doesn't match any of the previous types, an error is thrown.
		default:
			panic("unreachable")
		}
	}
}
