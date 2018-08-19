package spaces

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	minio "github.com/minio/minio-go"

	"upspin.io/cloud/storage"
	"upspin.io/errors"
)

// Keys used for storing dial options.
const (
	regionName = "spacesRegion"
	spacesName = "spacesName"
	spacesRoot = "spacesRoot"
)

// this meta data is used to make all the files used in
// upspin public
var metaData = map[string]string{
	"x-amz-acl": "public-read",
}

// spacesImpl is an implementation of Storage that connects to an Amazon Simple
// Storage (S3) backend.
type spacesImpl struct {
	client     *minio.Client
	spacesName string
	endpoint   string
	root       string
}

// New initializes a Storage implementation that stores data to Spaces Simple
// Storage Service.
func New(opts *storage.Opts) (storage.Storage, error) {
	const op errors.Op = "cloud/storage/spaces.New"
	const ssl = true

	accessKey := os.Getenv("SPACES_KEY")
	if accessKey == "" {
		return nil, errors.E(op, errors.Invalid, errors.Errorf("SPACES_KEY env variable is required"))
	}

	secKey := os.Getenv("SPACES_SECRET")
	if secKey == "" {
		return nil, errors.E(op, errors.Invalid, errors.Errorf("SPACES_SECRET env variable is required"))
	}

	region, ok := opts.Opts[regionName]
	if !ok {
		return nil, errors.E(op, errors.Invalid, errors.Errorf("%q option is required", regionName))
	}

	name, ok := opts.Opts[spacesName]
	if !ok {
		return nil, errors.E(op, errors.Invalid, errors.Errorf("%q option is required", name))
	}

	root, ok := opts.Opts[spacesRoot]
	if ok {
		if strings.HasPrefix(root, "/") {
			return nil, errors.E(op, errors.Invalid, errors.Errorf("%q option is shouldn't start with slash", name))
		}
	}

	endpoint := fmt.Sprintf("%s.digitaloceanspaces.com", region)

	// Initiate a client using DigitalOcean Spaces.
	client, err := minio.NewV4(endpoint, accessKey, secKey, ssl)
	if err != nil {
		return nil, errors.E(op, errors.IO, errors.Errorf("unable to create minio session: %s", err))
	}

	return &spacesImpl{
		client:     client,
		spacesName: name,
		endpoint:   endpoint,
		root:       root,
	}, nil
}

func init() {
	storage.Register("Spaces", New)
}

// Guarantee we implement the Storage interface.
var _ storage.Storage = (*spacesImpl)(nil)

func (s *spacesImpl) refPath(ref string) string {
	return filepath.Join(s.root, ref)
}

// LinkBase implements Storage.
func (s *spacesImpl) LinkBase() (base string, err error) {
	if s.root != "" {
		return fmt.Sprintf("https://%s.%s/%s/", s.spacesName, s.endpoint, s.root), nil
	}
	return fmt.Sprintf("https://%s.%s/", s.spacesName, s.endpoint), nil
}

// Download implements Storage.
func (s *spacesImpl) Download(ref string) ([]byte, error) {
	const op errors.Op = "cloud/storage/spaces.Download"

	ref = s.refPath(ref)

	obj, err := s.client.GetObject(s.spacesName, ref, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, errors.IO, errors.Errorf(
			"unable to download ref %q from bucket %q: %s", ref, s.spacesName, err))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(obj)
	return buf.Bytes(), nil
}

// Put implements Storage.
func (s *spacesImpl) Put(ref string, contents []byte) error {
	const op errors.Op = "cloud/storage/spaces.Put"

	ref = s.refPath(ref)

	_, err := s.client.PutObject(s.spacesName, ref, bytes.NewReader(contents), int64(len(contents)), minio.PutObjectOptions{
		UserMetadata: metaData,
	})

	if err != nil {
		return errors.E(op, errors.IO, errors.Errorf(
			"unable to upload ref %q to bucket %q: %s", ref, s.spacesName, err))
	}

	return nil
}

// Delete implements Storage.
func (s *spacesImpl) Delete(ref string) error {
	const op errors.Op = "cloud/storage/spaces.Delete"

	ref = s.refPath(ref)

	err := s.client.RemoveObject(s.spacesName, ref)
	if err != nil {
		return errors.E(op, errors.IO, errors.Errorf(
			"unable to delete ref %q from bucket %q: %s", ref, s.spacesName, err))
	}

	return nil
}

// Close implements Storage.
func (s *spacesImpl) Close() {
	s.client = nil
	s.spacesName = ""
}
