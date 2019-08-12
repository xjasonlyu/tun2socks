package session

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/stats"
)

const maxCompletedSessions = 100

var (
	StatsAddr = "localhost:6001"
	StatsPath = "/stats/session/plain"

	StatsVersion = ""
)

type simpleSessionStater struct {
	sync.Mutex

	server            *http.Server
	activeSessionMap  sync.Map
	completedSessions []*stats.Session
}

func NewSimpleSessionStater() stats.SessionStater {
	return &simpleSessionStater{}
}

func (s *simpleSessionStater) Start() error {
	log.Debugf("Start session stater")
	sessionStatsHandler := func(resp http.ResponseWriter, req *http.Request) {
		// Make a snapshot.
		var activeSessions []*stats.Session
		s.activeSessionMap.Range(func(key, value interface{}) bool {
			sess := value.(*stats.Session)
			activeSessions = append(activeSessions, sess)
			return true
		})

		tablePrint := func(w io.Writer, sessions []*stats.Session) {
			// Sort by session start time.
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].SessionStart.Sub(sessions[j].SessionStart) < 0
			})
			_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
			_, _ = fmt.Fprintf(w, "<tr><td>Process Name</td><td>Network</td><td>Date</td><td>Duration</td><td>Client Addr</td><td>Target Addr</td><td>Upload</td><td>Download</td></tr>")
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].SessionStart.After(sessions[j].SessionStart)
			})

			for _, sess := range sessions {
				_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>",
					sess.ProcessName,
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
		_, _ = fmt.Fprintf(w, `<head><style>table, th, td {
  border: 1px solid black;
  border-collapse: collapse;
  text-align: right;
  padding: 4;
}</style><title>Go-tun2socks Sessions</title></head>`)
		_, _ = fmt.Fprintf(w, "<h2>Go-tun2socks %s</h2>", StatsVersion)
		_, _ = fmt.Fprintf(w, "<h3>Now: %s ; Uptime: %s</h3>", now(), uptime())
		_, _ = fmt.Fprintf(w, "<p>Active sessions %d</p>", len(activeSessions))
		tablePrint(w, activeSessions)
		_, _ = fmt.Fprintf(w, "<br/><br/>")
		_, _ = fmt.Fprintf(w, "<p>Recently completed sessions %d</p>", len(s.completedSessions))
		tablePrint(w, s.completedSessions)
		_, _ = fmt.Fprintf(w, "</html>")
		_ = w.Flush()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, StatsPath, 301)
	})
	mux.HandleFunc(StatsPath, sessionStatsHandler)
	s.server = &http.Server{Addr: StatsAddr, Handler: mux}
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
		s.Lock()
		s.completedSessions = append(s.completedSessions, sess.(*stats.Session))
		if len(s.completedSessions) > maxCompletedSessions {
			s.completedSessions = s.completedSessions[1:]
		}
		s.Unlock()
		// delete
		s.activeSessionMap.Delete(key)
	}
}
