/*
** Copyright [2013-2017] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package carton

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
	"time"
)

var (

	//the state actions available are.
	STATE        = "state"
	CREATE       = "create"
	BOOTSTRAPPED = "bootstrapped"
	DESTROY      = "destroy"
	STATEUP      = "stateup"
	STATEDOWN    = "statedown"

	DONE    = "done"
	RUNNING = "running"
	FAILURE = "failure"

	//the control actions available are.
	CONTROL      = "control"
	STOP         = "stop"
	START        = "start"
	RESTART      = "restart"
	HARD_RESTART = "hard-restart"
	HARD_STOP    = "hard-stop"
	SUSPEND      = "suspend"

	//the operation actions is just one called upgrade
	OPERATIONS = "operations"
	UPGRADE    = "upgrade"

	//snapshot actions
	SNAPSHOT    = "snapshot"
	SNAPCREATE  = "snapcreate"
	SNAPRESTORE = "snaprestore"
	SNAPDELETE  = "snapremove"
	SNAPSAVE    = "snapsave"

	//vmbackup actions
	BACKUPS      = "backup"
	IMAGECREATE  = "backupcreate"
	IMAGEDESTROY = "backupremove"

	NETWORK_UPDATE = "assembly.network.update"

	// disks actions
	DISKS      = "disks"
	ATTACHDISK = "attachdisk"
	DETACHDISK = "detachdisk"
)

type ReqParser struct {
	name string
}

// NewParser returns a new instance of Parser.
func NewReqParser(n string) *ReqParser {
	return &ReqParser{name: n}
}

// ParseRequest parses a request string and returns its MegdProcess representation.
// eg: (state, create) => CreateProcess{}
// After figuring out the process, we operate on it.
func ParseRequest(r *Requests) (MegdProcessor, error) {
	return NewReqParser(r.CatId).ParseRequest(r.Category, r.Action)
}

func (p *ReqParser) ParseRequest(category string, action string) (MegdProcessor, error) {
	switch category {
	case STATE:
		return p.parseState(action)
	case CONTROL:
		return p.parseControl(action)
	case OPERATIONS:
		return p.parseOperations(action)
	case BACKUPS:
		return p.parseBackups(action)
	case SNAPSHOT:
		return p.parseSnapshot(action)
	case DISKS:
		return p.parseDisks(action)
	case DONE:
		return p.parseDone(action)
	default:
		return nil, newParseError([]string{category, action}, []string{STATE, CONTROL, OPERATIONS, SNAPSHOT, DISKS, BACKUPS})
	}
}

func (p *ReqParser) parseState(action string) (MegdProcessor, error) {
	switch action {
	case CREATE:
		return CreateProcess{
			Name: p.name,
		}, nil
	case DESTROY:
		return DestroyProcess{
			Name: p.name,
		}, nil
	case BOOTSTRAPPED:
		return StateupProcess{
			Name: p.name,
		}, nil
	case STATEDOWN:
		return StateupProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{STATE, action}, []string{CREATE, DESTROY, STATEUP, STATEDOWN})
	}
}

func (p *ReqParser) parseControl(action string) (MegdProcessor, error) {
	switch action {
	case START:
		return StartProcess{
			Name: p.name,
		}, nil
	case STOP:
		return StopProcess{
			Name: p.name,
		}, nil
	case SUSPEND:
		return SuspendProcess{
			Name: p.name,
		}, nil
	case RESTART:
		return RestartProcess{
			Name: p.name,
		}, nil
	case HARD_STOP:
		return StopProcess{
			Name: p.name,
			Hard: true,
		}, nil

	case HARD_RESTART:
		return RestartProcess{
			Name: p.name,
			Hard: true,
		}, nil

	default:
		return nil, newParseError([]string{CONTROL, action}, []string{START, STOP, RESTART})
	}
}

func (p *ReqParser) parseOperations(action string) (MegdProcessor, error) {
	switch action {
	case UPGRADE:
		return UpgradeProcess{
			Name: p.name,
		}, nil
	case NETWORK_UPDATE:
		return UpdateNetworkProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{OPERATIONS, action}, []string{UPGRADE})
	}
}

func (p *ReqParser) parseDone(action string) (MegdProcessor, error) {
	switch action {
	case RUNNING:
		return RunningProcess{
			Name: p.name,
		}, nil
	case FAILURE:
		return FailureProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{OPERATIONS, action}, []string{UPGRADE})
	}
}

func (p *ReqParser) parseBackups(action string) (MegdProcessor, error) {
	switch action {
	case IMAGECREATE:
		return ImageCreateProcess{
			Name: p.name,
		}, nil
	case IMAGEDESTROY:
		return ImageDestroyProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{BACKUPS, action}, []string{IMAGECREATE, IMAGEDESTROY})
	}
}

func (p *ReqParser) parseSnapshot(action string) (MegdProcessor, error) {
	switch action {
	case SNAPCREATE:
		return SnapCreateProcess{
			Name: p.name,
		}, nil
	case SNAPRESTORE:
		return SnapRestoreProcess{
			Name: p.name,
		}, nil
	case SNAPDELETE:
		return SnapDestroyProcess{
			Name: p.name,
		}, nil
	case SNAPSAVE:
		return SnapSaveAsProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{SNAPSHOT, action}, []string{SNAPCREATE, SNAPDELETE})
	}
}

func (p *ReqParser) parseDisks(action string) (MegdProcessor, error) {
	switch action {
	case ATTACHDISK:
		return DiskAttachProcess{
			Name: p.name,
		}, nil
	case DETACHDISK:
		return DiskDetachProcess{
			Name: p.name,
		}, nil
	default:
		return nil, newParseError([]string{SNAPSHOT, action}, []string{ATTACHDISK, DETACHDISK})
	}
}

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Found    string
	Expected []string
}

// newParseError returns a new instance of ParseError.
func newParseError(found []string, expected []string) *ParseError {
	return &ParseError{Found: strings.Join(found, ","), Expected: expected}
}

// Error returns the string representation of the error.
func (e *ParseError) Error() string {
	return fmt.Sprintf("found %s, expected %s", e.Found, strings.Join(e.Expected, ", "))
}

type Requests struct {
	Id        string    `json:"id" cql:"id"`
	Name      string    `json:"name" cql:"name"`
	AccountId string    `json:"account_id" cql:"account_id"`
	CatId     string    `json:"cat_id" cql:"cat_id"`
	Action    string    `json:"action" cql:"action"`
	Category  string    `json:"category" cql:"category"`
	CreatedAt time.Time `json:"created_at" cql:"created_at"`
}

type ApiRequests struct {
	JsonClaz string     `json:"json_claz" cql:"json_claz"`
	Results  []Requests `json:"results" cql:"results"`
}

func (r *Requests) String() string {
	if d, err := yaml.Marshal(r); err != nil {
		return err.Error()
	} else {
		return string(d)
	}
}
