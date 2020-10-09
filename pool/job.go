// Copyright (c) 2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pool

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Job represents cached copies of work delivered to clients.
type Job struct {
	UUID   string `json:"uuid"`
	Height uint32 `json:"height"`
	Header string `json:"header"`
}

// nanoToBigEndianBytes returns an 8-byte big endian representation of
// the provided nanosecond time.
func nanoToBigEndianBytes(nano int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(nano))
	return b
}

// jobID generates a unique job id of the provided block height.
func jobID(height uint32) string {
	var buf bytes.Buffer
	_, _ = buf.Write(heightToBigEndianBytes(height))
	_, _ = buf.Write(nanoToBigEndianBytes(time.Now().UnixNano()))
	return hex.EncodeToString(buf.Bytes())
}

// NewJob creates a job instance.
func NewJob(header string, height uint32) *Job {
	return &Job{
		UUID:   jobID(height),
		Height: height,
		Header: header,
	}
}

// fetchJobBucket is a helper function for getting the job bucket.
func fetchJobBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	const funcName = "fetchJobBucket"
	pbkt := tx.Bucket(poolBkt)
	if pbkt == nil {
		desc := fmt.Sprintf("%s: bucket %s not found", funcName,
			string(poolBkt))
		return nil, dbError(ErrBucketNotFound, desc)
	}
	bkt := pbkt.Bucket(jobBkt)
	if bkt == nil {
		desc := fmt.Sprintf("%s: bucket %s not found", funcName,
			string(jobBkt))
		return nil, dbError(ErrBucketNotFound, desc)
	}
	return bkt, nil
}

// FetchJob fetches the job referenced by the provided id.
func FetchJob(db *bolt.DB, id string) (*Job, error) {
	const funcName = "FetchJob"
	var job Job
	err := db.View(func(tx *bolt.Tx) error {
		bkt, err := fetchJobBucket(tx)
		if err != nil {
			return err
		}

		v := bkt.Get([]byte(id))
		if v == nil {
			desc := fmt.Sprintf("%s: no job found for id %s", funcName, id)
			return dbError(ErrValueNotFound, desc)
		}
		err = json.Unmarshal(v, &job)
		if err != nil {
			desc := fmt.Sprintf("%s: unable to unmarshal job bytes: %v",
				funcName, err)
			return dbError(ErrParse, desc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &job, err
}

// Persist saves the job to the database.
func (job *Job) Persist(db *bolt.DB) error {
	const funcName = "Job.Persist"
	return db.Update(func(tx *bolt.Tx) error {
		bkt, err := fetchJobBucket(tx)
		if err != nil {
			return err
		}

		jobBytes, err := json.Marshal(job)
		if err != nil {
			desc := fmt.Sprintf("%s: unable to marshal job bytes: %v",
				funcName, err)
			return dbError(ErrParse, desc)
		}
		err = bkt.Put([]byte(job.UUID), jobBytes)
		if err != nil {
			desc := fmt.Sprintf("%s: unable to persist job entry: %v",
				funcName, err)
			return dbError(ErrPersistEntry, desc)
		}
		return nil
	})
}

// Delete removes the associated job from the database.
func (job *Job) Delete(db *bolt.DB) error {
	return deleteEntry(db, jobBkt, job.UUID)
}
