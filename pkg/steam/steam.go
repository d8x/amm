package steam

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/otiai10/copy"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	tmpSteamLocation   = "./steam"
	windowsSteamCMDURL = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip"
	windowsZipFileName = "steamcmd.zip"
	linuxSteamCMDURL   = "http://media.steampowered.com/installer/steamcmd_linux.tar.gz"
	linuxTarGzFileName = "steamcmd.tar.gz"
	steamCMD           = "steamcmd"
	arkGameID          = "346110"
)

// var ErrSteamCLINotAvailable = errors.New("steam cli not available")

type SteamHandler struct {
	CMDLocation string
	workDir     string
}

func NewSteamHandler(workDir string) (*SteamHandler, error) {
	currPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(currPath, workDir), 0755); err != nil {
		return nil, err
	}
	return &SteamHandler{
		workDir: workDir,
	}, nil
}

func (s *SteamHandler) DownloadMod(modID string) (string, error) {
	if err := s.setSteamCMDPath(); err != nil {
		return "", err
	}
	tmpDir := filepath.Join(os.TempDir(), "amm")
	c := exec.Command(s.CMDLocation, "+login", "anonymous", "+force_install_dir", tmpDir, "+workshop_download_item", arkGameID, modID, "+quit")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("could not cleanup tmp directory %v\n", err)
		}

	}()
	if err := os.MkdirAll(s.workDir, 0755); err != nil {
		return "", err
	}
	finalModLocation, err := s.copyMod(tmpDir, modID)
	if err != nil {
		return "", err
	}
	return finalModLocation, nil
}

func (s *SteamHandler) copyMod(srcLocation, modID string) (string, error) {
	srcModLocation := filepath.Join(srcLocation, "/steamapps/workshop/content", arkGameID, modID)
	dstLocation := filepath.Join(s.workDir, modID)
	_, err := os.Stat(dstLocation)
	if os.IsExist(err) {
		return "", errors.New("destination mod location already exists")
	}
	if err := copy.Copy(srcModLocation, dstLocation); err != nil {
		return "", err
	}
	return dstLocation, nil
}

func (s *SteamHandler) setSteamCMDPath() error {
	absolutePath, err := exec.LookPath(steamCMD)
	if err != nil {
		return err
	}
	s.CMDLocation = absolutePath
	return nil
}

func (s *SteamHandler) DownloadCMD() error {
	switch runtime.GOOS {
	case "windows":
		resp, err := http.Get(windowsSteamCMDURL)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := ioutil.WriteFile(filepath.Join(tmpSteamLocation, windowsZipFileName), data, 0755); err != nil {
			return err
		}
		return s.unpackWindows(filepath.Join(tmpSteamLocation, windowsZipFileName))
	case "linux":
		resp, err := http.Get(linuxSteamCMDURL)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := ioutil.WriteFile(filepath.Join(tmpSteamLocation, linuxTarGzFileName), data, 0755); err != nil {
			return err
		}
		return s.unpackLinux(filepath.Join(tmpSteamLocation, linuxTarGzFileName))
	default:
		fmt.Println("not supported")
	}

	return nil
}

func (s *SteamHandler) unpackWindows(zipLocation string) error {
	z, err := zip.OpenReader(zipLocation)
	if err != nil {
		return err
	}
	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filepath.Join(tmpSteamLocation, f.Name), f.Mode()); err != nil {
				return err
			}
		} else {
			outputFile, err := os.OpenFile(
				filepath.Join(tmpSteamLocation, f.Name),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				f.Mode(),
			)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SteamHandler) unpackLinux(tarGzLocation string) error {
	reader, err := os.Open(tarGzLocation)
	if err != nil {
		return err
	}
	defer reader.Close()
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		p := filepath.Join(tmpSteamLocation, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(p, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
