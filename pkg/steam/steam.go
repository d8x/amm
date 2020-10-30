package steam

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	tmpSteamLocation   = "./steam/"
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
}

func NewSteamHandler() (*SteamHandler, error) {
	// TODO: check path for binary, check PATH for system availability
	if err := os.MkdirAll(tmpSteamLocation, 0755); err != nil {
		return nil, err
	}
	return &SteamHandler{}, nil
}

func (s *SteamHandler) DownloadMod(modID string) error {
	if err := s.setSteamCMDPath(); err != nil {
		return err
	}
	c := exec.Command(s.CMDLocation, "+login", "anonymous", "+workshop_download_item", arkGameID, modID,
		"+force_install_dir", tmpSteamLocation, "+quit")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
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
	fmt.Println("OS: ", runtime.GOOS)
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
