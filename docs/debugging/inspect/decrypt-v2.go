// Copyright (c) 2015-2022 B33S, Inc.
//
// This file is part of B33SObject Storage stack
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

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/minio/madmin-go/v2/estream"
)

func extractInspectV2(pk []byte, r io.Reader, w io.Writer) error {
	privKey, err := bytesToPrivateKey(pk)
	if err != nil {
		return fmt.Errorf("decoding key returned: %w", err)
	}

	sr, err := estream.NewReader(r)
	if err != nil {
		return err
	}

	sr.SetPrivateKey(privKey)
	sr.ReturnNonDecryptable(true)

	// Debug corrupted streams.
	if false {
		sr.SkipEncrypted(true)
		return sr.DebugStream(os.Stdout)
	}

	for {
		stream, err := sr.NextStream()
		if err != nil {
			if err == io.EOF {
				return errors.New("no data found on stream")
			}
			if errors.Is(err, estream.ErrNoKey) {
				if stream.Name == "inspect.zip" {
					return errors.New("incorrect private key")
				}
				if err := stream.Skip(); err != nil {
					return fmt.Errorf("stream skip: %w", err)
				}
				continue
			}
			return fmt.Errorf("next stream: %w", err)
		}
		if stream.Name == "inspect.zip" {
			_, err := io.Copy(w, stream)
			if err != nil {
				return fmt.Errorf("reading inspect stream: %w", err)
			}
			return nil
		}
		if err := stream.Skip(); err != nil {
			return fmt.Errorf("stream skip: %w", err)
		}
	}
}
