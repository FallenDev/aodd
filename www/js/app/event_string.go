// generated by stringer -type=event; DO NOT EDIT

package main

import "fmt"

const _event_name = "EV_ERROREV_CONNECTEDEV_ACTOR_DOESNT_EXISTEV_ACTOR_EXISTSEV_AUTH_FAILEDEV_LOGIN_SUCCESSEV_CREATE_SUCCESSEV_RECV_INPUT_CONNEV_RECV_INITIAL_STATEEV_RECV_UPDATEEV_PACKETEV_SIZE"

var _event_index = [...]uint8{0, 8, 20, 41, 56, 70, 86, 103, 121, 142, 156, 165, 172}

func (i event) String() string {
	if i < 0 || i+1 >= event(len(_event_index)) {
		return fmt.Sprintf("event(%d)", i)
	}
	return _event_name[_event_index[i]:_event_index[i+1]]
}
