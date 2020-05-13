package filesystem

import (
	"fmt"
	"io"
	"os"
	"path"
)

type byteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	kb byteSize = 1 << (10 * iota)
	mb
	gb
	tb
	pb
	eb
	zb
	yb
)

var (
	permanentDir string
)

func (b byteSize) String() string {
	switch {
	case b >= yb:
		return fmt.Sprintf("%.2fYB", b/yb)
	case b >= zb:
		return fmt.Sprintf("%.2fZB", b/zb)
	case b >= eb:
		return fmt.Sprintf("%.2fEB", b/eb)
	case b >= pb:
		return fmt.Sprintf("%.2fPB", b/pb)
	case b >= tb:
		return fmt.Sprintf("%.2fTB", b/tb)
	case b >= gb:
		return fmt.Sprintf("%.2fGB", b/gb)
	case b >= mb:
		return fmt.Sprintf("%.2fMB", b/mb)
	case b >= kb:
		return fmt.Sprintf("%.2fKB", b/kb)
	}
	return fmt.Sprintf("%.2fB", b)
}

type fileInfo struct {
	os.FileInfo
}

func (fi *fileInfo) size() string {
	return byteSize(fi.FileInfo.Size()).String()
}

type fileInfoList []fileInfo

func (fil fileInfoList) Len() int {
	return len(fil)
}

func (fil fileInfoList) Swap(i, j int) {
	fil[i], fil[j] = fil[j], fil[i]
}

func (fil fileInfoList) Less(i, j int) bool {
	if permanentDir != "" {
		if fil[i].IsDir() && fil[i].Name() == permanentDir {
			return true
		}

		if fil[j].IsDir() && fil[j].Name() == permanentDir {
			return false
		}
	}

	return fil[i].ModTime().After(fil[j].ModTime())
}

func (fil fileInfoList) Exist(name string) bool {
	for _, fi := range fil {
		if fi.Name() == name {
			return true
		}
	}

	return false
}

// Storage represent file manipulation fasade
type Storage struct {
	rootPath string
}

// NewStorage create a new Storage object with abs root path
func NewStorage(root string) *Storage {
	return &Storage{
		rootPath: root,
	}
}

func (s *Storage) join(p string) string {
	return path.Join(s.rootPath, p)
}

// func (s *Storage) IsExists(p string) bool {
// 	_, err := os.Stat(s.Join(p))
// 	if err != nil {
// 		return !os.IsNotExist(err)
// 	}
// 	return true
// }

// func (s *Storage) mkdir(dir string) error {
// 	return os.Mkdir(s.join(dir), os.ModePerm)
// }

// Files returns files from current storage directory
func (s *Storage) Files(p string) fileInfoList {
	return newDirectory(s.join(p)).list()
}

// File return existing file with name fileName insice directory dir
func (s *Storage) File(dir, fileName string) (*File, error) {
	file, err := openFile(s.join(path.Join(dir, fileName)))
	if err != nil {
		err = fmt.Errorf("storage file: %w", err)
	}
	return file, err
}

// Store save file with fileName inside direcory dir, returns a closed newly created file object
func (s *Storage) Store(dir string, fileName string, data io.Reader) (*File, error) {
	f, err := newDirectory(s.join(dir)).createUniqueFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("[store file] unable to create file dir = %s, name = %s, err = %w", dir, fileName, err)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return nil, fmt.Errorf("[store file] error while copy file data dir = %s, name = %s, err = %w", dir, fileName, err)
	}

	return f, nil
}

// Move move file into direcotyr dir with name fileName
func (s *Storage) Move(file *File, dir string, fileName string) error {
	err := newDirectory(s.join(dir)).moveFile(file, fileName)
	if err != nil {
		return fmt.Errorf("[move file] unable to move file dir = %s, name = %s, err = %w", dir, fileName, err)
	}
	return nil
}

// Remove just remove file from storage
func (s *Storage) Remove(dir string, fileName string) error {
	f, err := openFile(s.join(path.Join(dir, fileName)))
	if err != nil {
		return fmt.Errorf("[remove file] unable to open file dir = %s, name = %s, err = %w", dir, fileName, err)
	}
	err = f.remove()
	if err != nil {
		return fmt.Errorf("[remove file] unable to remove file dir = %s, name = %s, err = %w", dir, fileName, err)
	}
	return nil
}
