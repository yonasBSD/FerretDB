// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Find implements [ServerInterface].
func (s *Server) Find(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, fmt.Sprintf("Request:\n%s", must.NotFail(httputil.DumpRequest(r, true))))
	}

	var req api.FindManyRequestBody
	if err := decodeJsonRequest(r, &req); err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	msg, err := prepareOpMsg(
		"find", req.Collection,
		"$db", req.Database,
		"filter", req.Filter,
		"limit", req.Limit,
		"projection", req.Projection,
		"skip", req.Skip,
		"sort", req.Sort,
	)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	// We should use s.handler.Handle().
	// TODO https://github.com/FerretDB/FerretDB/issues/5046
	resMsg, err := s.handler.Commands()["find"].Handler(ctx, msg)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	resRaw := must.NotFail(resMsg.OpMsg.RawDocument())
	cursor := must.NotFail(resRaw.Decode()).Get("cursor").(wirebson.AnyDocument)

	res := must.NotFail(wirebson.NewDocument(
		"documents", must.NotFail(cursor.Decode()).Get("firstBatch"),
	))

	s.writeJsonResponse(ctx, w, res)
}
