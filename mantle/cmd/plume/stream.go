package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coreos/stream-metadata-go/release"
	"github.com/coreos/stream-metadata-go/stream"
	"github.com/spf13/cobra"
)

// Substituted by Makefile
var Version = "v1.0"
var generator = "nestos-stream-generator " + Version
var errReleaseIndexMissing = errors.New("Please specify release index url or release override")

var (
	releasesBaseURL     string
	overrideReleasePath string
	overrideFilename    string
	outputFile          string
	prettyPrint         bool
	showVersion         bool
	specstream          string
	user                string
	host                string
	path                string
	key                 string

	cmdGenerateStreamMetadata = &cobra.Command{
		Use:   "stream-generate",
		Short: "Generate NestOS stream metadata",
		Run:   runGenerateStreamMetadata,
		Long:  "Generate NestOS stream metadata",
	}
)

func getReleaseURL(releaseIndexBaseURL string, specstream string) (string, error) {
	// Ensure releaseIndexBaseURL ends with a slash
	if !strings.HasSuffix(releaseIndexBaseURL, "/") {
		releaseIndexBaseURL += "/"
	}

	// Ensure specstream does not start with a slash
	if strings.HasPrefix(specstream, "/") {
		specstream = strings.TrimPrefix(specstream, "/")
	}

	// Construct the full URL
	releaseIndexURL := fmt.Sprintf("%s%s/releases.json", releaseIndexBaseURL, specstream)

	var relIndex release.Index

	resp, err := http.Get(releaseIndexURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&relIndex); err != nil {
		return "", err
	}
	if len(relIndex.Releases) < 1 {
		return "", fmt.Errorf("No release available to process")
	}

	return relIndex.Releases[len(relIndex.Releases)-1].MetadataURL, nil
}

func overrideData(original, override interface{}) interface{} {
	switch override1 := override.(type) {
	case map[string]interface{}:
		original1, ok := original.(map[string]interface{})
		if !ok {
			return override1
		}
		for key, value1 := range original1 {
			if value2, ok := override1[key]; ok {
				override1[key] = overrideData(value1, value2)
			} else {
				override1[key] = value1
			}
		}
	case nil:
		original1, ok := original.(map[string]interface{})
		if ok {
			return original1
		}
	}
	return override
}

