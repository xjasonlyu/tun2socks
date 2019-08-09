package session

import (
	"bufio"
	"fmt"
	"io"
	"net"
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

const maxCompletedSessions = 100

var (
	StatsAddr = "localhost:6001"
	StatsPath = "/stats/session/plain"

	StatsVersion = ""
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
	log.Infof("Start session stater")
	sessionStatsHandler := func(resp http.ResponseWriter, req *http.Request) {
		// Make a snapshot.
		var sessions []stats.Session
		s.sessions.Range(func(key, value interface{}) bool {
			sess := value.(*stats.Session)
			// check conn is closed or not
			if sess.Network == "tcp" {
				conn := key.(net.Conn)
				if isClosed(conn) {
					s.RemoveSession(conn)
					return true
				}
			}
			sessions = append(sessions, *sess)
			return true
		})

		p := message.NewPrinter(language.English)
		tablePrint := func(w io.Writer, sessions []stats.Session) {
			// Sort by session start time.
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].SessionStart.Sub(sessions[j].SessionStart) < 0
			})

			_, _ = fmt.Fprintf(w, "<table style=\"border=4px solid\">")
			_, _ = fmt.Fprintf(w, "<tr><td>Process Name</td><td>Network</td><td>Duration</td><td>Dialer Addr</td><td>Client Addr</td><td>Target Addr</td><td>Upload Bytes</td><td>Download Bytes</td></tr>")
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].SessionStart.After(sessions[j].SessionStart)
			})
			for _, sess := range sessions {
				_, _ = fmt.Fprintf(w, "<tr><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td></tr>",
					sess.ProcessName,
					sess.Network,
					time.Now().Sub(sess.SessionStart).Round(time.Second),
					sess.DialerAddr,
					sess.ClientAddr,
					sess.TargetAddr,
					p.Sprintf("%d", atomic.LoadInt64(&sess.UploadBytes)),
					p.Sprintf("%d", atomic.LoadInt64(&sess.DownloadBytes)),
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
	s.server = &http.Server{Addr: StatsAddr, Handler: mux}
	go s.server.ListenAndServe()
	return nil
}

func (s *simpleSessionStater) Stop() error {
	log.Infof("Stop session stater")
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
		// move to completed sessions
		s.mux.Lock()
		s.completedSessions = append(s.completedSessions, *(sess.(*stats.Session)))
		if len(s.completedSessions) > maxCompletedSessions {
			s.completedSessions = s.completedSessions[1:]
		}
		s.mux.Unlock()
		// delete
		s.sessions.Delete(key)
	}
}
