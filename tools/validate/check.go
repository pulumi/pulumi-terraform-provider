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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/spf13/cobra"
)

const rootPath = "providers"

func checkCmd() *cobra.Command {
	var (
		source string
	)
	cmd := &cobra.Command{
		Use: "check",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			version, err := getDynamicProviderVersion(ctx)
			if err != nil {
				exit(fmt.Errorf("unable to get terraform-provider version: %w", err))
			}
			report := providerReport{
				version:   version,
				args:      []string{source},
				languages: map[string]languageReport{},
			}

			report.pulumiVersion, err = getPulumiVersion(ctx)
			if err != nil {
				exit(fmt.Errorf("unable to get pulumi version: %w", err))
			}

			report.schema, report.schemaStderr, err = getDynamicProviderSchema(ctx, source, "")
			if err != nil {
				fmt.Printf("unable to get schema for %q: %s\n", report.args, err.Error())
				exit(report.write(rootPath))
			}

			tmpDir, err := os.MkdirTemp("", "check")
			if err != nil {
				exit(err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					exit(err)
				}
			}()

			schemaPath := filepath.Join(tmpDir, "schema.json")
			if err := os.WriteFile(schemaPath, marshal(report.schema), 0600); err != nil {
				exit(err)
			}

			for lang, check := range languages {
				report.languages[lang] = check(ctx, schemaPath, tmpDir)
			}

			fmt.Printf("All data collected\n")
			fmt.Printf("Schema generated: %t\n", report.schema != nil)
			for lang, r := range report.languages {
				fmt.Printf("%s: generated=%t, built=%t\n", lang, r.sdkPath != "", r.succeeded)
			}
			exit(report.write(rootPath))
		},
	}

	cmd.PersistentFlags().StringVarP(&source, "source", "", "",
		"The source string for the TF provider")

	return cmd
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Printf("error: %s\n", err.Error())
	os.Exit(1)
}

func getDynamicProviderVersion(ctx context.Context) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "pulumi", "package", "get-schema", "terraform-provider")
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	var spec schema.PackageSpec
	err = json.Unmarshal(stdout.Bytes(), &spec)
	return spec.Version, err

}

func getPulumiVersion(ctx context.Context) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "pulumi", "version")
	cmd.Stdout = &stdout
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), err

}

func getDynamicProviderSchema(ctx context.Context, source, version string) (*schema.PackageSpec, []byte, error) {
	args := []string{
		"package",
		"get-schema",
		"terraform-provider",
		source,
	}
	if version != "" {
		args = append(args, version)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "pulumi", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, stderr.Bytes(), err
	}
	var spec schema.PackageSpec
	err = json.Unmarshal(stdout.Bytes(), &spec)
	if err != nil {
		return nil, stderr.Bytes(), err
	}
	return &spec, stderr.Bytes(), err
}

type providerReport struct {
	version       string
	args          []string
	pulumiVersion string

	schema       *schema.PackageSpec
	schemaStderr []byte

	languages map[string]languageReport
}

type languageReport struct {
	buildCommand string
	genSdkStderr []byte

	succeeded   bool
	buildStderr []byte
	sdkPath     string
}

func (r providerReport) path() string {
	// No schema, so the get-schema call didn't succeed.
	if r.schema == nil {
		return filepath.Join(
			r.version,
			"failures",
			strings.Join(r.args, "-"),
		)
	}
	return filepath.Join(
		r.version,
		r.args[0],
		r.schema.Version,
	)
}

type metadata struct {
	Args          []string `json:"args"`
	PulumiVersion string   `json:"pulumiVersion"`
}

func marshal(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err := enc.Encode(v)
	contract.AssertNoErrorf(err, "Failed to marshal %T", v)
	return b.Bytes()
}

func (r providerReport) write(root string) error {
	dir := filepath.Join(root, r.path())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if r.schema != nil {
		if err := os.WriteFile(filepath.Join(dir, "schema.json"), marshal(r.schema), 0600); err != nil {
			return err
		}
	}
	if len(r.schemaStderr) > 0 {
		if err := os.WriteFile(filepath.Join(dir, "schema-stderr.txt"), r.schemaStderr, 0600); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), marshal(metadata{
		Args:          r.args,
		PulumiVersion: r.pulumiVersion,
	}), 0600); err != nil {
		return err
	}

	for lang, report := range r.languages {
		dir := filepath.Join(dir, lang)
		if len(report.buildStderr) > 0 {
			if err := os.WriteFile(
				filepath.Join(dir, "build-stderr.txt"),
				report.buildStderr, 0600); err != nil {
				return err
			}
		}

		if len(report.genSdkStderr) > 0 {
			if err := os.WriteFile(
				filepath.Join(dir, "gen-sdk-stderr.txt"),
				report.genSdkStderr, 0600); err != nil {
				return err
			}
		}

		if report.sdkPath != "" {
			err := os.CopyFS(filepath.Join(dir, "sdk"), os.DirFS(report.sdkPath))
			if err != nil {
				return err
			}
		}

		if err := os.WriteFile(filepath.Join(dir, "metadata.json"), marshal(struct {
			BuildCommand string `json:"buildCommand"`
			Succeded     bool   `json:"succeded"`
		}{
			BuildCommand: report.buildCommand,
			Succeded:     report.succeeded,
		}), 0600); err != nil {
			return err
		}

	}

	return nil

}

func (r *providerReport) read(root string) error {
	panic("UNIMPLIMENTED")
}
