package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

const filenameTemplate = "%027d"

type Manifest struct {
	ManifestPath  string
	ManifestFile  *os.File
	NextSSTableId int64
	SSTFiles      []string
}

const manifestPath = "manifest"

func NewManifest(storagePath string) *Manifest {
	manifestPath := filepath.Join(storagePath, manifestPath)
	return &Manifest{ManifestPath: manifestPath}
}
func (m *Manifest) InitManifest() error {
	_, err := os.Stat(m.ManifestPath)
	if err == nil {
		m.ManifestFile, err = os.OpenFile(m.ManifestPath, os.O_RDWR|os.O_APPEND, 0644)
		return m.findLastSSTFile(m.ManifestFile)
	}

	if errors.Is(err, os.ErrNotExist) {
		m.ManifestFile, err = os.OpenFile(m.ManifestPath, os.O_CREATE|os.O_RDWR, 0644)
		m.NextSSTableId = 1
		if err != nil {
			return err
		}
	}
	return err
}

func (m *Manifest) Append(fileRecord string) error {
	_, err := m.ManifestFile.WriteString(fmt.Sprintf("%s\n", fileRecord))
	return err
}

func (m *Manifest) Close() error {
	return m.ManifestFile.Close()
}

func (m *Manifest) findLastSSTFile(file *os.File) error {
	buf := make([]byte, 1)
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()
	strBuf := bytes.Repeat([]byte{'0'}, 27)
	pos := 26
	for i := size - 1; i >= 0; i-- {
		_, err = file.ReadAt(buf, i)
		if err != nil && err != io.EOF {
			return err
		}
		if buf[0] == '\n' && i == size-1 {
			continue
		}

		if buf[0] == '\n' && i != size-1 {
			break
		}
		strBuf[pos] = buf[0]
		pos--
	}
	line := string(strBuf)
	if line == "" {
		m.NextSSTableId = 1
		return nil
	}
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return err
	}
	m.NextSSTableId = n + 1
	return nil
}

func (m *Manifest) readFileListFromManifest(file *os.File) error {
	buf := make([]byte, 4096)
	var partial []byte
	var sstFileList []string
	var err error
	var n int
	for {
		n, err = file.Read(buf)
		if n == 0 {
			break
		}

		data := append(partial, buf[:n]...)

		lines := bytes.Split(data, []byte("\n"))

		partial = lines[len(lines)-1]

		for _, line := range lines[:len(lines)-1] {
			sstFileList = append(sstFileList, string(line))
		}

		if err != nil {
			return err
		}
	}
	sort.Strings(sstFileList)
	m.SSTFiles = sstFileList
	m.NextSSTableId, err = strconv.ParseInt(m.SSTFiles[len(m.SSTFiles)-1], 10, 64)
	if err != nil {
		return err
	}
	return nil
}
