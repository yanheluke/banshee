// Copyright 2016 Eleme Inc. All rights reserved.

package webapp

import (
	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage/eventdb"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

func getEventsByProjectID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Options
	n, err := strconv.Atoi(r.URL.Query().Get("past"))
	if err != nil {
		n = 3600 * 24 // 1 day
	}
	past := uint32(n)
	if past > cfg.Expiration {
		ResponseError(w, ErrEventPast)
		return
	}
	level, err := strconv.Atoi(r.URL.Query().Get("level"))
	if err != nil {
		level = 0 // low
	}
	if err := models.ValidateRuleLevel(level); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}
	// Params
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ResponseError(w, ErrProjectID)
		return
	}
	end := uint32(time.Now().Unix())
	start := end - past
	ews, err := db.Event.GetByProjectID(id, level, start, end)
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if ews == nil {
		ews = make([]eventdb.EventWrapper, 0) // Note: Avoid js null
	}
	// Reverse
	for i, j := 0, len(ews)-1; i < j; i, j = i+1, j-1 {
		ews[i], ews[j] = ews[j], ews[i]
	}
	ResponseJSONOK(w, ews)
}

func getEvents(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	values := r.URL.Query()
	n, err := strconv.Atoi(values.Get("past"))
	if err != nil {
		n = 3600 * 24 // 1 day
	}
	past := uint32(n)
	if past > cfg.Expiration {
		ResponseError(w, ErrEventPast)
		return
	}

	level, err := strconv.Atoi(r.URL.Query().Get("level"))
	if err != nil {
		level = 0 // low
	}
	if err := models.ValidateRuleLevel(level); err != nil {
		ResponseError(w, NewValidationWebError(err))
		return
	}

	now := uint32(time.Now().Unix())
	var end uint32
	ed, err := strconv.Atoi(values.Get("end"))
	if err != nil {
		end = now
	} else {
		end = uint32(ed)
	}
	start := end - past
	if start < now-cfg.Expiration {
		ResponseError(w, ErrEventTimeRange)
		return
	}

	ews, err := db.Event.GetRange(level, start, end)
	if err != nil {
		ResponseError(w, NewUnexceptedWebError(err))
		return
	}
	if ews == nil {
		ews = make([]eventdb.EventWrapper, 0)
	}
	ResponseJSONOK(w, ews)
}
