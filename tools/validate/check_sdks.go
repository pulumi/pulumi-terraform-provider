// Copyright 2016-2024, Pulumi Corporation.
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

package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
)

var languages = map[string]func(ctx context.Context, schemaPath, tmpDir string) languageReport{
	"dotnet": checkDotnet,
}

func checkDotnet(ctx context.Context, schemaPath, tmpDir string) languageReport {

	sdkPath, genSdkStderr, err := genPulumiSDK(ctx, schemaPath, tmpDir, "dotnet")
	if err != nil {
		return languageReport{
			genSdkStderr: genSdkStderr,
			succeeded:    false,
		}
	}

	cmd := exec.CommandContext(ctx, "dotnet", "build")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Dir = sdkPath

	err = cmd.Run()

	return languageReport{
		buildCommand: "dotnet build",
		genSdkStderr: genSdkStderr,
		succeeded:    err == nil,
		buildStderr:  stderr.Bytes(),
		sdkPath:      sdkPath,
	}
}

func genPulumiSDK(ctx context.Context, schemaPath, tmpDir, language string) (string, []byte, error) {
	out := filepath.Join(tmpDir, "language")
	cmd := exec.CommandContext(ctx, "pulumi",
		"package", "gen-sdk",
		schemaPath,
		"--language="+language,
		"--out="+out,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	return filepath.Join(out, language), stderr.Bytes(), err
}
