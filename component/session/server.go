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

	"github.com/xjasonlyu/tun2socks/common/queue"
	C "github.com/xjasonlyu/tun2socks/constant"
)

const maxCompletedSessions = 100

var (
	ServeAddr = "localhost:6001"
	ServePath = "/session/plain"
)

type Server struct {
	server                *http.Server
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
		if sess, ok := item.(*Session); ok {
			completedSessions = append(completedSessions, sess)
		}
	}

	tablePrint := func(w io.Writer, sessions []*Session) {
		// Sort by session start time.
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.Sub(sessions[j].SessionStart) < 0
		})
		_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
		_, _ = fmt.Fprintf(w, "<tr><td>Process</td><td>Network</td><td>Date</td><td>Duration</td><td>Client Addr</td><td>Target Addr</td><td>Upload</td><td>Download</td></tr>")
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.After(sessions[j].SessionStart)
		})

		for _, sess := range sessions {
			_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>",
				sess.Process,
				sess.Network,
				date(sess.SessionStart),
				duration(sess.SessionStart, sess.SessionClose),
				// sess.DialerAddr,
				sess.ClientAddr,
				sess.TargetAddr,
				byteCountSI(atomic.LoadInt64(&sess.UploadBytes)),
				byteCountSI(atomic.LoadInt64(&sess.DownloadBytes)),
			)
		}
		_, _ = fmt.Fprintf(w, "</table>")
	}

	w := bufio.NewWriter(resp)
	_, _ = fmt.Fprintf(w, "<html>")
	_, _ = fmt.Fprintf(w, `<head><style>
table, th, td {
  border: 1px solid black;
  border-collapse: collapse;
  text-align: right;
  padding: 4;
}</style><title>Go-tun2socks Sessions</title></head>`)
	_, _ = fmt.Fprintf(w, "<h2>Go-tun2socks %s</h2>", C.Version)
	_, _ = fmt.Fprintf(w, "<h3>Now: %s ; Uptime: %s</h3>", now(), uptime())
	_, _ = fmt.Fprintf(w, "<p>Active sessions %d</p>", len(activeSessions))
	tablePrint(w, activeSessions)
	_, _ = fmt.Fprintf(w, "<br/><br/>")
	_, _ = fmt.Fprintf(w, "<p>Recently completed sessions %d</p>", len(completedSessions))
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
	s.server = &http.Server{Addr: ServeAddr, Handler: mux}
	go func() {
		s.server.Serve(c)
	}()

	return nil
}

func (s *Server) Stop() error {
	return s.server.Close()
}

func (s *Server) AddSession(key interface{}, session *Session) {
	s.activeSessionMap.Store(key, session)
}

func (s *Server) GetSession(key interface{}) *Session {
	if sess, ok := s.activeSessionMap.Load(key); ok {
		return sess.(*Session)
	}
	return nil
}

func (s *Server) RemoveSession(key interface{}) {
	if sess, ok := s.activeSessionMap.Load(key); ok {
		// move to completed sessions
		s.completedSessionQueue.Put(sess)
		if s.completedSessionQueue.Len() > maxCompletedSessions {
			s.completedSessionQueue.Pop()
		}
		// delete
		s.activeSessionMap.Delete(key)
	}
}
