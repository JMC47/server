// Copyright (C) 2018-2023, John Chadwick <john@jchw.io>
//
// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
//
// SPDX-FileCopyrightText: Copyright (c) 2018-2023 John Chadwick
// SPDX-License-Identifier: ISC

package web

import (
	"io/fs"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pangbox/pangfiles/crypto/pyxtea"
	"github.com/pangbox/server/database/accounts"
	log "github.com/sirupsen/logrus"
)

const maxFormSize = 4096

type UpdateListOptions struct {
	Key pyxtea.Key
	Dir string
}

type Options struct {
	ServePangYaData bool
	UpdateList      *UpdateListOptions
	AccountsService *accounts.Service
}

type Handler struct {
	router          httprouter.Router
	updateHandler   *updateHandler
	accountsService *accounts.Service
}

func New(opt Options) *Handler {
	listener := &Handler{
		router:          *httprouter.New(),
		accountsService: opt.AccountsService,
	}

	if opt.UpdateList != nil {
		listener.updateHandler = newUpdateListHandler(opt.UpdateList.Key, opt.UpdateList.Dir)
	}

	assets, err := fs.Sub(assetFS, "assets")
	if err != nil {
		log.Fatalf("Error getting assets directory: %v", err)
	}
	listener.router.ServeFiles("/static/*filepath", http.FS(assets))
	listener.router.GET("/register", listener.handleRegisterGet)
	listener.router.POST("/register", listener.handleRegisterPost)
	listener.router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		http.Redirect(w, r, "/register", http.StatusFound)
	})

	// Update list paths
	if listener.updateHandler != nil {
		listener.router.GET("/pangya/season4/patch/updatelist", listener.handleUpdateList)
		listener.router.GET("/pangya/season4/patch/qa/updatelist", listener.handleUpdateList)
		listener.router.GET("/new/Service/S4_Patch/updatelist", listener.handleUpdateList)
	}

	// PangYa game data
	if opt.ServePangYaData {
		listener.router.GET("/Translation/Read.aspx", listener.serveTranslations)
		listener.router.GET("/new/Service/S4_Patch/extracontents/extracontents.xml", listener.serveExtraContents)
		listener.router.GET("/pangya/season4/patch/extracontents/extracontents.xml", listener.serveExtraContents)
		listener.router.GET("/S4_Patch/extracontents/default/pangya_default.xml", listener.servePangyaDefault)
	}

	return listener
}

func (l *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("WEB: %s %s", r.Method, r.URL.String())
	l.router.ServeHTTP(w, r)
}
