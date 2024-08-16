// Copyright 2016 CoreOS, Inc.
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/coreos/coreos-assembler/mantle/platform/api/aws"
	"github.com/coreos/stream-metadata-go/release"
	"github.com/spf13/cobra"
)

var (
	awsCredentialsFile string
	scpUser            string
	scpHost            string
	scpKeyFile         string
	scpTargetPath      string
	specProfile        string
	specRegion         string
	specStream         string
	specVersion        string

	specBucketPrefix string

	cmdMakeAmisPublic = &cobra.Command{
		Use:   "make-amis-public [options]",
		Short: "Make the AMIs of a CoreOS release public.",
		Run:   runMakeAmisPublic,
		Long:  "Make the AMIs of a CoreOS release public.",
	}

	cmdUpdateReleaseIndex = &cobra.Command{
		Use:   "update-release-index [options]",
		Short: "Update a stream's release index for a NestOS release.",
		Run:   runUpdateReleaseIndex,
		Long:  "Update a stream's release index for a NestOS release.",
	}
)

func init() {
	cmdMakeAmisPublic.Flags().StringVar(&awsCredentialsFile, "aws-credentials", "", "AWS credentials file")
	cmdMakeAmisPublic.Flags().StringVar(&specBucketPrefix, "bucket-prefix", "", "S3 bucket and prefix")
	cmdMakeAmisPublic.Flags().StringVar(&specProfile, "profile", "default", "AWS profile")
	cmdMakeAmisPublic.Flags().StringVar(&specRegion, "region", "us-east-1", "S3 bucket region")
	cmdMakeAmisPublic.Flags().StringVarP(&specStream, "stream", "", "", "target stream")
	cmdMakeAmisPublic.Flags().StringVarP(&specVersion, "version", "", "", "release version")
	root.AddCommand(cmdMakeAmisPublic)

	cmdUpdateReleaseIndex.Flags().StringVar(&awsCredentialsFile, "aws-credentials", "", "AWS credentials file")
	cmdUpdateReleaseIndex.Flags().StringVar(&specBucketPrefix, "bucket-prefix", "", "S3 bucket and prefix")
	cmdUpdateReleaseIndex.Flags().StringVar(&specProfile, "profile", "default", "AWS profile")
	cmdUpdateReleaseIndex.Flags().StringVar(&specRegion, "region", "us-east-1", "S3 bucket region")
	cmdUpdateReleaseIndex.Flags().StringVarP(&specStream, "stream", "", "", "target stream")
	cmdUpdateReleaseIndex.Flags().StringVarP(&specVersion, "version", "", "", "release version")
	cmdUpdateReleaseIndex.Flags().StringVar(&scpUser, "scp-user", "", "SCP user")
	cmdUpdateReleaseIndex.Flags().StringVar(&scpHost, "scp-host", "", "SCP host")
	cmdUpdateReleaseIndex.Flags().StringVar(&scpKeyFile, "scp-key-file", "", "SCP private key file")
	cmdUpdateReleaseIndex.Flags().StringVar(&scpTargetPath, "scp-target-path", "", "SCP target path")
	root.AddCommand(cmdUpdateReleaseIndex)

}

func validateArgs(args []string) {
	if len(args) > 0 {
		plog.Fatal("No args accepted")
	}
	if specVersion == "" {
		plog.Fatal("--version is required")
	}
	if specStream == "" {
		plog.Fatal("--stream is required")
	}
	if specBucketPrefix == "" {
		plog.Fatal("--bucket-prefix is required")
	}
	if specRegion == "" {
		plog.Fatal("--region is required")
	}
}

func runMakeAmisPublic(cmd *cobra.Command, args []string) {
	validateArgs(args)
	api := getAWSApi()
	rel := getReleaseMetadataFromS3(api)
	incomplete := makeReleaseAMIsPublic(rel)
	if incomplete {
		os.Exit(77)
	}
}

func runUpdateReleaseIndex(cmd *cobra.Command, args []string) {
	validateArgs(args)
	var rel release.Release
	var api *aws.API = nil
	if strings.HasPrefix(specBucketPrefix, "https://") {
		rel = getReleaseMetadataFromHTTPS(specBucketPrefix)
	} else {
		api = getAWSApi()
		rel = getReleaseMetadataFromS3(api)
	}
	modifyReleaseMetadataIndex(api, rel)
}

