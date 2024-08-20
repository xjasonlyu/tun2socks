package restapi

import (
	"bytes"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"gvisor.dev/gvisor/pkg/tcpip"
)

var _stackStatsFunc func() tcpip.Stats

func SetStatsFunc(s func() tcpip.Stats) {
	_stackStatsFunc = s
}

func init() {
	registerEndpoint("/netstats", http.HandlerFunc(getNetStats))
}

func getNetStats(w http.ResponseWriter, r *http.Request) {
	if _stackStatsFunc == nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrUninitialized)
		return
	}

	b := &bytes.Buffer{}
	snapshot := func() []byte {
		s := _stackStatsFunc()
		b.Reset() /* reset buffer */
		encodeToJSON(reflect.ValueOf(&s).Elem(), b)
		return b.Bytes()
	}

	if !websocket.IsWebSocketUpgrade(r) {
		w.Header().Set("Content-Type", "application/json")
		render.Status(r, http.StatusOK)
		w.Write(snapshot())
		w.(http.Flusher).Flush()
		return
	}

	conn, err := _upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for range tick.C {
		if err = conn.WriteMessage(websocket.TextMessage, snapshot()); err != nil {
			break
		}
	}
}

func encodeToJSON(value reflect.Value, b *bytes.Buffer) {
	b.WriteByte('{')
	defer b.WriteByte('}')

	for i, numField := 0, value.NumField(); i < numField; i++ {
		field := value.Type().Field(i)
		value := value.Field(i)

		b.WriteString("\"" + field.Name + "\":")

		switch v := value.Addr().Interface().(type) {
		case **tcpip.StatCounter:
			b.WriteString(strconv.FormatUint((*v).Value(), 10))
		case **tcpip.IntegralStatCounterMap:
			b.WriteByte('{')
			for j, keys := 0, (*v).Keys(); j < len(keys); j++ {
				if counter, ok := (*v).Get(keys[j]); ok {
					k := strconv.FormatUint(keys[j], 10)
					v := strconv.FormatUint(counter.Value(), 10)
					b.WriteString("\"" + k + "\":" + v)
					if j < len(keys)-1 {
						b.WriteByte(',')
					}
				}
			}
			b.WriteByte('}')
		default:
			encodeToJSON(value, b)
		}
		if i < numField-1 {
			b.WriteByte(',')
		}
	}
}
