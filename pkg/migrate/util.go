package migrate

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	compressedExtension       = "zip"
	stateExtenstion           = "rkestate"
	decompressedPathPrefix    = "snapshot-"
	controlPlaneRole          = "controlplane"
	etcdRole                  = "etcd"
	workerRole                = "worker"
	flannelPublicIPAnnotation = "flannel.alpha.coreos.com/public-ip"
	calicoIPAnnotation        = "projectcalico.org/IPv4Address"
)

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0700)

	for _, f := range r.File {
		err := extract(f, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func extract(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); err != nil {
			panic(err)
		}
	}()

	path := filepath.Join(dest, f.Name)

	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", path)
	}

	if f.FileInfo().IsDir() {
		os.MkdirAll(path, f.Mode())
	} else {
		os.MkdirAll(filepath.Dir(path), f.Mode())
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

func isCompressed(filename string) bool {
	return strings.HasSuffix(filename, fmt.Sprintf(".%s", compressedExtension))
}

func isStateFile(filename string) bool {
	return strings.HasSuffix(filename, fmt.Sprintf(".%s", stateExtenstion))
}

func findStateFile(destDir string) (string, error) {
	fileList, err := fileList(destDir)
	if err != nil {
		return "", err
	}
	for _, file := range fileList {
		if isStateFile(file) {
			return file, nil
		}
	}
	return "", fmt.Errorf("file not found")
}

func findSnapshotFile(destDir string) (string, error) {
	fileList, err := fileList(destDir)
	if err != nil {
		return "", err
	}
	for _, file := range fileList {
		f, err := os.Stat(file)
		if err != nil {
			return "", nil
		}
		if f.IsDir() {
			continue
		}
		if !isStateFile(file) {
			return file, nil
		}
	}
	return "", fmt.Errorf("snapshot file not found")
}

func fileList(destDir string) ([]string, error) {
	fileList := make([]string, 0)
	e := filepath.Walk(destDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		return nil, fmt.Errorf("failed to search content of snapshot")
	}
	return fileList, nil
}
