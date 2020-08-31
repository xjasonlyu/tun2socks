package session

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	C "github.com/xjasonlyu/tun2socks/constant"

	"github.com/gobuffalo/packr/v2"
)

const maxClosedSessions = 100

type Server struct {
	sync.Mutex
	*http.Server

	ServeAddr string

	trafficUp   int64
	trafficDown int64

	activeSessionMap  sync.Map
	closedSessionList []Session
}

func New(addr string) *Server {
	return &Server{
		ServeAddr: addr,
	}
}

func (s *Server) getSessions() (activeSessions, closedSessions []Session) {
	// Slice of active sessions
	s.activeSessionMap.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		activeSessions = append(activeSessions, *session)
		return true
	})

	// Slice of closed sessions
	s.Lock()
	defer s.Unlock()
	closedSessions = append([]Session(nil), s.closedSessionList...)
	return
}

func (s *Server) serveJSON(w http.ResponseWriter, _ *http.Request) {
	activeSessions, closedSessions := s.getSessions()

	// calculate traffic
	trafficUp := atomic.LoadInt64(&s.trafficUp)
	trafficDown := atomic.LoadInt64(&s.trafficDown)
	for _, session := range activeSessions {
		trafficUp += session.UploadBytes
		trafficDown += session.DownloadBytes
	}

	status := &Status{
		platform(),
		C.Version,
		cpu(),
		mem(),
		uptime(),
		trafficUp + trafficDown,
		trafficUp,
		trafficDown,
		runtime.NumGoroutine(),
		activeSessions,
		closedSessions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) serveHTML(resp http.ResponseWriter, _ *http.Request) {
	activeSessions, closedSessions := s.getSessions()

	tablePrint := func(w io.Writer, sessions []Session) {
		// Sort by session start time.
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.After(sessions[j].SessionStart)
		})
		_, _ = fmt.Fprintf(w, "<div class=\"table-responsive\"><table class=\"table table-bordered\">\n")
		_, _ = fmt.Fprintf(w, "<thead class=\"thead-light\"><tr><th>Process</th><th>Network</th><th>Date</th><th>Duration</th><th>Client Addr</th><th>Target Addr</th><th>Upload</th><th>Download</th></tr></thead><tbody>\n")

		for _, session := range sessions {
			_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
				session.Process,
				session.Network,
				date(session.SessionStart),
				duration(session.SessionStart, session.SessionClose),
				// session.DialerAddr,
				session.ClientAddr,
				session.TargetAddr,
				byteCountSI(session.UploadBytes),
				byteCountSI(session.DownloadBytes),
			)
		}
		_, _ = fmt.Fprintf(w, "</tbody></table></div>\n")
	}

	w := bufio.NewWriter(resp)
	// Html head
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Go-tun2socks</title>
    <link href="css/bootstrap.min.css" rel="stylesheet">
  </head>
  <body>
    <div class="container">
    <h1>Go-tun2socks %s</h1>`, C.Version)
	// calculate traffic
	trafficUp := atomic.LoadInt64(&s.trafficUp)
	trafficDown := atomic.LoadInt64(&s.trafficDown)
	for _, session := range activeSessions {
		trafficUp += session.UploadBytes
		trafficDown += session.DownloadBytes
	}
	// statistics
	_, _ = fmt.Fprintf(w, `<h3 class="sub-header">Statistics (%d)</h3>
      <div class="table-responsive">
        <table class="table table-bordered">
          <thead class="thead-light"><tr><th>Last Refresh Time</th><th>Platform Version</th><th>CPU</th><th>MEM</th><th>Uptime</th><th>Total</th><th>Upload</th><th>Download</th></tr></thead>
          <tbody>
            <tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>
          </tbody>
        </table>
      </div>`,
		runtime.NumGoroutine(),
		date(time.Now()),
		platform(),
		cpu(),
		mem(),
		uptime(),
		byteCountSI(trafficUp+trafficDown),
		byteCountSI(trafficUp),
		byteCountSI(trafficDown),
	)
	// Session table
	_, _ = fmt.Fprintf(w, "<h3 class=\"sub-header\">Active sessions (%d)</h3>\n", len(activeSessions))
	tablePrint(w, activeSessions)
	_, _ = fmt.Fprintf(w, "<h3 class=\"sub-header\">Closed sessions (%d)</h3>\n", len(closedSessions))
	tablePrint(w, closedSessions)
	_, _ = fmt.Fprintf(w, "</div></body></html>\n")
	_ = w.Flush()
}

func (s *Server) Start() error {
	_, port, err := net.SplitHostPort(s.ServeAddr)
	if port == "0" || port == "" || err != nil {
		return errors.New("address format error")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", s.ServeAddr)
	if err != nil {
		return err
	}

	c, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveHTML)
	mux.HandleFunc("/json", s.serveJSON)

	box := packr.New("CSSBox", "./css")
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(box)))

	s.Server = &http.Server{Addr: s.ServeAddr, Handler: mux}
	go func() {
		s.Serve(c)
	}()

	return nil
}

func (s *Server) Stop() error {
	return s.Close()
}

func (s *Server) AddSession(key interface{}, session *Session) {
	if session != nil {
		s.activeSessionMap.Store(key, session)
	}
}

func (s *Server) RemoveSession(key interface{}) {
	if item, ok := s.activeSessionMap.Load(key); ok {
		session := item.(*Session)
		// delete first
		s.activeSessionMap.Delete(key)
		// record up & down traffic
		atomic.AddInt64(&s.trafficUp, atomic.LoadInt64(&session.UploadBytes))
		atomic.AddInt64(&s.trafficDown, atomic.LoadInt64(&session.DownloadBytes))
		// move to closed sessions
		s.Lock()
		s.closedSessionList = append(s.closedSessionList, *session)
		if len(s.closedSessionList) > maxClosedSessions {
			s.closedSessionList = s.closedSessionList[1:]
		}
		s.Unlock()
	}
}
