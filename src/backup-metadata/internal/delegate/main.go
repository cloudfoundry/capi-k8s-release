package delegate

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/cfmetadata"
)

//go:generate counterfeiter . Reader
type Reader = io.Reader

const (
	reportMetadata  = 1
	compareMetadata = 2
)

func Main(args []string, input Reader, env map[string]string) error {
	cfAPI := env["CF_API"]
	cfUsername := env["CF_USER"]
	cfPassword := env["CF_PASSWORD"]

	if cfAPI == "" {
		return fmt.Errorf("CF_API Environment is not set")
	}

	if cfUsername == "" {
		return fmt.Errorf("CF_USER Environment is not set")
	}

	if cfPassword == "" {
		return fmt.Errorf("CF_PASSWORD Environment is not set")
	}

	switch len(args) {
	case reportMetadata:
		err := printMetadata(cfAPI, cfUsername, cfPassword)
		if err != nil {
			return err
		}
	case compareMetadata:
		if args[1] != "compare" {
			return fmt.Errorf("unknown option: %s, Did you mean compare ?", args[1])
		}

		if err := printComparison(cfAPI, cfUsername, cfPassword, input); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid arguments, usage: %s or "+
			"cat cf-metadata-another-system.json | %s compare", args[0], args[0])
	}

	return nil
}

func printMetadata(cfAPI, cfUsername, cfPassword string) error {
	backupMetadata, err := getCurrentCFMetadata(cfAPI, cfUsername, cfPassword)
	if err != nil {
		return err
	}

	fmt.Println("CF Metadata: " + toStr(*backupMetadata))

	return nil
}

func printComparison(cfAPI, cfUsername, cfPassword string, input Reader) error {
	var (
		diff        string
		err         error
		pipeContent []byte
	)

	pipeContent, err = ioutil.ReadAll(input)
	if err != nil {
		return err
	}

	backupMetadata, err := getCurrentCFMetadata(cfAPI, cfUsername, cfPassword)
	if err != nil {
		return err
	}

	if diff, err = cfmetadata.Compare(pipeContent, *backupMetadata); err != nil {
		return fmt.Errorf("could not compare current CF metadata with input JSON: %s", err)
	}

	if diff == "" {
		fmt.Println("No differences found between input and current state")
	} else {
		fmt.Println(diff)
	}

	return nil
}

func toStr(metadata cfmetadata.Metadata) string {
	ret, _ := json.Marshal(metadata)

	return string(ret)
}

func getCurrentCFMetadata(cfAPI string, cfUsername string, cfPassword string) (*cfmetadata.Metadata, error) {
	client, err := cfmetadata.NewClient(
		cfAPI,
		cfUsername,
		cfPassword)
	if err != nil {
		return &cfmetadata.Metadata{}, err
	}

	mg, err := cfmetadata.NewMetadataGetter(client)
	if err != nil {
		return &cfmetadata.Metadata{}, err
	}

	m, err := mg.Execute()
	if err != nil {
		return &cfmetadata.Metadata{}, err
	}

	return m, err
}
