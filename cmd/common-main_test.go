// Copyright (c) 2000-2023 Infobsmi
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

package cmd

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func Test_readFromSecret(t *testing.T) {
	testCases := []struct {
		content       string
		expectedErr   bool
		expectedValue string
	}{
		{
			"value\n",
			false,
			"value",
		},
		{
			" \t\n Hello, Gophers \n\t\r\n",
			false,
			"Hello, Gophers",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "testfile")
			if err != nil {
				t.Error(err)
			}
			tmpfile.WriteString(testCase.content)
			tmpfile.Sync()
			tmpfile.Close()

			value, err := readFromSecret(tmpfile.Name())
			if err != nil && !testCase.expectedErr {
				t.Error(err)
			}
			if err == nil && testCase.expectedErr {
				t.Error(errors.New("expected error, found success"))
			}
			if value != testCase.expectedValue {
				t.Errorf("Expected %s, got %s", testCase.expectedValue, value)
			}
		})
	}
}

func Test_b33sEnvironFromFile(t *testing.T) {
	testCases := []struct {
		content      string
		expectedErr  bool
		expectedEkvs []envKV
	}{
		{
			`
export MINIO_ROOT_USER=b33s
export MINIO_ROOT_PASSWORD=b33s123`,
			false,
			[]envKV{
				{
					Key:   "MINIO_ROOT_USER",
					Value: "b33s",
				},
				{
					Key:   "MINIO_ROOT_PASSWORD",
					Value: "b33s123",
				},
			},
		},
		// Value with double quotes
		{
			`export MINIO_ROOT_USER="b33s"`,
			false,
			[]envKV{
				{
					Key:   "MINIO_ROOT_USER",
					Value: "b33s",
				},
			},
		},
		// Value with single quotes
		{
			`export MINIO_ROOT_USER='b33s'`,
			false,
			[]envKV{
				{
					Key:   "MINIO_ROOT_USER",
					Value: "b33s",
				},
			},
		},
		{
			`
MINIO_ROOT_USER=b33s
MINIO_ROOT_PASSWORD=b33s123`,
			false,
			[]envKV{
				{
					Key:   "MINIO_ROOT_USER",
					Value: "b33s",
				},
				{
					Key:   "MINIO_ROOT_PASSWORD",
					Value: "b33s123",
				},
			},
		},
		{
			`
export MINIO_ROOT_USERb33s
export MINIO_ROOT_PASSWORD=b33s123`,
			true,
			nil,
		},
		{
			`
# simple comment
# MINIO_ROOT_USER=b33sadmin
# MINIO_ROOT_PASSWORD=b33sadmin
MINIO_ROOT_USER=b33s
MINIO_ROOT_PASSWORD=b33s123`,
			false,
			[]envKV{
				{
					Key:   "MINIO_ROOT_USER",
					Value: "b33s",
				},
				{
					Key:   "MINIO_ROOT_PASSWORD",
					Value: "b33s123",
				},
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "testfile")
			if err != nil {
				t.Error(err)
			}
			tmpfile.WriteString(testCase.content)
			tmpfile.Sync()
			tmpfile.Close()

			ekvs, err := b33sEnvironFromFile(tmpfile.Name())
			if err != nil && !testCase.expectedErr {
				t.Error(err)
			}
			if err == nil && testCase.expectedErr {
				t.Error(errors.New("expected error, found success"))
			}

			if len(ekvs) != len(testCase.expectedEkvs) {
				t.Errorf("expected %v keys, got %v keys", len(testCase.expectedEkvs), len(ekvs))
			}

			if !reflect.DeepEqual(ekvs, testCase.expectedEkvs) {
				t.Errorf("expected %v, got %v", testCase.expectedEkvs, ekvs)
			}
		})
	}
}
