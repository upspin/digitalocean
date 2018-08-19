package spaces_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"upspin.io/cloud/storage"
	"upspin.io/log"
)

var (
	client      storage.Storage
	testDataStr = fmt.Sprintf("This is test at %v", time.Now())
	testData    = []byte(testDataStr)
	fileName    = fmt.Sprintf("test-file-%d", time.Now().Second())

	testSpacesName   = flag.String("test_space", "", "bucket name to use for testing")
	testSpacesRegion = flag.String("test_region", "", "region to use for the test bucket")
	testSpacesRoot   = flag.String("test_root", "", "region to use for the test bucket")
	useSpaces        = flag.Bool("use_spaces", false, "enable to run aws tests; requires aws credentials")

	testFilePath = ""
)

// NOTE: test_root should not have trailing slash. for example, /test-upspin is wrong
// use test-upspin

// This is more of a regression test as it uses the running cloud
// storage in prod. However, since S3 is always available, we accept
// relying on it.
func TestPutAndDownload(t *testing.T) {
	err := client.Put(testFilePath, testData)
	if err != nil {
		t.Fatalf("Can't put: %v", err)
	}
	data, err := client.Download(testFilePath)
	if err != nil {
		t.Fatalf("Can't Download: %v", err)
	}
	if string(data) != testDataStr {
		t.Errorf("Expected %q got %q", testDataStr, string(data))
	}
}

func TestDelete(t *testing.T) {
	err := client.Put(testFilePath, testData)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Delete(testFilePath)
	if err != nil {
		t.Fatalf("Expected no errors, got %v", err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !*useSpaces {
		log.Printf(`
cloud/storage/spaces: skipping test as it requires Digitalocean's spaces access. To enable this
test, ensure you are properly authorized to upload to an Spaces bucket named by flag
-test_space and then set this test's flag -use_spaces.
`)
		os.Exit(0)
	}

	// Create client that writes to test bucket.
	var err error
	client, err = storage.Dial("Spaces",
		storage.WithKeyValue("spacesRegion", *testSpacesRegion),
		storage.WithKeyValue("spacesName", *testSpacesName))
	if err != nil {
		log.Fatalf("cloud/storage/spaces: couldn't set up client: %v", err)
	}

	testFilePath = filepath.Join(*testSpacesRoot, fileName)

	code := m.Run()

	os.Exit(code)
}
