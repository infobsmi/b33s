// Copyright (c) 2015-2022 B33S, Inc.
//
// This file is part of B33S Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"net/url"
	"strings"
	"time"

	"github.com/b33s/madmin-go/v2"
	b33s "github.com/infobsmi/b33s-go/v7"
	"github.com/infobsmi/b33s-go/v7/pkg/credentials"
)

type warmBackendB33S struct {
	warmBackendS3
}

var _ WarmBackend = (*warmBackendB33S)(nil)

func newWarmBackendB33S(conf madmin.TierB33S) (*warmBackendB33S, error) {
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, "")

	getRemoteTierTargetInstanceTransportOnce.Do(func() {
		getRemoteTierTargetInstanceTransport = newHTTPTransport(10 * time.Minute)
	})
	opts := &b33s.Options{
		Creds:     creds,
		Secure:    u.Scheme == "https",
		Transport: getRemoteTierTargetInstanceTransport,
	}
	client, err := b33s.New(u.Host, opts)
	if err != nil {
		return nil, err
	}
	core, err := b33s.NewCore(u.Host, opts)
	if err != nil {
		return nil, err
	}
	return &warmBackendB33S{
		warmBackendS3{
			client: client,
			core:   core,
			Bucket: conf.Bucket,
			Prefix: strings.TrimSuffix(conf.Prefix, slashSeparator),
		},
	}, nil
}
