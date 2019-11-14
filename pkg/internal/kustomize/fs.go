/*
Copyright 2019 The KubeCarrier Authors.

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

package kustomize

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

// httpFSToKustomizeFS is reading the content of a http.FileSystem and writes it into a new kustomize fs.FileSystem.
func httpFSToKustomizeFS(httpFS http.FileSystem) (kFS fs.FileSystem, err error) {
	kFS = fs.MakeFsInMemory()
	if err = walk(httpFS, kFS, "/"); err != nil {
		return
	}
	return
}

// walk recurses the http.FileSystem and adds files and folders to the Kustomize file system.
func walk(httpFS http.FileSystem, kFS fs.FileSystem, filePath string) (err error) {
	file, err := httpFS.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return
	}

	// Directory handling
	if stat.IsDir() {
		if err = kFS.Mkdir(filePath); err != nil {
			return
		}

		var files []os.FileInfo
		if files, err = file.Readdir(-1); err != nil {
			return
		}
		for _, subFile := range files {
			if err = walk(httpFS, kFS, path.Join(filePath, subFile.Name())); err != nil {
				return
			}
		}
		return
	}

	// File handling
	newFile, err := kFS.Create(filePath)
	if err != nil {
		return
	}
	defer newFile.Close()
	// copy the file content into a buffer first,
	// because the fake file system of kustomize is kind of broken
	// and we would override our file with each written chunk
	var fileBuf bytes.Buffer
	if _, err = io.Copy(&fileBuf, file); err != nil {
		return
	}

	if _, err = newFile.Write(fileBuf.Bytes()); err != nil {
		return
	}
	return
}
