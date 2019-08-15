package session

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xjasonlyu/tun2socks/common/queue"
	"github.com/xjasonlyu/tun2socks/component/stats"
	C "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/log"
)

const maxCompletedSessions = 100

var (
	ServeAddr = "localhost:6001"
	ServePath = "/stats/session/plain"
)

type simpleSessionStater struct {
	server                *http.Server
	activeSessionMap      sync.Map
	completedSessionQueue *queue.Queue
}

func NewSimpleSessionStater() stats.SessionStater {
	return &simpleSessionStater{
		completedSessionQueue: queue.New(maxCompletedSessions),
	}
}

func (s *simpleSessionStater) sessionStatsHandler(resp http.ResponseWriter, req *http.Request) {
	// Slice of active sessions
	var activeSessions []*stats.Session
	s.activeSessionMap.Range(func(key, value interface{}) bool {
		activeSessions = append(activeSessions, value.(*stats.Session))
		return true
	})

	// Slice of completed sessions
	var completedSessions []*stats.Session
	for _, item := range s.completedSessionQueue.Copy() {
		if sess, ok := item.(*stats.Session); ok {
			completedSessions = append(completedSessions, sess)
		}
	}

	tablePrint := func(w io.Writer, sessions []*stats.Session) {
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

func (s *simpleSessionStater) Start() error {
	log.Debugf("Start session stater")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ServePath, 301)
	})
	mux.HandleFunc(ServePath, s.sessionStatsHandler)
	s.server = &http.Server{Addr: ServeAddr, Handler: mux}
	go s.server.ListenAndServe()
	return nil
}

func (s *simpleSessionStater) Stop() error {
	log.Debugf("Stop session stater")
	return s.server.Close()
}

func (s *simpleSessionStater) AddSession(key interface{}, session *stats.Session) {
	s.activeSessionMap.Store(key, session)
}

func (s *simpleSessionStater) GetSession(key interface{}) *stats.Session {
	if sess, ok := s.activeSessionMap.Load(key); ok {
		return sess.(*stats.Session)
	}
	return nil
}

func (s *simpleSessionStater) RemoveSession(key interface{}) {
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