func runGenerateStreamMetadata(cmd *cobra.Command, args []string) {
	err := GenerateStreamMetadata(releasesBaseURL, overrideReleasePath, overrideFilename, outputFile, prettyPrint, showVersion, specstream)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GenerateStreamMetadata(releasesBaseURL, overrideReleasePath, overrideFilename, outputFile string, prettyPrint, version bool, specstream string) error {
	if version {
		fmt.Println(generator)
		return nil
	}

	var releasePath string
	if releasesBaseURL == "" && overrideReleasePath == "" {
		return errReleaseIndexMissing
	} else if releasesBaseURL != "" && overrideReleasePath != "" {
		return fmt.Errorf("Can't specify both -releases and -release")
	} else if overrideReleasePath != "" {
		releasePath = overrideReleasePath
	} else {
		var err error
		releasePath, err = getReleaseURL(releasesBaseURL, specstream)
		if err != nil {
			return fmt.Errorf("Error with Release Index: %v", err)
		}
	}

	parsedURL, err := url.Parse(releasePath)
	if err != nil {
		return fmt.Errorf("Error while parsing release path: %v", err)
	}

	var decoder *json.Decoder
	if parsedURL.Scheme == "" {
		releaseMetadataFile, err := os.Open(releasePath)
		if err != nil {
			return fmt.Errorf("Error opening file: %v", err)
		}
		defer releaseMetadataFile.Close()
		decoder = json.NewDecoder(releaseMetadataFile)
	} else {
		resp, err := http.Get(releasePath)
		if err != nil {
			return fmt.Errorf("Error while fetching: %v", err)
		}
		defer resp.Body.Close()
		decoder = json.NewDecoder(resp.Body)
	}

	var rel release.Release
	if err = decoder.Decode(&rel); err != nil {
		return fmt.Errorf("Error while decoding json: %v", err)
	}

	streamMetadata := stream.Stream{
		Stream: rel.Stream,
		Metadata: stream.Metadata{
			LastModified: time.Now().UTC().Format(time.RFC3339),
			Generator:    generator,
		},
		Architectures: rel.ToStreamArchitectures(),
	}

	if overrideFilename != "" {
		overrideFile, err := os.Open(overrideFilename)
		if err != nil {
			return fmt.Errorf("Can't open file %s: %v", overrideFilename, err)
		}
		defer overrideFile.Close()

		streamMetadataJSON, err := json.Marshal(&streamMetadata)
		if err != nil {
			return fmt.Errorf("Error during Marshal: %v", err)
		}
		streamMetadataMap := make(map[string]interface{})
		if err = json.Unmarshal(streamMetadataJSON, &streamMetadataMap); err != nil {
			return fmt.Errorf("Error during Unmarshal: %v", err)
		}

		overrideMap := make(map[string]interface{})
		overrideDecoder := json.NewDecoder(overrideFile)
		if err = overrideDecoder.Decode(&overrideMap); err != nil {
			return fmt.Errorf("Error while decoding: %v", err)
		}

		streamMetadataInterface := overrideData(streamMetadataMap, overrideMap)
		streamMetadataMap = streamMetadataInterface.(map[string]interface{})

		streamMetadataJSON, err = json.Marshal(streamMetadataMap)
		if err != nil {
			return fmt.Errorf("Error during Marshal: %v", err)
		}
		if err = json.Unmarshal(streamMetadataJSON, &streamMetadata); err != nil {
			return fmt.Errorf("Error during Unmarshal: %v", err)
		}
	}

	tmpFile, err := ioutil.TempFile("", "stream_metadata_*.json")
	if err != nil {
		return fmt.Errorf("Can't create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 修改临时文件权限为 644
	if err := os.Chmod(tmpFile.Name(), 0644); err != nil {
		return fmt.Errorf("Error setting file permissions: %v", err)
	}

	encoder := json.NewEncoder(tmpFile)
	if prettyPrint {
		encoder.SetIndent("", "    ")
	}
	if err := encoder.Encode(&streamMetadata); err != nil {
		return fmt.Errorf("Error while encoding: %v", err)
	}
	tmpFile.Close()

	// SCP transfer
	if user != "" && host != "" && path != "" && key != "" {
		scpCmd := fmt.Sprintf("scp -i %s %s %s@%s:%s/%s.json", key, tmpFile.Name(), user, host, path, specstream)
		cmd := exec.Command("bash", "-c", scpCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Error while transferring file: %v", err)
		}
	}

	return nil
}

func init() {
	cmdGenerateStreamMetadata.Flags().StringVar(&releasesBaseURL, "releases", "", "Release index base URL for the required stream")
	cmdGenerateStreamMetadata.Flags().StringVar(&overrideReleasePath, "release", "", "Override release metadata location")
	cmdGenerateStreamMetadata.Flags().StringVar(&overrideFilename, "override", "", "Override file location for the required stream")
	cmdGenerateStreamMetadata.Flags().StringVar(&outputFile, "output-file", "", "Save output into a file")
	cmdGenerateStreamMetadata.Flags().BoolVar(&prettyPrint, "pretty-print", false, "Pretty-print output")
	cmdGenerateStreamMetadata.Flags().BoolVar(&showVersion, "version", false, "Show version")
	cmdGenerateStreamMetadata.Flags().StringVar(&specstream, "stream", "", "Stream name (e.g. stable)")
	cmdGenerateStreamMetadata.Flags().StringVar(&user, "user", "", "Username for scp transfer")
	cmdGenerateStreamMetadata.Flags().StringVar(&host, "host", "", "Host for scp transfer")
	cmdGenerateStreamMetadata.Flags().StringVar(&path, "path", "", "Path for scp transfer")
	cmdGenerateStreamMetadata.Flags().StringVar(&key, "key", "", "Key file for scp transfer")

	root.AddCommand(cmdGenerateStreamMetadata)
}