func getAWSApi() *aws.API {
	api, err := aws.New(&aws.Options{
		CredentialsFile: awsCredentialsFile,
		Profile:         specProfile,
		Region:          specRegion,
	})
	if err != nil {
		plog.Fatalf("creating aws client: %v", err)
	}

	return api
}

func getBucketAndStreamPrefix() (string, string) {
	split := strings.SplitN(specBucketPrefix, "/", 2)
	if len(split) != 2 {
		plog.Fatalf("can't split %q into bucket and prefix", specBucketPrefix)
	}
	return split[0], split[1]
}

func getReleaseMetadataFromS3(api *aws.API) release.Release {
	bucket, prefix := getBucketAndStreamPrefix()
	releasePath := filepath.Join(prefix, "builds", specVersion, "release.json")
	releaseFile, err := api.DownloadFile(bucket, releasePath)
	if err != nil {
		plog.Fatalf("downloading release metadata at %s: %v", releasePath, err)
	}
	defer releaseFile.Close()

	releaseData, err := io.ReadAll(releaseFile)
	if err != nil {
		plog.Fatalf("reading release metadata: %v", err)
	}

	var rel release.Release
	err = json.Unmarshal(releaseData, &rel)
	if err != nil {
		plog.Fatalf("unmarshaling release metadata: %v", err)
	}

	return rel
}

func getReleaseMetadataFromHTTPS(baseURL string) release.Release {
	releaseURL := fmt.Sprintf("%s/%s/builds/%s/release.json", baseURL, specStream, specVersion)
	resp, err := http.Get(releaseURL)
	if err != nil {
		plog.Fatalf("downloading release metadata from %s: %v", releaseURL, err)
	}
	defer resp.Body.Close()

	releaseData, err := io.ReadAll(resp.Body)
	if err != nil {
		plog.Fatalf("reading release metadata: %v", err)
	}

	var rel release.Release
	err = json.Unmarshal(releaseData, &rel)
	if err != nil {
		plog.Fatalf("unmarshaling release metadata: %v", err)
	}

	return rel
}

func makeReleaseAMIsPublic(rel release.Release) bool {
	at_least_one_tried := false
	at_least_one_passed := false
	at_least_one_failed := false
	for _, archs := range rel.Architectures {
		awsmedia := archs.Media.Aws
		if awsmedia == nil {
			continue
		}
		for region, ami := range awsmedia.Images {
			at_least_one_tried = true

			aws_api, err := aws.New(&aws.Options{
				CredentialsFile: awsCredentialsFile,
				Profile:         specProfile,
				Region:          region,
			})
			if err != nil {
				plog.Warningf("creating AWS API for region %s modifying launch permissions: %v", region, err)
				at_least_one_failed = true
				continue
			}

			plog.Noticef("making AMI %s in region %s public", ami.Image, region)
			err = aws_api.PublishImage(ami.Image)
			if err != nil {
				plog.Warningf("couldn't publish image in %v: %v", region, err)
				at_least_one_failed = true
				continue
			}

			at_least_one_passed = true
		}
	}

	if !at_least_one_tried {
		// if none were found, then we no-op
		return false
	} else if !at_least_one_passed {
		// if none passed, then it's likely a more fundamental issue like wrong
		// permissions or API usage, etc... let's just hard fail in that case
		plog.Fatal("failed to make AMIs public in all regions")
	}

	// all passed or some failed
	return at_least_one_failed
}

