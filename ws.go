package surrealdb

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type WS struct {
	ws   *websocket.Conn        // websocket connection
	quit chan error             // stops: MAIN LOOP
	send chan<- *RPCRequest     // sender channel
	recv <-chan *RawRPCResponse // receive channel
	emit struct {
		lock sync.Mutex                            // pause threads to avoid conflicts
		once map[interface{}][]func(error, []byte) // once listeners
		when map[interface{}][]func(error, []byte) // when listeners
	}
}

func NewWebsocket(url string) (*WS, error) {
	dialer := websocket.DefaultDialer
	dialer.EnableCompression = true

	// stablish connection
	so, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	ws := &WS{ws: so}
	// setup loops and channels
	ws.initialise()

	return ws, nil

}

// --------------------------------------------------
// Public methods
// --------------------------------------------------

func (socket *WS) Close() error {

	msg := websocket.FormatCloseMessage(1000, "")
	return socket.ws.WriteMessage(websocket.CloseMessage, msg)

}

func (socket *WS) Send(id string, method string, params []interface{}) {

	request := RPCRequest{
		ID:     id,
		Method: method,
		Params: params,
	}

	fmt.Println("request:", request)

	go func() {
		socket.send <- &RPCRequest{
			ID:     id,
			Method: method,
			Params: params,
		}
	}()

}

// Subscribe to once()
func (socket *WS) Once(id, _ string) (<-chan []byte, <-chan error) {

	err := make(chan error)
	res := make(chan []byte)

	socket.once(id, func(e error, r []byte) {
		switch {
		case e != nil:
			err <- e
			close(err)
			close(res)
		case e == nil:
			res <- r
			close(err)
			close(res)
		}
	})

	return res, err

}

// Subscribe to when()
func (socket *WS) When(id, _ string) (<-chan []byte, <-chan error) {

	err := make(chan error)
	res := make(chan []byte)

	socket.when(id, func(e error, r []byte) {
		switch {
		case e != nil:
			err <- e
		case e == nil:
			res <- r
		}
	})

	return res, err

}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

func (socket *WS) once(id interface{}, fn func(error, []byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts
	socket.emit.lock.Lock()
	defer socket.emit.lock.Unlock()

	// if its our first listener, we need to setup the map
	if socket.emit.once == nil {
		socket.emit.once = make(map[interface{}][]func(error, []byte))
	}

	socket.emit.once[id] = append(socket.emit.once[id], fn)

}

// WHEN SYSTEM ISN'T BEING USED, MAYBE FOR FUTURE IN-DATABASE EVENTS AND/OR REAL TIME stuffs.

func (socket *WS) when(id interface{}, fn func(error, []byte)) {

	// pauses traffic in others threads, so we can add the new listener without conflicts
	socket.emit.lock.Lock()
	defer socket.emit.lock.Unlock()

	// if its our first listener, we need to setup the map
	if socket.emit.when == nil {
		socket.emit.when = make(map[interface{}][]func(error, []byte))
	}

	socket.emit.when[id] = append(socket.emit.when[id], fn)

}

func (socket *WS) done(id string, err error, result []byte) {

	// pauses traffic in others threads, so we can modify listeners without conflicts
	socket.emit.lock.Lock()
	defer socket.emit.lock.Unlock()

	// if our events map exist
	if socket.emit.when != nil {

		// if theres some listener aiming to this id response
		if _, ok := socket.emit.when[id]; ok {

			// dispatch the event, starting from the end, so we prioritize the new ones
			for i := len(socket.emit.when[id]) - 1; i >= 0; i-- {

				// invoke callback
				socket.emit.when[id][i](err, result)

			}
		}
	}

	// if our events map exist
	if socket.emit.once != nil {

		// if theres some listener aiming to this id response
		if _, ok := socket.emit.once[id]; ok {

			// dispatch the event, starting from the end, so we prioritize the new ones
			for i := len(socket.emit.once[id]) - 1; i >= 0; i-- {

				// invoke callback
				socket.emit.once[id][i](err, result)

				// erase this listener
				socket.emit.once[id][i] = nil

				// remove this listener from the list
				socket.emit.once[id] = socket.emit.once[id][:i]
			}
		}
	}

}

func (socket *WS) read() (*RawRPCResponse, error) {
	var err error
	var bytes []byte
	_, bytes, err = socket.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	fmt.Println("read:", string(bytes))

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

func (socket *WS) write(v interface{}) (err error) {

	w, err := socket.ws.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		return err
	}

	return w.Close()

}

func (socket *WS) initialise() {
	send := make(chan *RPCRequest)
	recv := make(chan *RawRPCResponse)
	quit := make(chan error, 1) // stops: MAIN LOOP
	exit := make(chan int, 1)   // stops: RECEIVER LOOP, SENDER LOOP

	// RECEIVER LOOP

	go func() {
	loop:
		for {
			select {
			case <-exit:
				break loop // stops: THIS LOOP
			default:

				response, err := socket.read() // wait and unmarshal UPCOMING response

				if err != nil {
					_ = socket.Close()
					quit <- err // stops: MAIN LOOP
					exit <- 0   // stops: RECEIVER LOOP, SENDER LOOP
					break loop  // stops: THIS LOOP
				}

				recv <- response // redirect response to: MAIN LOOP
			}
		}
	}()

	// SENDER LOOP

	go func() {
	loop:
		for {
			select {
			case <-exit:
				break loop // stops: THIS LOOP
			case res := <-send:

				err := socket.write(res) // marshal and send

				if err != nil {
					_ = socket.Close()
					quit <- err // stops: MAIN LOOP
					exit <- 0   // stops: RECEIVER LOOP, SENDER LOOP
					break loop  // stops: THIS LOOP
				}

			}
		}
	}()

	// MAIN LOOP

	go func() {
		for {
			select {
			case <-socket.quit:
				break
			case res := <-socket.recv:
				switch {
				case res.Error == nil:
					socket.done(res.ID, nil, res.Result)
				case res.Error != nil:
					socket.done(res.ID, res.Error, res.Result)
				}
			}
		}
	}()

	socket.send = send
	socket.recv = recv
	socket.quit = quit // stops: MAIN LOOP
}
