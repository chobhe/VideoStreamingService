package main

import (
	"io"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/yutopp/go-rtmp"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	// Create the relay service for all the streams. This is so we can handle multiple streams.
	// Each stream maps to a pubsub
	relayService := NewRelayService()
	relayService.httpClient = &http.Client{}

	// Create a new server
	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			l := log.StandardLogger()
			//l.SetLevel(logrus.DebugLevel)

			h := &Handler{
				relayService: relayService,
			}

			return conn, &rtmp.ConnConfig{
				Handler: h,

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},

				Logger: l,
			}
		},
	})
	// Listens on this port for each of the RTMP Commands then eventhandles them appropriately
	// Serve calls the OnConnect which creates a logger and handler that references our relayservice
	log.Info("RTMP Server started. Listening on :1935")
	if err := srv.Serve(listener); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}
