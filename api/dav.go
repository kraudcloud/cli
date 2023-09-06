package api

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	webdav "github.com/emersion/go-webdav"
)

const endpoint = "https://files.kraudcloud.com"

func (c *Client) UploadDir(ctx context.Context, namespace string, localPath, remotePath string) error {
	webc, err := webdav.NewClient(webdav.HTTPClientWithBasicAuth(nil, "", c.authToken), endpoint)
	if err != nil {
		return fmt.Errorf("error connecting to remote webdav: %w", err)
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	remotePathPrefix := path.Join(namespace, "volumes", remotePath)
	if info.IsDir() {
		return copyDirDav(os.DirFS(localPath), webc, remotePathPrefix, localPath)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return copyFileDav(webc, f, path.Join(remotePathPrefix, localPath))
}

func copyDirDav(lfs fs.FS, webc *webdav.Client, davPrefix, localPath string) error {
	// make initial directory
	// ignore any error, split on / and just make each directory
	// individually
	davPrefix = filepath.ToSlash(davPrefix)
	parts := strings.Split(davPrefix, "/")
	for i := range parts {
		webc.Mkdir(path.Join(parts[:i]...))
	}

	return fs.WalkDir(lfs, ".", func(localPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error while walking files: %w", err)
		}

		if d.IsDir() {
			_, err := webc.Stat(davPrefix)
			if err != nil {
				err = webc.Mkdir(davPrefix)
				if err != nil {
					return err
				}

				return nil
			}

			return nil
		}

		file, err := lfs.Open(localPath)
		if err != nil {
			return err
		}

		return copyFileDav(webc, file, path.Join(davPrefix, localPath))
	})
}

func copyFileDav(webc *webdav.Client, file fs.File, finalPath string) error {
	rw, err := webc.Create(finalPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(rw, file)
	return err
}
