package surrealdb

import (
	"context"
	"encoding/json"
	"github.com/buger/jsonparser"
	"sync"

	"github.com/gorilla/websocket"
)

type WS struct {
	ws   *websocket.Conn        // websocket connection
	send chan<- *RPCRequest     // sender channel
	recv <-chan *RawRPCResponse // receive channel
	emit struct {
		// TODO: use the lock less, through smaller locks (separate once/when locks ?)
		// or ideally by removing locks altogether
		lock sync.Mutex // pause threads to avoid conflicts

		// do the callbacks really need to be a list ?
		once map[interface{}][]func([]byte) // once listeners
		when map[interface{}][]func([]byte) // when listeners
	}
}

type responseValue struct {
	Value []byte
}

func NewWebsocket(ctx *context.Context, url string) (*WS, error) {
	dialer := websocket.DefaultDialer
	dialer.EnableCompression = true

	// stablish connection
	so, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	ws := &WS{ws: so}

	// initilialize the callback maps here so we don't need to check them at runtime
	ws.emit.once = make(map[interface{}][]func([]byte))
	ws.emit.when = make(map[interface{}][]func([]byte))

	// setup loops and channels
	ws.initialise(*ctx)

	return ws, nil

}

// --------------------------------------------------
// Public methods
// --------------------------------------------------

func (ws *WS) Close() error {

	msg := websocket.FormatCloseMessage(1000, "")
	return ws.ws.WriteMessage(websocket.CloseMessage, msg)

}

func (ws *WS) Send(id string, method string, params []interface{}) {

	go func() {
		ws.send <- &RPCRequest{
			ID:     id,
			Method: method,
			Params: params,
		}
	}()

}

// Once subscribes to once()
func (ws *WS) Once(id, _ string) <-chan responseValue {

	out := make(chan responseValue)

	ws.once(id, func(r []byte) {
		out <- responseValue{
			Value: r,
		}
		close(out)
	})

	return out

}

// When subscribes to when()
func (ws *WS) When(id, _ string) <-chan responseValue {
	// TODO: make this cancellable (use of context.Context ?)

	out := make(chan responseValue)

	ws.when(id, func(r []byte) {
		out <- responseValue{
			Value: r,
		}
	})

	return out

}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

func (ws *WS) once(id interface{}, fn func([]byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts

	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	ws.emit.once[id] = append(ws.emit.once[id], fn)

}

// WHEN SYSTEM ISN'T BEING USED, MAYBE FOR FUTURE IN-DATABASE EVENTS AND/OR REAL TIME stuffs.

func (ws *WS) when(id interface{}, fn func([]byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts
	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	ws.emit.when[id] = append(ws.emit.when[id], fn)

}

func (ws *WS) done(r *RawRPCResponse) {

	// pauses traffic in others threads, so we can modify listeners without conflicts
	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	// if theres some listener aiming to this id response
	if when, ok := ws.emit.when[r.ID]; ok {

		// dispatch the event, starting from the end, so we prioritize the new ones
		for i := len(when) - 1; i >= 0; i-- {

			// invoke callback
			when[i](r.Result)

		}
	}

	// if theres some listener aiming to this id response
	if once, ok := ws.emit.once[r.ID]; ok {

		// dispatch the event, starting from the end, so we prioritize the new ones
		for i := len(once) - 1; i >= 0; i-- {

			// invoke callback
			once[i](r.Result)

			// erase this listener
			once[i] = nil

		}

		// remove all listeners
		ws.emit.once[r.ID] = once[0:]
	}
}

func (ws *WS) read() (*RawRPCResponse, error) {
	_, bytes, err := ws.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	id, err := jsonparser.GetString(bytes, "id")

	return &RawRPCResponse{
		ID:     id,
		Result: bytes,
	}, nil
}

func (ws *WS) write(v interface{}) (err error) {

	w, err := ws.ws.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	// the default HTML escaping messes with select arrows
	enc.SetEscapeHTML(false)
	err = enc.Encode(v)
	if err != nil {
		return err
	}

	return w.Close()

}

func (ws *WS) initialise(ctx context.Context) {
	send := make(chan *RPCRequest)
	recv := make(chan *RawRPCResponse)
	ctx, cancel := context.WithCancel(ctx)

	// RECEIVER LOOP

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:

				response, err := ws.read() // wait and unmarshal UPCOMING response

				if err != nil {
					_ = ws.Close()
					cancel()
					return
				}

				recv <- response // redirect response to: MAIN LOOP
			}
		}
	}()

	// SENDER LOOP

	go func() {
		for {
			select {
			case <-ctx.Done():
				return // stops: THIS LOOP
			case res := <-send:

				err := ws.write(res) // marshal and send

				if err != nil {
					_ = ws.Close()
					cancel()
					return // stops: THIS LOOP
				}

			}
		}
	}()

	// MAIN LOOP

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case response := <-ws.recv:
				ws.done(response)
			}
		}
	}()

	ws.send = send
	ws.recv = recv
}
