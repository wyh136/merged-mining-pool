package pool

import (
	"encoding/json"
)

// type stratumRequest struct {
// 	stratumRequestDetail
// 	Worker string `json:"worker"`
// }

type stratumRequest struct {
	Id     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func miningNotify(work Work) stratumRequest {
	var request stratumRequest

	params, err := json.Marshal(work)
	logOnError(err)

	request.Method = "mining.notify"
	request.Params = params

	return request
}

func miningSetExtranonce(extranonce string) stratumRequest {
	var request stratumRequest

	// TODO build request; I need a better example

	return request
}

func miningSetDifficulty(difficulty float32) stratumRequest {
	var request stratumRequest

	request.Method = "mining.set_difficulty"

	diff := []float32{difficulty}

	var err error
	request.Params, err = json.Marshal(diff)
	logOnError(err)

	return request
}
