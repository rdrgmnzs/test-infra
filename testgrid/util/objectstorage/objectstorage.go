/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/go-cloud/blob"
)

// Path parses gs://bucket/obj urls
type Path struct {
	url url.URL
}

// String returns the gs://bucket/obj url
func (g Path) String() string {
	return g.url.String()
}

// Set updates value from a gs://bucket/obj string, validating errors.
func (g *Path) Set(v string) error {
	u, err := url.Parse(v)
	if err != nil {
		return fmt.Errorf("invalid bucket url %s: %v", v, err)
	}
	return g.SetURL(u)
}

// SetURL updates value to the passed in gs://bucket/obj url
func (g *Path) SetURL(u *url.URL) error {
	var scheme = regexp.MustCompile(`gs|s3`)

	switch {
	case u == nil:
		return errors.New("nil url")
	case !scheme.MatchString(u.Scheme):
		return fmt.Errorf("must use a gs:// or s3:// url: %s", u)
	case strings.Contains(u.Host, ":"):
		return fmt.Errorf("bucket may not contain a port: %s", u)
	case u.Opaque != "":
		return fmt.Errorf("url must start with gs:// or s3://: %s", u)
	case u.User != nil:
		return fmt.Errorf("bucket may not contain an user@ prefix: %s", u)
	case u.RawQuery != "":
		return fmt.Errorf("bucket may not contain a ?query suffix: %s", u)
	case u.Fragment != "":
		return fmt.Errorf("bucket may not contain a #fragment suffix: %s", u)
	}
	g.url = *u
	return nil
}

// ResolveReference returns the path relative to the current path
func (g Path) ResolveReference(ref *url.URL) (*Path, error) {
	var newP Path
	if err := newP.SetURL(g.url.ResolveReference(ref)); err != nil {
		return nil, err
	}
	return &newP, nil
}

// Bucket returns bucket in gs://bucket/obj
func (g Path) Bucket() string {
	return g.url.Host
}

// Object returns path/to/something in gs://bucket/path/to/something
func (g Path) Object() string {
	if g.url.Path == "" {
		return g.url.Path
	}
	return g.url.Path[1:]
}

func calcCRC(buf []byte) uint32 {
	return crc32.Checksum(buf, crc32.MakeTable(crc32.Castagnoli))
}

// Upload writes bytes to the specified Path
func Upload(ctx context.Context, client *blob.Bucket, path Path, buf []byte) error {
	w, err := client.NewWriter(ctx, path.Object(), nil)
	if err != nil {
		return fmt.Errorf("Creating new writer failed: %v", err)
	}
	if n, err := w.Write(buf); err != nil {
		return fmt.Errorf("writing %s failed: %v", path, err)
	} else if n != len(buf) {
		return fmt.Errorf("partial write of %s: %d < %d", path, n, len(buf))
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing %s failed: %v", path, err)
	}
	return nil
}
