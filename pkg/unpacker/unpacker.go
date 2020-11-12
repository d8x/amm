package unpacker

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ModsUnpacker struct {
	currentPath         string
	rawModsDirName      string
	unpackedWorkDirName string
}

func NewModsUnpacker(rawModPath, unpackModDirectory string) (*ModsUnpacker, error) {
	currPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(unpackModDirectory) {
		unpackModDirectory = filepath.Join(currPath, unpackModDirectory)
	}
	if err := os.MkdirAll(unpackModDirectory, 0755); err != nil {
		return nil, err
	}
	return &ModsUnpacker{
		currentPath:         currPath,
		rawModsDirName:      rawModPath,
		unpackedWorkDirName: unpackModDirectory,
	}, nil
}

func (m *ModsUnpacker) Unpack() error {
	archivedFilesPathsSizes, err := m.getArchivedFilesPathsSizes(m.rawModsDirName)
	if err != nil {
		return err
	}
	for _, archiveFile := range archivedFilesPathsSizes {
		fileReader, err := m.getFileReader(archiveFile.AbsPath)
		if err != nil {
			return err
		}
		unpackedData, err := m.unpackArchive(fileReader)
		if err != nil {
			fmt.Printf("could not unpack archive: %s\n", archiveFile.AbsPath)
			return err
		}
		if archiveFile.Size != len(unpackedData) {
			fmt.Printf("size missmatch should: %d, is: %d\n", archiveFile.Size, len(unpackedData))
		}
		unpackFile := filepath.Join(m.unpackedWorkDirName, strings.TrimRight(archiveFile.RelPath, ".z"))

		err = m.writeFile(unpackedData, unpackFile)
		if err != nil {
			return err
		}
	}
	return nil
}

type archiveFile struct {
	AbsPath string
	RelPath string
	Size    int
}

