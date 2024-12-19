// Copyright 2019 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log_test

import (
	"bytes"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"

	"github.com/scionproto/scion/pkg/log"
	"github.com/scionproto/scion/pkg/log/logtest"
	"github.com/scionproto/scion/private/config"
)

func TestLoggingSample(t *testing.T) {
	id := "logID"
	var sample bytes.Buffer
	var cfg log.Config
	cfg.Sample(&sample, nil, map[string]string{config.ID: id})
	logtest.InitTestLogging(&cfg)
	err := toml.NewDecoder(bytes.NewReader(sample.Bytes())).DisallowUnknownFields().Decode(&cfg)
	assert.NoError(t, err)
	logtest.CheckTestLogging(t, &cfg, id)
}