func modifyReleaseMetadataIndex(api *aws.API, rel release.Release) {
	var data []byte
	var err error
	bucket, prefix := getBucketAndStreamPrefix()

	if api == nil {
		// 从 HTTPS 获取 release index
		path := filepath.Join(prefix, specStream, "releases.json")
		url := fmt.Sprintf("%s/%s", bucket, path)
		resp, err := http.Get(url)
		if err != nil {
			plog.Fatalf("downloading release metadata index from %s: %v", url, err)
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			plog.Fatalf("reading release metadata index: %v", err)
		}
	} else {
		// 从 S3 获取 release index
		path := filepath.Join(prefix, "releases.json")
		data, err = func() ([]byte, error) {
			f, err := api.DownloadFile(bucket, path)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == "NoSuchKey" {
						return []byte("{}"), nil
					}
				}
				return []byte{}, fmt.Errorf("downloading release metadata index: %v", err)
			}
			defer f.Close()
			d, err := io.ReadAll(f)
			if err != nil {
				return []byte{}, fmt.Errorf("reading release metadata index: %v", err)
			}
			return d, nil
		}()
		if err != nil {
			plog.Fatal(err)
		}
	}

	var releaseIdx release.Index
	err = json.Unmarshal(data, &releaseIdx)
	if err != nil {
		plog.Fatalf("unmarshaling release metadata json: %v", err)
	}

	// 修改 release index
	metadataURL := fmt.Sprintf("%s/%s/builds/%s/release.json", specBucketPrefix, specStream, specVersion)

	var commits []release.IndexReleaseCommit
	for arch, vals := range rel.Architectures {
		commits = append(commits, release.IndexReleaseCommit{
			Architecture: arch,
			Checksum:     vals.Commit,
		})
	}

	newIdxRelease := release.IndexRelease{
		Commits:     commits,
		Version:     specVersion,
		MetadataURL: metadataURL,
	}

	for i, rel := range releaseIdx.Releases {
		if compareStaticReleaseInfo(rel, newIdxRelease) {
			if i != (len(releaseIdx.Releases) - 1) {
				plog.Fatalf("build is already present and is not the latest release")
			}

			comp := compareCommits(rel.Commits, newIdxRelease.Commits)
			if comp == 0 {
				// the build is already the latest release, exit
				plog.Notice("build is already present and is the latest release")
				return
			} else if comp == -1 {
				// the build is present and contains a subset of the new release data,
				// pop the old entry and add the new version
				releaseIdx.Releases = releaseIdx.Releases[:len(releaseIdx.Releases)-1]
				break
			} else {
				// the commit hash of the new build is not a superset of the current release
				plog.Fatalf("build is present but commit hashes are not a superset of latest release")
			}
		}
	}

	releaseIdx.Releases = append(releaseIdx.Releases, newIdxRelease)
	releaseIdx.Metadata.LastModified = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	releaseIdx.Note = "For use only by NestOS internal tooling. All other applications should obtain release info from stream metadata endpoints."
	releaseIdx.Stream = specStream

	out, err := json.Marshal(releaseIdx)
	if err != nil {
		plog.Fatalf("marshalling release metadata json: %v", err)
	}

	if api == nil {
		// 将 release index 写入本地文件，然后通过 SCP 上传
		tempDir, err := os.MkdirTemp("", "release-index")
		if err != nil {
			plog.Fatalf("creating temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir) // 确保在函数结束时删除临时目录

		localPath := filepath.Join(tempDir, "releases.json")
		err = os.WriteFile(localPath, out, 0644)
		if err != nil {
			plog.Fatalf("writing release metadata json to %s: %v", localPath, err)
		}

		// 使用 SCP 上传文件
		uploadReleaseIndexViaSCP(localPath)
	} else {
		// 上传 release index 到 S3
		var releases_max_age = 60 * 5
		err = api.UploadObjectExt(bytes.NewReader(out), bucket, path, true, "public-read", aws.ContentTypeJSON, releases_max_age)
		if err != nil {
			plog.Fatalf("uploading release metadata json: %v", err)
		}
	}
}

func uploadReleaseIndexViaSCP(localPath string) {
	remotePath := filepath.Join(scpTargetPath, specStream)

	// 确保远程目录存在
	ensureRemoteDirExists(remotePath)

	cmd := exec.Command("scp", "-i", scpKeyFile, localPath, fmt.Sprintf("%s@%s:%s", scpUser, scpHost, remotePath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		plog.Fatalf("uploading release index via SCP: %v", err)
	}
}

func ensureRemoteDirExists(remotePath string) {
	dir := filepath.Dir(remotePath)
	cmd := exec.Command("ssh", "-i", scpKeyFile, fmt.Sprintf("%s@%s", scpUser, scpHost), fmt.Sprintf("mkdir -p %s", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		plog.Fatalf("creating remote directory via SSH: %v", err)
	}
}

func compareStaticReleaseInfo(a, b release.IndexRelease) bool {
	if a.Version != b.Version || a.MetadataURL != b.MetadataURL {
		return false
	}
	return true
}

// returns -1 if a is a subset of b, 0 if equal, 1 if a is not a subset of b
func compareCommits(a, b []release.IndexReleaseCommit) int {
	if len(a) > len(b) {
		return 1
	}
	sameLength := len(a) == len(b)
	for _, aHash := range a {
		found := false
		for _, bHash := range b {
			if aHash.Architecture == bHash.Architecture && aHash.Checksum == bHash.Checksum {
				found = true
				break
			}
		}
		if !found {
			return 1
		}
	}
	if sameLength {
		return 0
	}
	return -1
}
