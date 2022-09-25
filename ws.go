package surrealdb

import (
	"context"
	"encoding/json"
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
		once map[interface{}][]func(error, []byte) // once listeners
		when map[interface{}][]func(error, []byte) // when listeners
	}
}

func NewWebsocket(ctx context.Context, url string) (*WS, error) {
	dialer := websocket.DefaultDialer
	dialer.EnableCompression = true

	// stablish connection
	so, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	ws := &WS{ws: so}

	// initilialize the callback maps here so we don't need to check them at runtime
	ws.emit.once = make(map[interface{}][]func(error, []byte))
	ws.emit.when = make(map[interface{}][]func(error, []byte))

	// setup loops and channels
	ws.initialise(ctx)

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

type responseValue struct {
	value []byte
	err   error
}

// Subscribe to once()
func (ws *WS) Once(id, method string) <-chan responseValue {

	out := make(chan responseValue)

	ws.once(id, func(e error, r []byte) {
		out <- responseValue{
			value: r,
			err:   e,
		}
		close(out)
	})

	return out

}

// Subscribe to when()
func (ws *WS) When(id, _ string) <-chan responseValue {
	// TODO: make this cancellable (use of context.Context ?)

	out := make(chan responseValue)

	ws.when(id, func(e error, r []byte) {
		out <- responseValue{
			value: r,
			err:   e,
		}
	})

	return out

}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

func (ws *WS) once(id interface{}, fn func(error, []byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts

	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	ws.emit.once[id] = append(ws.emit.once[id], fn)

}

// WHEN SYSTEM ISN'T BEING USED, MAYBE FOR FUTURE IN-DATABASE EVENTS AND/OR REAL TIME stuffs.

func (ws *WS) when(id interface{}, fn func(error, []byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts
	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	ws.emit.when[id] = append(ws.emit.when[id], fn)

}

func (ws *WS) done(id string, err error, result []byte) {

	// pauses traffic in others threads, so we can modify listeners without conflicts
	ws.emit.lock.Lock()
	defer ws.emit.lock.Unlock()

	// if our events map exist
	if ws.emit.when != nil {

		// if theres some listener aiming to this id response
		if when, ok := ws.emit.when[id]; ok {

			// dispatch the event, starting from the end, so we prioritize the new ones
			for i := len(when) - 1; i >= 0; i-- {

				// invoke callback
				when[i](err, result)

			}
		}
	}

	// if our events map exist
	if ws.emit.once != nil {

		// if theres some listener aiming to this id response
		if once, ok := ws.emit.once[id]; ok {

			// dispatch the event, starting from the end, so we prioritize the new ones
			for i := len(once) - 1; i >= 0; i-- {

				// invoke callback
				once[i](err, result)

				// erase this listener
				once[i] = nil

			}

			// remove all listeners
			ws.emit.once[id] = once[0:]
		}
	}
}

func (ws *WS) read() (*RawRPCResponse, error) {
	var err error
	var bytes []byte
	_, bytes, err = ws.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	var rawJson map[string]json.RawMessage
	err = json.Unmarshal(bytes, &rawJson)
	if err != nil {
		return nil, ErrInvalidSurrealResponse{Cause: err}
	}

	idBytes, ok := rawJson["id"]
	if !ok {
		return nil, ErrInvalidSurrealResponse{}
	}
	idBytesLen := len(idBytes)
	if idBytesLen < 2 {
		return nil, ErrInvalidSurrealResponse{}
	}
	id := string(idBytes[1:(idBytesLen - 1)])

	var result []byte
	result, _ = rawJson["result"]

	var rpcErr *RPCError
	rpcErrBytes, errOccurred := rawJson["error"]
	if errOccurred {
		var errJson RPCError
		err = json.Unmarshal(rpcErrBytes, &errJson)
		if err != nil {
			return nil, ErrInvalidSurrealResponse{Cause: err}
		}
		rpcErr = &errJson
	}

	return &RawRPCResponse{
		ID:     id,
		Error:  rpcErr,
		Result: result,
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
	recv := make(chan *RPCResponse)
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
					ws.Close()
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
			case res := <-ws.recv:
				switch {
				case res.Error == nil:
					ws.done(res.ID, nil, res.Result)
				case res.Error != nil:
					ws.done(res.ID, res.Error, res.Result)
				}
			}
		}
	}()

	ws.send = send
	ws.recv = recv
}
