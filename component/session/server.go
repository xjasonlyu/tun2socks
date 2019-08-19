package session

import (
	"bufio"
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
)

const maxCompletedSessions = 100

type Server struct {
	sync.Mutex
	*http.Server

	ServeAddr string
	ServePath string

	trafficUp   int64
	trafficDown int64

	activeSessionMap  sync.Map
	completedSessions []Session
}

func New(addr string) *Server {
	return &Server{
		ServeAddr: addr,
		ServePath: "/session/plain",
	}
}

func (s *Server) handler(resp http.ResponseWriter, req *http.Request) {
	// Slice of active sessions
	var activeSessions []Session
	s.activeSessionMap.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		activeSessions = append(activeSessions, *session)
		return true
	})

	// Slice of completed sessions
	s.Lock()
	completedSessions := append([]Session(nil), s.completedSessions...)
	s.Unlock()

	tablePrint := func(w io.Writer, sessions []Session) {
		// Sort by session start time.
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.After(sessions[j].SessionStart)
		})
		_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
		_, _ = fmt.Fprintf(w, "<tr><th>Process</th><th>Network</th><th>Date</th><th>Duration</th><th>Client Addr</th><th>Target Addr</th><th>Upload</th><th>Download</th></tr>\n")

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
		_, _ = fmt.Fprintf(w, "</table>")
	}

	w := bufio.NewWriter(resp)
	// Html head
	_, _ = fmt.Fprintf(w, "<html>")
	_, _ = fmt.Fprintf(w, `<head><style>
table, th, td {
  border: 1px solid black;
  border-collapse: collapse;
  text-align: right;
  padding: 4;
}</style><title>Go-tun2socks Monitor</title></head>`)
	_, _ = fmt.Fprintf(w, "<h2>Go-tun2socks %s</h2>", C.Version)

	// Statistics table
	_, _ = fmt.Fprintf(w, "<p>Statistics (%d)</p>", runtime.NumGoroutine())
	_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
	_, _ = fmt.Fprintf(w, "<tr><th>Last Refresh Time</th><th>Platform Version</th><th>CPU</th><th>MEM</th><th>Uptime</th><th>Total</th><th>Upload</th><th>Download</th></tr>\n")
	// calculate traffic
	trafficUp := atomic.LoadInt64(&s.trafficUp)
	trafficDown := atomic.LoadInt64(&s.trafficDown)
	for _, session := range activeSessions {
		trafficUp += session.UploadBytes
		trafficDown += session.DownloadBytes
	}
	_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
		date(time.Now()),
		platform(),
		cpu(),
		mem(),
		uptime(),
		byteCountSI(trafficUp+trafficDown),
		byteCountSI(trafficUp),
		byteCountSI(trafficDown),
	)
	runtime.NumGoroutine()
	_, _ = fmt.Fprintf(w, "</table>")

	// Session table
	_, _ = fmt.Fprintf(w, "<p>Active sessions: %d</p>", len(activeSessions))
	tablePrint(w, activeSessions)
	_, _ = fmt.Fprintf(w, "<p>Recently completed sessions: %d</p>", len(completedSessions))
	tablePrint(w, completedSessions)
	_, _ = fmt.Fprintf(w, "</html>")
	_ = w.Flush()
}

func (s *Server) Start() error {
	if s.ServePath == "" || s.ServePath == "/" {
		return errors.New("invalid serve path")
	}

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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.ServePath, 301)
	})
	mux.HandleFunc(s.ServePath, s.handler)
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
		// move to completed sessions
		s.Lock()
		s.completedSessions = append(s.completedSessions, *session)
		if len(s.completedSessions) > maxCompletedSessions {
			s.completedSessions = s.completedSessions[1:]
		}
		s.Unlock()
	}
}
