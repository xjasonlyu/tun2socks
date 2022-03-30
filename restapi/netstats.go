package restapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"gvisor.dev/gvisor/pkg/tcpip"
)

var _statsFunc func() tcpip.Stats

func SetStatsFunc(s func() tcpip.Stats) {
	_statsFunc = s
}

func getNetStats(w http.ResponseWriter, r *http.Request) {
	if _statsFunc == nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrUninitialized)
		return
	}

	snapshot := func() any {
		s := _statsFunc()
		return dump(reflect.ValueOf(&s).Elem())
	}

	if !websocket.IsWebSocketUpgrade(r) {
		render.JSON(w, r, snapshot())
		return
	}

	conn, err := _upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	buf := &bytes.Buffer{}
	for range tick.C {
		buf.Reset()

		if err = json.NewEncoder(buf).Encode(snapshot()); err != nil {
			break
		}

		if err = conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
			break
		}
	}
}

func dump(value reflect.Value) map[string]any {
	numField := value.NumField()
	structure := make(map[string]any, numField)

	for i := 0; i < numField; i++ {
		field := value.Type().Field(i)
		value := value.Field(i)

		switch v := value.Addr().Interface().(type) {
		case **tcpip.StatCounter:
			structure[field.Name] = (*v).Value()
		case **tcpip.IntegralStatCounterMap:
			counterMap := make(map[uint64]uint64)
			for _, k := range (*v).Keys() {
				if counter, ok := (*v).Get(k); ok {
					counterMap[k] = counter.Value()
				}
			}
			structure[field.Name] = counterMap
		default:
			structure[field.Name] = dump(value)
		}
	}
	return structure
}
