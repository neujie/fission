/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"io/ioutil"
	"net/http"

	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"encoding/base64"
	"github.com/fission/fission"
)

func (api *API) FunctionApiList(w http.ResponseWriter, r *http.Request) {
	funcs, err := api.FunctionStore.List()
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	resp, err := json.Marshal(funcs)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, resp)
}

func (api *API) FunctionApiCreate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.respondWithError(w, err)
	}

	var f fission.Function
	err = json.Unmarshal(body, &f)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	dec, err := base64.StdEncoding.DecodeString(f.Code)
	if err != nil {
		api.respondWithError(w, err)
		return
	}
	f.Code = string(dec)

	uid, err := api.FunctionStore.Create(&f)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	m := &fission.Metadata{Name: f.Name, Uid: uid}
	resp, err := json.Marshal(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, resp)
}

func (api *API) FunctionApiGet(w http.ResponseWriter, r *http.Request) {
	var m fission.Metadata

	vars := mux.Vars(r)
	m.Name = vars["function"]
	m.Uid = r.FormValue("uid") // empty if uid is absent
	raw := r.FormValue("raw")  // just the code

	f, err := api.FunctionStore.Get(&m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	var resp []byte
	if raw != "" {
		resp = []byte(f.Code)
	} else {
		f.Code = base64.StdEncoding.EncodeToString([]byte(f.Code))
		resp, err = json.Marshal(f)
		if err != nil {
			api.respondWithError(w, err)
			return
		}
	}
	api.respondWithSuccess(w, resp)
}

func (api *API) FunctionApiUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	funcName := vars["function"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.respondWithError(w, err)
	}

	var f fission.Function
	err = json.Unmarshal(body, &f)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	if funcName != f.Metadata.Name {
		err = fission.MakeError(fission.ErrorInvalidArgument, "Function name doesn't match URL")
		api.respondWithError(w, err)
		return
	}

	dec, err := base64.StdEncoding.DecodeString(f.Code)
	if err != nil {
		api.respondWithError(w, err)
		return
	}
	f.Code = string(dec)

	uid, err := api.FunctionStore.Update(&f)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	m := &fission.Metadata{Name: f.Name, Uid: uid}
	resp, err := json.Marshal(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}
	api.respondWithSuccess(w, resp)
}

func (api *API) FunctionApiDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var m fission.Metadata
	m.Name = vars["function"]

	m.Uid = r.FormValue("uid") // empty if uid is absent
	if len(m.Uid) == 0 {
		log.WithFields(log.Fields{"function": m.Name}).Info("Deleting all versions")
	}

	err := api.FunctionStore.Delete(m)
	if err != nil {
		api.respondWithError(w, err)
		return
	}

	api.respondWithSuccess(w, []byte(""))
}
