package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xjasonlyu/tun2socks/common/queue"
	C "github.com/xjasonlyu/tun2socks/constant"
)

const maxCompletedSessions = 100

var (
	ServeAddr = "localhost:6001"
	ServePath = "/session/plain"
)

type Server struct {
	*http.Server

	trafficUp   int64
	trafficDown int64

	activeSessionMap      sync.Map
	completedSessionQueue *queue.Queue
}

func NewServer() *Server {
	return &Server{
		completedSessionQueue: queue.New(maxCompletedSessions),
	}
}

func (s *Server) handler(resp http.ResponseWriter, req *http.Request) {
	// Slice of active sessions
	var activeSessions []*Session
	s.activeSessionMap.Range(func(key, value interface{}) bool {
		activeSessions = append(activeSessions, value.(*Session))
		return true
	})

	// Slice of completed sessions
	var completedSessions []*Session
	for _, item := range s.completedSessionQueue.Copy() {
		if session, ok := item.(*Session); ok {
			completedSessions = append(completedSessions, session)
		}
	}

	tablePrint := func(w io.Writer, sessions []*Session) {
		// Sort by session start time.
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.Sub(sessions[j].SessionStart) < 0
		})
		_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
		_, _ = fmt.Fprintf(w, "<tr><th>Process</th><th>Network</th><th>Date</th><th>Duration</th><th>Client Addr</th><th>Target Addr</th><th>Upload</th><th>Download</th></tr>\n")
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.After(sessions[j].SessionStart)
		})

		for _, session := range sessions {
			_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>\n",
				session.Process,
				session.Network,
				date(session.SessionStart),
				duration(session.SessionStart, session.SessionClose),
				// session.DialerAddr,
				session.ClientAddr,
				session.TargetAddr,
				byteCountSI(atomic.LoadInt64(&session.UploadBytes)),
				byteCountSI(atomic.LoadInt64(&session.DownloadBytes)),
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
	_, _ = fmt.Fprintf(w, "<p>Statistics</p>")
	_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
	_, _ = fmt.Fprintf(w, "<tr><th align=\"center\">Latest Refresh</th><th>Uptime</th><th>Total Traffic</th><th>Upload</th><th>Download</th></tr>\n")
	trafficUp := atomic.LoadInt64(&s.trafficUp)
	trafficDown := atomic.LoadInt64(&s.trafficDown)
	_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
		date(time.Now()),
		uptime(),
		byteCountSI(trafficUp+trafficDown),
		byteCountSI(trafficUp),
		byteCountSI(trafficDown),
	)
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
	_, port, err := net.SplitHostPort(ServeAddr)
	if port == "0" || port == "" || err != nil {
		return errors.New("address format error")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", ServeAddr)
	if err != nil {
		return err
	}

	c, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ServePath, 301)
	})
	mux.HandleFunc(ServePath, s.handler)
	s.Server = &http.Server{Addr: ServeAddr, Handler: mux}
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
		s.completedSessionQueue.Put(session)
		if s.completedSessionQueue.Len() > maxCompletedSessions {
			s.completedSessionQueue.Pop()
		}
	}
}
