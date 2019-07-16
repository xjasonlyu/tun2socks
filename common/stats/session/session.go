package session

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/stats"
)

const maxCompletedSessions = 50

var (
	StatsAddr = "localhost:6001"
	StatsPath = "/stats/session/plain"
)

type simpleSessionStater struct {
	mux               sync.Mutex
	sessions          sync.Map
	completedSessions []stats.Session
	server            *http.Server
}

func NewSimpleSessionStater() stats.SessionStater {
	return &simpleSessionStater{}
}

func (s *simpleSessionStater) Start() error {
	log.Infof("Start session stater.")
	sessionStatsHandler := func(respw http.ResponseWriter, req *http.Request) {
		// Make a snapshot.
		var sessions []stats.Session
		s.sessions.Range(func(key, value interface{}) bool {
			sess := value.(*stats.Session)
			sessions = append(sessions, *sess)
			return true
		})

		// Sort by session start time.
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].SessionStart.Sub(sessions[j].SessionStart) < 0
		})

		p := message.NewPrinter(language.English)
		tablePrint := func(w io.Writer, sessions []stats.Session) {
			_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
			_, _ = fmt.Fprintf(w, "<tr><td>Process Name</td><td>Network</td><td>Duration</td><td>Local Addr</td><td>Remote Addr</td><td>Upload Bytes</td><td>Download Bytes</td></tr>")
			for _, sess := range sessions {
				_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>",
					sess.ProcessName,
					sess.Network,
					time.Now().Sub(sess.SessionStart).Round(time.Second),
					sess.LocalAddr,
					sess.RemoteAddr,
					p.Sprintf("%d", atomic.LoadInt64(&sess.UploadBytes)),
					p.Sprintf("%d", atomic.LoadInt64(&sess.DownloadBytes)),
				)
			}
			_, _ = fmt.Fprintf(w, "</table>")
		}

		w := bufio.NewWriter(respw)
		_, _ = fmt.Fprintf(w, "<html>")
		_, _ = fmt.Fprintf(w, `<head><style>table, th, td {
  border: 1px solid black;
  border-collapse: collapse;
  text-align: right;
  padding: 4;
}</style></head>`)
		_, _ = fmt.Fprintf(w, "<p>Active sessions %d</p>", len(sessions))
		tablePrint(w, sessions)
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
	server := &http.Server{Addr: StatsAddr, Handler: mux}
	s.server = server
	go func() {
		_ = s.server.ListenAndServe()
	}()
	return nil
}

func (s *simpleSessionStater) Stop() error {
	log.Infof("Stop session stater.")
	return s.server.Close()
}

func (s *simpleSessionStater) AddSession(key interface{}, session *stats.Session) {
	s.sessions.Store(key, session)
}

func (s *simpleSessionStater) GetSession(key interface{}) *stats.Session {
	if sess, ok := s.sessions.Load(key); ok {
		return sess.(*stats.Session)
	}
	return nil
}

func (s *simpleSessionStater) RemoveSession(key interface{}) {
	if sess, ok := s.sessions.Load(key); ok {
		// temporary workaround for slice race condition
		s.mux.Lock()
		s.completedSessions = append(s.completedSessions, *(sess.(*stats.Session)))
		if len(s.completedSessions) > maxCompletedSessions {
			s.completedSessions = s.completedSessions[1:]
		}
		s.mux.Unlock()
	}
	s.sessions.Delete(key)
}
