package chaos

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
)

func Request(addr string, message any) ([]byte, error) {
	var resp []byte
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return resp, err
	}
	data, err := EncodeMessage(message)
	if err != nil {
		return resp, err
	}
	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return resp, err
	}
	return resp, json.NewDecoder(conn).Decode(&resp)
}

func Response(conn net.Conn, req any) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var written int
	for written < len(data) {
		n, err := conn.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}

	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}
	return nil
}