func (m *ModsUnpacker) getArchivedFilesPathsSizes(dir string) ([]*archiveFile, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("provided path is not a directory")
	}

	// take the dir which is one above
	dir = path.Dir(dir)

	archiveRegexp, e := regexp.Compile("^.+\\.(z)$")
	if e != nil {
		return nil, e
	}
	var archivedFilesPaths []*archiveFile

	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if archiveRegexp.MatchString(f.Name()) {
			uncompressedSizeFilePath := fmt.Sprintf("%s.uncompressed_size", path)
			fu, err := os.Open(uncompressedSizeFilePath)
			if err != nil {
				fmt.Printf("cannot open file %s with uncompressedSize size\n", uncompressedSizeFilePath)
			} else {
				defer fu.Close()
			}
			var uncompressedSize int
			scanner := bufio.NewScanner(fu)
			for scanner.Scan() {
				lineStr := scanner.Text()
				uncompressedSize, err = strconv.Atoi(lineStr)
				if err != nil {
					fmt.Printf("cannot get file size")
				}
			}
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			archivedFilesPaths = append(archivedFilesPaths, &archiveFile{
				AbsPath: path,
				RelPath: relPath,
				Size:    uncompressedSize,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return archivedFilesPaths, nil
}

/*
 Read header information from archive:
            - 00 (8 bytes) signature (6 bytes) and format ver (2 bytes)
            - 08 (8 byes) unpacked/uncompressedSize chunk size
            - 10 (8 bytes) packed/compressedSize full size
            - 18 (8 bytes) unpacked/uncompressedSize full size
            - 20 (8 bytes) first chunk packed/compressedSize size
            - 26 (8 bytes) first chunk unpacked/uncompressedSize size
            - 20 and 26 repeat until the total of all the unpacked/uncompressedSize chunk sizes matches the unpacked/uncompressedSize full size.
Read all the archive data and verify integrity (there should only be one partial chunk, and each chunk should match the archives header).
https://github.com/barrycarey/Ark_Mod_Downloader/blob/master/arkit.py
*/

func (m *ModsUnpacker) getFileReader(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (m *ModsUnpacker) writeFile(data []byte, location string) error {
	if err := m.ensureDir(location); err != nil {
		return err
	}
	return ioutil.WriteFile(location, data, 0644)
}

func (m *ModsUnpacker) ensureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}
	return nil
}

func (m *ModsUnpacker) unpackArchive(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	header, err := m.unpackArchiveHeader(reader)
	if err != nil {
		return nil, err
	}
	fmt.Printf("header: %+v\n", header)
	chunksMetadata, err := m.unpackChunksMetadata(reader, header)
	if err != nil {
		return nil, err
	}
	data, err := m.unpackChunks(chunksMetadata, reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type archiveHeader struct {
	signature         int64 // signature (6 bytes) and format ver (2 bytes)
	unpackedChunkSize int64 // unpacked/uncompressedSize chunk size
	packedSize        int64 // packed/compressedSize full size
	unpackedSize      int64 // unpacked/uncompressedSize full size
}

func (m *ModsUnpacker) unpackArchiveHeader(reader io.ReadCloser) (*archiveHeader, error) {
	archiveHeader := new(archiveHeader)
	err := binary.Read(reader, binary.LittleEndian, &archiveHeader.signature)
	if err != nil {
		fmt.Printf("err signature: %v\n", err)
		return nil, err
	}
	// fmt.Println("signature: ", signature)
	err = binary.Read(reader, binary.LittleEndian, &archiveHeader.unpackedChunkSize)
	if err != nil {
		fmt.Printf("err sizeUnpackedChunk: %v\n", err)
		return nil, err
	}
	// fmt.Println("sizeUnpackedChunk: ", sizeUnpackedChunk)

	err = binary.Read(reader, binary.LittleEndian, &archiveHeader.packedSize)
	if err != nil {
		fmt.Printf("err sizePacked: %v\n", err)
		return nil, err
	}
	// fmt.Println("sizeUnpackedChunk: ", sizePacked)
	err = binary.Read(reader, binary.LittleEndian, &archiveHeader.unpackedSize)
	if err != nil {
		fmt.Printf("err unpackedSize: %v\n", err)
		return nil, err
	}
	return archiveHeader, nil
}

/*
Chunks metadata
    Need to read all chunks metadata. Iterating until we reach match of uncompressed size
*/

type chunksMetadata struct {
	compressedSize   int64 // chunk packed/compressedSize size
	uncompressedSize int64 // chunk unpacked/uncompressedSize size
}

func (m *ModsUnpacker) unpackChunksMetadata(reader io.ReadCloser, header *archiveHeader) ([]*chunksMetadata, error) {
	var chunks []*chunksMetadata
	var sizeIndex int64 = 0
	var rawCompressed, rawUncompressed [8]byte
	for sizeIndex < header.packedSize {
		_, err := reader.Read(rawCompressed[:])
		if err != nil {
			fmt.Printf("err rawCompressed: %v\n", err)
			return nil, err
		}
		_, err = reader.Read(rawUncompressed[:])
		if err != nil {
			fmt.Printf("err rawUnCompressed: %v\n", err)
			return nil, err
		}
		compressed := int64(binary.LittleEndian.Uint64(rawCompressed[:]))
		uncompressed := int64(binary.LittleEndian.Uint64(rawUncompressed[:]))
		chunks = append(chunks, &chunksMetadata{
			compressedSize:   compressed,
			uncompressedSize: uncompressed,
		})
		sizeIndex += compressed
	}
	if header.unpackedSize != sizeIndex {
		fmt.Printf("size mismatch, unpackedSize %d, sizeIndes: %d\n", header.unpackedSize, sizeIndex)
	}

	return chunks, nil
}

func (m *ModsUnpacker) unpackChunks(chunksMetadata []*chunksMetadata, reader io.ReadCloser) ([]byte, error) {
	var readChunks = 0
	var data []byte
	for _, chunk := range chunksMetadata {
		b := make([]byte, chunk.compressedSize)
		n, err := reader.Read(b)
		if err != nil {
			fmt.Printf("err %v read: %d uncompressedData, should: %d uncompressedData\n", err, n, chunk.compressedSize)
			return nil, err
		}
		buff := bytes.NewBuffer(b)
		z, err := zlib.NewReader(buff)
		if err != nil {
			fmt.Printf("err zlibReader: %v\n", err)
			return nil, err
		}
		uncompressedData, err := ioutil.ReadAll(z)
		if err != nil {
			fmt.Printf("err readALL: %v\n", err)
			return nil, err
		}
		z.Close()

		if len(uncompressedData) == int(chunk.uncompressedSize) {
			data = append(data, uncompressedData...)
			readChunks += 1
			// 	TODO more chunksMetadata size validation needed
		} else {
			fmt.Printf("error missmatch\n")
		}
	}
	return data, nil
}

/*
How To Parse modmeta.info:
            1. Read 4 bytes to tell how many key value pairs are in the file
            2. Read next 4 bytes tell us how many bytes to read ahead to get the key
            3. Read ahead by the number of bytes retrieved from step 2
            4. Read next 4 bytes to tell how many bytes to read ahead to get value
            5. Read ahead by the number of bytes retrieved from step 4
            6. Start at step 2 again
*/
func (m *ModsUnpacker) unMarshalModMetaInfo(fileData []byte) (map[string]string, error) {
	var totalPairs int32
	pairs := make(map[string]string)
	reader := bytes.NewReader(fileData)
	if err := binary.Read(reader, binary.LittleEndian, &totalPairs); err != nil {
		return nil, err
	}
	fmt.Printf("totalPairs: %d\n", totalPairs)
	for i := 0; i < int(totalPairs); i++ {
		var keySize int32
		var valueSize int32
		err := binary.Read(reader, binary.LittleEndian, &keySize)
		if err != nil {
			return nil, err
		}
		d := make([]byte, keySize)
		err = binary.Read(reader, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		err = binary.Read(reader, binary.LittleEndian, &valueSize)
		if err != nil {
			return nil, err
		}
		p := make([]byte, valueSize)

		err = binary.Read(reader, binary.LittleEndian, &p)
		if err != nil {
			return nil, err
		}
		pairs[string(d)] = string(p)
		fmt.Printf("keySize: %d\n", keySize)
		fmt.Printf("key: %s\n", string(d))
		fmt.Printf("valueSize: %d\n", valueSize)
		fmt.Printf("value: %s\n", string(p))

	}

	return pairs, nil
}

func (m *ModsUnpacker) unMarshalModInfo(fileData []byte) ([]string, error) {
	var totalPairs int32
	var pairs []string
	reader := bytes.NewReader(fileData)
	var s int32
	err := binary.Read(reader, binary.LittleEndian, &s)
	if err != nil {
		return nil, err
	}
	fmt.Println("s:", s)
	d := make([]byte, s)
	err = binary.Read(reader, binary.LittleEndian, &d)
	if err != nil {
		return nil, err
	}
	fmt.Println("d:", string(d))

	if err := binary.Read(reader, binary.LittleEndian, &totalPairs); err != nil {
		return nil, err
	}
	fmt.Printf("totalPairs: %d\n", totalPairs)
	for i := 0; i < int(totalPairs); i++ {
		var keySize int32
		err := binary.Read(reader, binary.LittleEndian, &keySize)
		if err != nil {
			return nil, err
		}
		d := make([]byte, keySize)
		err = binary.Read(reader, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, string(d))
	}
	return pairs, nil
}
