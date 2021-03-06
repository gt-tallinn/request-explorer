// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package command

import (
	"context"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/description"
	"github.com/mongodb/mongo-go-driver/core/options"
	"github.com/mongodb/mongo-go-driver/core/readpref"
	"github.com/mongodb/mongo-go-driver/core/result"
	"github.com/mongodb/mongo-go-driver/core/wiremessage"
)

// Distinct represents the disctinct command.
//
// The distinct command returns the distinct values for a specified field
// across a single collection.
type Distinct struct {
	NS       Namespace
	Field    string
	Query    *bson.Document
	Opts     []options.DistinctOptioner
	ReadPref *readpref.ReadPref

	result result.Distinct
	err    error
}

// Encode will encode this command into a wire message for the given server description.
func (d *Distinct) Encode(desc description.SelectedServer) (wiremessage.WireMessage, error) {
	if err := d.NS.Validate(); err != nil {
		return nil, err
	}

	command := bson.NewDocument(bson.EC.String("distinct", d.NS.Collection), bson.EC.String("key", d.Field))

	if d.Query != nil {
		command.Append(bson.EC.SubDocument("query", d.Query))
	}

	for _, option := range d.Opts {
		if option == nil {
			continue
		}
		option.Option(command)
	}

	return (&Command{DB: d.NS.DB, ReadPref: d.ReadPref, Command: command}).Encode(desc)
}

// Decode will decode the wire message using the provided server description. Errors during decoding
// are deferred until either the Result or Err methods are called.
func (d *Distinct) Decode(desc description.SelectedServer, wm wiremessage.WireMessage) *Distinct {
	rdr, err := (&Command{}).Decode(desc, wm).Result()
	if err != nil {
		d.err = err
		return d
	}

	d.err = bson.Unmarshal(rdr, &d.result)
	return d
}

// Result returns the result of a decoded wire message and server description.
func (d *Distinct) Result() (result.Distinct, error) {
	if d.err != nil {
		return result.Distinct{}, d.err
	}
	return d.result, nil
}

// Err returns the error set on this command.
func (d *Distinct) Err() error { return d.err }

// RoundTrip handles the execution of this command using the provided wiremessage.ReadWriter.
func (d *Distinct) RoundTrip(ctx context.Context, desc description.SelectedServer, rw wiremessage.ReadWriter) (result.Distinct, error) {
	wm, err := d.Encode(desc)
	if err != nil {
		return result.Distinct{}, err
	}

	err = rw.WriteWireMessage(ctx, wm)
	if err != nil {
		return result.Distinct{}, err
	}
	wm, err = rw.ReadWireMessage(ctx)
	if err != nil {
		return result.Distinct{}, err
	}
	return d.Decode(desc, wm).Result()
}
